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
	// Check if the script was provided with a file argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./printer-monitor <file_path>")
		return
	}

	filePath := os.Args[1] // Get the file path from the command-line argument

	// Define a list of Mainsail API endpoints for printer status
	statusURLs := []string{
		//"http://printer1-api.example.com/api/printer/info",
		//"http://printer2-api.example.com/api/printer/info",
		"https://ender.local.antnsn.dev/printer/info", // Add the new status URL here
	}

	// Create a map to track printers that have received a file
	printerFileSent := make(map[string]bool)

	// Define the polling interval (e.g., every 30 seconds)
	pollingInterval := 30 * time.Second

	// Infinite loop for monitoring
	for {
		// Iterate through the list of printer status URLs
		for _, statusURL := range statusURLs {
			// Check if the printer has already received a file
			if printerFileSent[statusURL] {
				continue
			}

			// Send a GET request to the API endpoint to get printer status
			response, err := http.Get(statusURL)
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
				processReadyPrinter(infoResponse.Result, statusURL, filePath)
				printerFileSent[statusURL] = true // Mark the printer as having received a file
			}
		}

		// Sleep for the polling interval before checking again
		time.Sleep(pollingInterval)
	}
}

// processReadyPrinter processes the action for a printer with state "ready"
func processReadyPrinter(printer PrinterStatus, statusURL string, filePath string) {
	fmt.Printf("Printer with hostname '%s' is 'ready'. Taking action...\n", printer.Hostname)
	// Add your code to send a file to OctoPrint or perform any other action for this printer

	// Example: Send the file specified in the command-line argument to OctoPrint
	sendFileToOctoPrint(printer, filePath)
}

// sendFileToOctoPrint sends a file to OctoPrint
func sendFileToOctoPrint(printer PrinterStatus, filePath string) {
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

	// Create a POST request to send the file to OctoPrint
	octoPrintURL := fmt.Sprintf("http://%s/api/files/local", printer.Hostname) // Change the URL as needed
	req, err := http.NewRequest("POST", octoPrintURL, body)
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		return
	}

	// Set the API key as a header (replace 'YOUR_API_KEY' with your OctoPrint API key)
	req.Header.Set("X-Api-Key", "YOUR_API_KEY")

	// Set the Content-Type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending file to OctoPrint:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("File sent to OctoPrint on printer '%s'\n", printer.Hostname)
}
