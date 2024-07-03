package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

const (
	kilobyte = 1024
	megabyte = kilobyte * 1024

	maxFiles = 1000
	maxSize  = 50 * megabyte
)

var supportedFileTypes = []string{"pdf", "doc", "docx", "ppt", "pptx"}
var gotenbergClient = &http.Client{Timeout: 10 * time.Second}

// Environment variables
var gotenbergUrl string
var port string

type document struct {
	name string
	data []byte
}

func main() {
	// Set environment variables
	if env := os.Getenv("GOTENBERG_URL"); env == "" {
		gotenbergUrl = "http://localhost:3000"
		log.Println("GOTENBERG_URL not specified, setting to " + gotenbergUrl)
	} else {
		gotenbergUrl = env
	}
	if env := os.Getenv("PORT"); env == "" {
		port = "3000"
		log.Println("PORT not specified, setting to " + port)
	} else {
		port = env
	}

	// Test connection to Gotenberg
	healthUrl := gotenbergUrl + "/health"
	healthRes, err := gotenbergClient.Get(healthUrl)
	if err != nil {
		log.Fatal(err.Error())
	}
	if healthRes.StatusCode != 200 {
		log.Fatal("Gotenberg health check did not return 200")
	}

	// Create router, API route, and static file server
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.HandleFunc("/combine", combineHandler).Methods("POST")

	// Start the HTTP server
	log.Println("Server is listening on port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err.Error())
	}
}

func logRequest(request *http.Request, message string) {
	log.Printf("[%s] (%s %s) %s\n", request.RemoteAddr, request.Method, request.RequestURI, message)
}

func notFoundHandler(resWriter http.ResponseWriter, request *http.Request) {
	http.Error(resWriter, "404 page not found", http.StatusNotFound)
	logRequest(request, "error: 404 page not found")
}

func combineHandler(resWriter http.ResponseWriter, request *http.Request) {
	// Set CORS
	resWriter.Header().Set("Access-Control-Allow-Origin", "*")

	// Limit the size of the request body
	request.Body = http.MaxBytesReader(resWriter, request.Body, maxSize)

	// Parse the multipart form
	if err := request.ParseMultipartForm(maxSize); err != nil { // TODO: figure out best max size for this
		http.Error(resWriter, "Failed to parse multipart form", http.StatusBadRequest)
		logRequest(request, "error: "+err.Error())
		return
	}

	// Parse files
	formData := request.MultipartForm
	files := formData.File["documents"]
	if len(files) > maxFiles {
		http.Error(resWriter, "Too many files", http.StatusRequestEntityTooLarge)
		logRequest(request, "error: Too many files")
		return
	}
	var documents []*document
	for i, fileHeader := range files {
		// Open file and put into document struct
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(resWriter, "Error opening file", http.StatusInternalServerError)
			logRequest(request, "error: "+err.Error())
			return
		}
		docData, err := io.ReadAll(file)
		if err != nil {
			http.Error(resWriter, "Error reading file", http.StatusInternalServerError)
			logRequest(request, "error: "+err.Error())
			return
		}
		document := document{name: fileHeader.Filename, data: docData}
		file.Close()

		// Check file type
		fileType, err := filetype.Match(document.data)
		if err != nil {
			http.Error(resWriter, "Error matching file type", http.StatusInternalServerError)
			logRequest(request, "error: "+err.Error())
			return
		}
		if fileType == types.Unknown || !slices.Contains(supportedFileTypes, fileType.Extension) {
			http.Error(resWriter, "Unsupported file type", http.StatusUnsupportedMediaType)
			logRequest(request, "error: Unsupported file type")
			return
		}
		if fileType != types.Get("pdf") {
			if err := document.convertToPdf(); err != nil {
				http.Error(resWriter, "Error converting to PDF", http.StatusInternalServerError)
				logRequest(request, "error: "+err.Error())
				return
			}
		}

		// Rename the file so Gutenberg combines in the right order
		document.name = fmt.Sprintf("%03d.pdf", i)

		documents = append(documents, &document)
	}

	// Combine documents and send new file as response
	combined, err := combineDocuments(documents)
	if err != nil {
		http.Error(resWriter, "Error combining documents", http.StatusInternalServerError)
		logRequest(request, "error: "+err.Error())
		return
	}
	resWriter.Header().Set("Content-Disposition", "attachment; filename=combined.pdf")
	resWriter.Header().Set("Content-Type", "application/pdf")
	if _, err := io.Copy(resWriter, bytes.NewReader(combined)); err != nil {
		http.Error(resWriter, "Error forming response", http.StatusInternalServerError) // TODO: check if this overwrites the data currently written
		logRequest(request, "error: "+err.Error())
		return
	}
	logRequest(request, "success")
}

func combineDocuments(documents []*document) ([]byte, error) {
	url := gotenbergUrl + "/forms/pdfengines/merge"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files to form
	for i, doc := range documents {
		part, err := writer.CreateFormFile(fmt.Sprintf("file%d", i), doc.name)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, bytes.NewReader(doc.data)); err != nil {
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
	url := gotenbergUrl + "/forms/libreoffice/convert"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file to form
	part, err := writer.CreateFormFile("file", document.name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, bytes.NewReader(document.data)); err != nil {
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
		document.name += ".pdf"
		document.data = respBody
		return nil
	} else {
		return errors.New(string(respBody))
	}
}
