package main

import (
	"encoding/json"
	"fmt"
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
	// Print the monitoring message with the polling interval
	fmt.Printf("Monitoring printers with a polling interval of %s...\n", pollingInterval)

	// Define a list of Mainsail API endpoints for printer status
	statusURLs := []string{
		"https://ender.local.antnsn.dev/printer/info", // Add the new status URL here
	}

	for {
		// Check if a file argument is provided
		if len(os.Args) == 2 {
			filePath := os.Args[1] // Get the file path from the command-line argument

			// Check if the provided file exists
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				// Process the file and send it to a printer
				fmt.Printf("Processing file '%s'...\n", filePath)
				processFile(filePath)
			}
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
				processReadyPrinter(infoResponse.Result, statusURL)
			}
		}

		// Sleep for the specified polling interval
		time.Sleep(pollingInterval)
	}
}

// processReadyPrinter processes the action for a printer with state "ready"
func processReadyPrinter(printer PrinterStatus, statusURL string) {
	// Implement your action here, e.g., send a file to OctoPrint with 'printer.Hostname'
	fmt.Printf("Printer with hostname '%s' is 'ready'. Taking action...\n", printer.Hostname)
	// Add your code to send a file to OctoPrint or perform any other action for this printer
}

// processFile processes a provided file and takes action
func processFile(filePath string) {
	// Implement your action here for the provided file
	// Example: Send the file to a printer
	fmt.Printf("File '%s' processed and sent to a printer.\n", filePath)
}
