package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

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
	r.HandleFunc("/api/combine", combineHandler).Methods("POST")

	// Start the HTTP server
	log.Println("Server is listening on port 8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func combineHandler(w http.ResponseWriter, r *http.Request) {

}

func testFiles() {
	// Open the first file
	file1, err := os.Open("test_files/test1.pdf")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	doc1, err := fileToDocument(file1)
	if err != nil {
		log.Fatal(err.Error())
	}
	file1.Close()

	// Open the second file
	file2, err := os.Open("test_files/test2.pdf")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	doc2, err := fileToDocument(file2)
	if err != nil {
		log.Fatal(err.Error())
	}
	file2.Close()

	// Open the docx file and convert it to a pdf
	docx, err := os.Open("test_files/doc.docx")
	if err != nil {
		log.Fatal(err.Error())
	}
	doc3, err := fileToDocument(docx)
	if err != nil {
		log.Fatal(err.Error())
	}
	docx.Close()

	// Open the pptx file and convert it to a pdf
	pptx, err := os.Open("test_files/slideshow.pptx")
	if err != nil {
		log.Fatal(err.Error())
	}
	doc4, err := fileToDocument(pptx)
	if err != nil {
		log.Fatal(err.Error())
	}
	pptx.Close()

	documents := []*document{doc1, doc2, doc3, doc4}
	combined, err := combineDocuments(documents)
	if err != nil {
		log.Fatal(err.Error())
	}
	combinedPdf, err := os.Create("combined.pdf")
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = combinedPdf.Write(combined)
	if err != nil {
		log.Fatal(err.Error())
	}
}
