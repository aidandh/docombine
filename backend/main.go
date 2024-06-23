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
	"path/filepath"
)

type document struct {
	Name string
	Data []byte
}

func main() {
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
	err = doc3.convertToPdf()
	if err != nil {
		log.Fatal(err.Error())
	}

	documents := []*document{doc1, doc2, doc3}
	combined, err := combinePdfs(documents)
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

func fileToDocument(file *os.File) (*document, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(file.Name())
	return &document{Name: name, Data: data}, nil
}

func combinePdfs(documents []*document) ([]byte, error) {
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
		_, err = io.Copy(part, bytes.NewReader(doc.Data))
		if err != nil {
			return nil, err
		}
	}
	writer.Close()

	// Send the request to Gotenberg
	response, err := http.Post(url, writer.FormDataContentType(), &requestBody)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Return the response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.Status == "200 OK" {
		return respBody, nil
	} else {
		return nil, errors.New(response.Status)
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
	_, err = io.Copy(part, bytes.NewReader(document.Data))
	if err != nil {
		return err
	}
	writer.Close()

	// Send the request to Gotenberg
	response, err := http.Post(url, writer.FormDataContentType(), &requestBody)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Return the response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.Status == "200 OK" {
		document.Name += ".pdf"
		document.Data = respBody
		return nil
	} else {
		return errors.New(response.Status)
	}
}
