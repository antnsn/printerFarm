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
	// Create the "uploaded_files" directory if it doesn't exist
	if _, err := os.Stat("uploaded_files"); os.IsNotExist(err) {
		err := os.Mkdir("uploaded_files", os.ModePerm)
		if err != nil {
			fmt.Println("Error creating 'uploaded_files' directory:", err)
			return
		}
	}

	// Print the monitoring message with the polling interval
	fmt.Printf("Monitoring printers with a polling interval of %s...\n", pollingInterval)

	// Create a channel to signal when a new file is uploaded
	uploadedFileCh := make(chan string)

	// Start a goroutine to handle file uploads
	go startUploadServer(uploadedFileCh)

	// Infinite loop for continuous monitoring
	for {
		// Check if there's a new file available for processing
		select {
		case newFilePath := <-uploadedFileCh:
			// Define a list of Mainsail API endpoints for printer status
			statusURLs := []string{
				"https://ender.local.antnsn.dev/printer/info", // Add the new status URL here
			}

			// Iterate through the list of printer status URLs
			for _, statusURL := range statusURLs {
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
					processReadyPrinter(infoResponse.Result, statusURL, newFilePath)
				}
			}

			// Remove the processed file
			os.Remove(newFilePath)
		default:
			// Sleep for the specified polling interval if no new file is uploaded
			time.Sleep(pollingInterval)
		}
	}
}

// getNewUploadedFile checks if there's a new uploaded file and returns its path if available
func startUploadServer(uploadedFileCh chan<- string) {
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the incoming form data, including files
		err := r.ParseMultipartForm(10 * 1024 * 1024) // 10 MB limit
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing form: %s", err), http.StatusBadRequest)
			return
		}

		// Get the file from the request
		file, _, err := r.FormFile("file") // "file" should match the field name in the form
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving file: %s", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a new file on the server to save the uploaded file
		newFilePath := fmt.Sprintf("uploaded_files/%d_%s", time.Now().Unix(), "uploaded_file.txt") // Unique filename
		newFile, err := os.Create(newFilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating file: %s", err), http.StatusInternalServerError)
			return
		}
		defer newFile.Close()

		// Copy the uploaded file data to the new file
		_, err = io.Copy(newFile, file)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error copying file data: %s", err), http.StatusInternalServerError)
			return
		}

		// Signal that a new file is uploaded
		uploadedFileCh <- newFilePath

		// Respond with a success message
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "File uploaded successfully!")
	})

	fmt.Printf("File upload server listening on :8081/upload...\n")
	http.ListenAndServe(":8081", nil)
}

// processReadyPrinter processes the action for a printer with state "ready"
func processReadyPrinter(printer PrinterStatus, statusURL string, filePath string) {
	// Implement your action here, e.g., send a file to OctoPrint with 'printer.Hostname'
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