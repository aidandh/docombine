package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type document struct {
	Name string
	Data []byte
}

// func fileToDocument(file *os.File) (*document, error) {
// 	data, err := io.ReadAll(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	name := filepath.Base(file.Name())
// 	return &document{Name: name, Data: data}, nil
// }

func combineDocuments(documents []*document) ([]byte, error) {
	url := "http://localhost:3000/forms/pdfengines/merge"

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add files to form
	for i, doc := range documents {
		// if !filetype.IsType(doc.Data, types.Get("pdf")) {
		// 	err := doc.convertToPdf()
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
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
	if response.StatusCode == 200 {
		document.Name += ".pdf"
		document.Data = respBody
		return nil
	} else {
		return errors.New(string(respBody))
	}
}
