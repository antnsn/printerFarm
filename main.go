package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Result struct {
		StateMessage    string `json:"state_message"`
		KlipperPath     string `json:"klipper_path"`
		ConfigFile      string `json:"config_file"`
		SoftwareVersion string `json:"software_version"`
		Hostname        string `json:"hostname"`
		CPUInfo         string `json:"cpu_info"`
		State           string `json:"state"`
		PythonPath      string `json:"python_path"`
		LogFile         string `json:"log_file"`
	} `json:"result"`
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {

	var printers [3]string
	printers[0] = "http://sw.home"
	printers[1] = "http://v0.home"
	printers[2] = "http://10.0.0.20"

	for i := 0; i < len(printers); i++ {
		// Get request
		resp, err := http.Get(printers[i] + "/printer/info")
		if err != nil {
			fmt.Println("No response from request")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body) // response body is []byte

		var result Response
		if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}
		fmt.Print("Printer: ")
		fmt.Println(result.Result.Hostname)
		fmt.Print("State: ")
		fmt.Println(result.Result.State)

	}
}
