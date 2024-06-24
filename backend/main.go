package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

const (
	kilobyte = 1024
	megabyte = kilobyte * 1024
)

var supportedFileTypes = []string{"pdf", "doc", "docx", "ppt", "pptx"}

func main() {
	// Test connection to Gotenberg
	healthRes, err := http.Get("http://localhost:3000/health")
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
			// fmt.Println(fileType.Extension)
			// fmt.Println("fileType == types.Unknown:", fileType == types.Unknown)
			// fmt.Println("slices.Contains(supportedFileTypes, fileType.Extension):", slices.Contains(supportedFileTypes, fileType.Extension))
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

// func testFiles() {
// 	// Open the first file
// 	file1, err := os.Open("test_files/test1.pdf")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 		return
// 	}
// 	doc1, err := fileToDocument(file1)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	file1.Close()

// 	// Open the second file
// 	file2, err := os.Open("test_files/test2.pdf")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 		return
// 	}
// 	doc2, err := fileToDocument(file2)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	file2.Close()

// 	// Open the docx file and convert it to a pdf
// 	docx, err := os.Open("test_files/doc.docx")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	doc3, err := fileToDocument(docx)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	docx.Close()

// 	// Open the pptx file and convert it to a pdf
// 	pptx, err := os.Open("test_files/slideshow.pptx")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	doc4, err := fileToDocument(pptx)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	pptx.Close()

// 	documents := []*document{doc1, doc2, doc3, doc4}
// 	combined, err := combineDocuments(documents)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	combinedPdf, err := os.Create("combined.pdf")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	if _, err := combinedPdf.Write(combined); err != nil {
// 		log.Fatal(err.Error())
// 	}
// }
