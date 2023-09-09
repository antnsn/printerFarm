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

const pollingInterval = 10 * time.Second
const uploadEndpoint = "/upload"
const uploadFolder = "uploads"

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

// Define a channel to signal when a new file has been uploaded
var newUploadChan = make(chan struct{})

func main() {
	// Check if the "uploads" folder exists, and create it if not
	if _, err := os.Stat(uploadFolder); os.IsNotExist(err) {
		err := os.Mkdir(uploadFolder, 0755) // Create the folder with read/write permissions
		if err != nil {
			fmt.Printf("Error creating folder '%s': %s\n", uploadFolder, err)
			return
		}
	}

	// Start the file upload server
	go startUploadServer()

	// Print the monitoring message with the polling interval
	fmt.Printf("Monitoring printers with a polling interval of %s...\n", pollingInterval)

	// Define a list of printer URLs
	printerURLs := []string{
		"https://ender.local.antnsn.dev", // Add the printer URLs here
		// Add more printer URLs as needed
	}

	// Infinite loop for continuous monitoring
	for {
		// Check for new uploads in the "uploads" folder
		filePath := checkForNewUploads()
		if filePath != "" {
			// Iterate through the list of printer URLs
			for _, printerURL := range printerURLs {
				// Send a GET request to the API endpoint to get printer status
				response, err := http.Get(printerURL + "/printer/info")
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
					// Send the file to the printer's ready status URL
					sendFileToOctoPrint(infoResponse.Result, printerURL, filePath)
					break // Exit the loop after sending the file to the first ready printer
				}
			}
		}

		// Wait for a signal that a new upload has been processed
		<-newUploadChan

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

// sendFileToOctoPrint sends a file to OctoPrint
func sendFileToOctoPrint(printer PrinterStatus, printerURL string, filePath string) {
	// Create a POST request to send the file to the printer's upload endpoint
	octoPrintURL := fmt.Sprintf("%s/files/upload", printerURL)
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create a new buffer for the request body
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
	req, err := http.NewRequest("POST", octoPrintURL, body)
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		return
	}

	// Set the API key as a header (replace 'YOUR_API_KEY' with your OctoPrint API key)
	req.Header.Set("X-Api-Key", "YOUR_API_KEY")

	// Set the Content-Type header to indicate multipart/form-data
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the file from the request
	file, fileHeader, err := r.FormFile("file") // "file" should match the field name in the form
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving file: %s", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get the original filename
	originalFilename := fileHeader.Filename

	// Save the file to the "uploads" folder with the original filename
	filePath := fmt.Sprintf("%s/%s", uploadFolder, originalFilename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating file: %s", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error copying file data: %s", err), http.StatusInternalServerError)
		return
	}

	// Signal that a new upload has been processed
	newUploadChan <- struct{}{}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' uploaded successfully!", originalFilename)
}

func startUploadServer() {
	http.HandleFunc(uploadEndpoint, uploadHandler)
	fmt.Printf("File upload server listening on :8081%s...\n", uploadEndpoint)
	http.ListenAndServe(":8081", nil)
}
