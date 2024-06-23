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

func main() {
	// Open the first file
	file1, err := os.Open("test_files/file1.pdf")
	if err != nil {
		fmt.Println("Error opening file1:", err)
		return
	}
	defer file1.Close()

	// Open the second file
	file2, err := os.Open("test_files/file2.pdf")
	if err != nil {
		fmt.Println("Error opening file2:", err)
		return
	}
	defer file2.Close()

	files := []*os.File{file1, file2}
	combined, err := combinePdfs(files)
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

func combinePdfs(files []*os.File) ([]byte, error) {
	url := "http://localhost:3000/forms/pdfengines/merge"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files to form
	for i, file := range files {
		part, err := writer.CreateFormFile(fmt.Sprintf("file%d", i), filepath.Base(file.Name()))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}
	writer.Close()

	// Use http.Post to send the request
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
