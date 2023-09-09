package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const pollingInterval = 10 * time.Second // Set the polling interval here (e.g., 10 seconds)
const uploadEndpoint = "/upload"         // Define the endpoint for file uploads
const uploadFolder = "uploads"           // Define the folder for uploaded files

type PrinterInfoResponse struct {
	Result PrinterStatus `json:"result"`
}

type PrinterStatus struct {
	State           string `json:"state"`
	StateMessage    string `json:"state_message"`
	Hostname        string `json:"hostname"`
	KlipperPath     string `json:"klipper_path"`
	PythonPath      string `json:"python_path"`
	LogFile         string `json:"log_file"`
	ConfigFile      string `json:"config_file"`
	SoftwareVersion string `json:"software_version"`
	CPUInfo         string `json:"cpu_info"`
}

func main() {
	// Check if the "uploads" folder exists, and create it if not
	if _, err := os.Stat(uploadFolder); os.IsNotExist(err) {
		err := os.Mkdir(uploadFolder, 0755) // Create the folder with read/write permissions
		if err != nil {
			fmt.Printf("Error creating folder '%s': %s\n", uploadFolder, err)
			return
		}
	}

	// Print the monitoring message with the polling interval
	fmt.Printf("Monitoring printers with a polling interval of %s...\n", pollingInterval)

	// Define a list of printer URLs (without /printer/info)
	printerURLs := []string{
		"https://ender.local.antnsn.dev", // Add your printer URLs here
	}

	// Modify the printerURLs slice to include /printer/info for each printer
	for i, url := range printerURLs {
		printerURLs[i] = fmt.Sprintf("%s/printer/info", url)
	}

	// Infinite loop for continuous monitoring
	for {
		// Iterate through the list of printer URLs
		for _, url := range printerURLs {
			// Send a GET request to the API endpoint to get printer status
			response, err := http.Get(url)
			if err != nil {
				fmt.Println("Error:", err)
				continue // Continue to the next printer URL on error
			}
			defer response.Body.Close()

			// Check if the response status code is 200 (OK)
			if response.StatusCode != http.StatusOK {
				fmt.Println("Error: Unexpected status code", response.StatusCode)
				continue // Continue to the next printer URL on error
			}

			// Decode the JSON response
			var infoResponse PrinterInfoResponse
			if err := json.NewDecoder(response.Body).Decode(&infoResponse); err != nil {
				fmt.Println("Error:", err)
				continue // Continue to the next printer URL on error
			}

			// Check if the printer is in the "ready" state
			if infoResponse.Result.State == "ready" {
				// Check for new uploads in the "uploads" folder
				filePath := checkForNewUploads()
				if filePath != "" {
					processReadyPrinter(infoResponse.Result, url, filePath)
				}
			}
		}

		// Sleep for the specified polling interval
		time.Sleep(pollingInterval)
	}
}

// checkForNewUploads checks if there are new uploads in the "uploads" folder
func checkForNewUploads() string {
	files, err := os.ReadDir(uploadFolder)
	if err != nil {
		fmt.Println("Error reading upload folder:", err)
		return ""
	}

	for _, file := range files {
		if !file.IsDir() {
			// Get the path of the uploaded file
			filePath := fmt.Sprintf("%s/%s", uploadFolder, file.Name())
			return filePath
		}
	}

	return ""
}

// processReadyPrinter processes the action for a printer with state "ready"
func processReadyPrinter(printer PrinterStatus, printerURL string, filePath string) {
	// Implement your action here, e.g., send a file to OctoPrint with 'printer.Hostname'
	fmt.Printf("Printer with hostname '%s' is 'ready'. Taking action...\n", printer.Hostname)
	// Add your code to send a file to OctoPrint or perform any other action for this printer

	// Example: Send the file to the printer's upload endpoint
	sendFileToPrinter(printer, printerURL, filePath)
}

// sendFileToPrinter sends a file to the printer's upload endpoint
func sendFileToPrinter(printer PrinterStatus, printerURL string, filePath string) {
	// Create a POST request to send the file to the printer's upload endpoint
	uploadURL := fmt.Sprintf("%s/files/upload", printerURL)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create a buffer for the file content
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form field for the file
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return
	}

	// Copy the file content to the form field
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file content:", err)
		return
	}

	// Close the multipart writer
	writer.Close()

	// Create a POST request to send the file to the printer's upload endpoint
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		return
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending file to printer:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("File sent to printer '%s'\n", printer.Hostname)
}
