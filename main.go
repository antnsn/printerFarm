package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Define your 3D printer URLs here.
var printerURLs = []string{
	"https://ender.local.antnsn.dev", // Add the printer URLs here

}

func main() {
	// Start an infinite loop to accept files and send them to printers.
	for {
		// Simulate accepting a file (you can implement your own logic to handle file uploads).
		fileToUpload := createSampleFile()

		// Find a printer that is ready.
		readyPrinterURL := findReadyPrinter()

		if readyPrinterURL != "" {
			// Upload the file to the ready printer.
			err := uploadFileToPrinter(fileToUpload, readyPrinterURL)
			if err != nil {
				fmt.Printf("Error uploading file to printer: %v\n", err)
			} else {
				fmt.Printf("File successfully uploaded to %s\n", readyPrinterURL)
			}
		} else {
			fmt.Println("No ready printers found. Waiting for a printer to become ready...")
		}

		// Sleep for a while before checking again (you can adjust the duration).
		time.Sleep(10 * time.Second)
	}
}

// Simulate creating a sample file (you can replace this with your file handling logic).
func createSampleFile() io.Reader {
	// Create a sample file as bytes.
	fileContent := []byte("This is a sample file content.")
	return bytes.NewReader(fileContent)
}

// Find a printer that is ready.
func findReadyPrinter() string {
	for _, url := range printerURLs {
		// Check the printer's status URL.
		statusURL := url + "/printer/info"
		resp, err := http.Get(statusURL)
		if err != nil {
			fmt.Printf("Error checking printer status at %s: %v\n", statusURL, err)
			continue
		}
		defer resp.Body.Close()

		// Check if the printer is ready (you can adjust this based on your printer's response format).
		if resp.StatusCode == http.StatusOK {
			return url
		}
	}
	return ""
}

// Upload a file to the specified printer.
func uploadFileToPrinter(file io.Reader, printerURL string) error {
	// Create a multipart request with the file.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file to the request.
	part, err := writer.CreateFormFile("file", "sample-file.gcode")
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	// Close the multipart writer.
	writer.Close()

	// Make a POST request to the printer's upload endpoint.
	uploadURL := printerURL + "/files"
	resp, err := http.Post(printerURL, writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("file upload failed with status code: %d", resp.StatusCode)
	}

	return nil
}
