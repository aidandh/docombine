package main

import (
	"log"
	"os"
)

func main() {
	testFiles()
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
