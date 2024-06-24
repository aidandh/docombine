package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

const (
	kilobyte = 1024
	megabyte = kilobyte * 1024
)

var supportedFileTypes = []string{"pdf", "doc", "docx", "ppt", "pptx"}
var gotenbergClient = &http.Client{Timeout: 10 * time.Second}

type document struct {
	Name string
	Data []byte
}

func main() {
	// Test connection to Gotenberg
	healthRes, err := gotenbergClient.Get("http://localhost:3000/health")
	if err != nil {
		log.Fatal(err.Error())
	}
	if healthRes.StatusCode != 200 {
		log.Fatal("Gotenberg health check did not return 200")
	}

	// Create router and API routes
	r := mux.NewRouter()
	r.HandleFunc("/combine", combineHandler).Methods("POST")

	// Start the HTTP server
	log.Println("Server is listening on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err.Error())
	}
}

func combineHandler(resWriter http.ResponseWriter, request *http.Request) {
	// Parse the multipart form
	if err := request.ParseMultipartForm(20 * megabyte); err != nil { // TODO: figure out best max size for this
		http.Error(resWriter, "Failed to parse multipart form", http.StatusBadRequest)
		log.Println(request.RemoteAddr, "error:", err.Error())
		return
	}

	// Parse files
	formData := request.MultipartForm
	files := formData.File["documents"]
	var documents []*document
	for i, fileHeader := range files {
		// Open file and put into document struct
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(resWriter, "Error opening file", http.StatusInternalServerError)
			log.Println(request.RemoteAddr, "error:", err.Error())
			return
		}
		docData, err := io.ReadAll(file)
		if err != nil {
			http.Error(resWriter, "Error reading file", http.StatusInternalServerError)
			log.Println(request.RemoteAddr, "error:", err.Error())
			return
		}
		document := document{Name: fileHeader.Filename, Data: docData}
		file.Close()

		// Check file type
		fileType, err := filetype.Match(document.Data)
		if err != nil {
			http.Error(resWriter, "Error matching file type", http.StatusInternalServerError)
			log.Println(request.RemoteAddr, "error:", err.Error())
			return
		}
		if fileType == types.Unknown || !slices.Contains(supportedFileTypes, fileType.Extension) {
			http.Error(resWriter, "Unsupported file type", http.StatusInternalServerError)
			log.Println(request.RemoteAddr, "error: Unsupported file type")
			return
		}
		if fileType != types.Get("pdf") {
			if err := document.convertToPdf(); err != nil {
				http.Error(resWriter, "Error converting to PDF", http.StatusInternalServerError)
				log.Println(request.RemoteAddr, "error:", err.Error())
				return
			}
		}

		// Rename the file so Gutenberg combines in the right order
		document.Name = fmt.Sprintf("%03d.pdf", i)

		documents = append(documents, &document)
	}

	// Combine documents and send new file as response
	combined, err := combineDocuments(documents)
	if err != nil {
		http.Error(resWriter, "Error combining documents", http.StatusInternalServerError)
		log.Println(request.RemoteAddr, "error:", err.Error())
		return
	}
	resWriter.Header().Set("Content-Disposition", "attachment; filename=file.pdf")
	resWriter.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(resWriter, bytes.NewReader(combined)); err != nil {
		http.Error(resWriter, "Error forming response", http.StatusInternalServerError) // TODO: check if this overwrites the data currently written
		log.Println(request.RemoteAddr, "error:", err.Error())
		return
	}
	log.Println(request.RemoteAddr, "success, sending combined file")
}

func combineDocuments(documents []*document) ([]byte, error) {
	url := "http://localhost:3000/forms/pdfengines/merge"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files to form
	for i, doc := range documents {
		part, err := writer.CreateFormFile(fmt.Sprintf("file%d", i), doc.Name)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, bytes.NewReader(doc.Data)); err != nil {
			return nil, err
		}
	}
	writer.Close()

	// Send the request to Gotenberg
	response, err := gotenbergClient.Post(url, writer.FormDataContentType(), &requestBody)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Return the response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == 200 {
		return respBody, nil
	} else {
		return nil, errors.New(string(respBody))
	}
}

func (document *document) convertToPdf() error {
	url := "http://localhost:3000/forms/libreoffice/convert"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file to form
	part, err := writer.CreateFormFile("file", document.Name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, bytes.NewReader(document.Data)); err != nil {
		return err
	}
	writer.Close()

	// Send the request to Gotenberg
	response, err := gotenbergClient.Post(url, writer.FormDataContentType(), &requestBody)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Return the response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		document.Name += ".pdf"
		document.Data = respBody
		return nil
	} else {
		return errors.New(string(respBody))
	}
}
