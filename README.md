# Go Script for Printer Monitoring and File Sending

This Go script is designed for monitoring multiple 3D printers' statuses and automatically sending a file to a printer when it's in the "ready" state. It communicates with each printer's API endpoint to retrieve its status and takes action based on the status.

## Features:

- Monitors multiple 3D printers.
- Automatically sends a file to a printer when it's "ready."
- Flexible and extensible for integrating with different printer APIs.

## Usage:

1. Define the list of printer API endpoints in the `statusURLs` variable.
2. The script iterates through the printer URLs and retrieves their statuses.
3. If a printer is found in the "ready" state, it triggers an action (e.g., file sending) for that printer.
4. Customize the action in the `processReadyPrinter` function as needed.

## Configuration:

- Customize the list of printer URLs in the `statusURLs` variable.
- Modify the action taken when a printer is "ready" in the `processReadyPrinter` function.

## Prerequisites:

- Printers accepting the same sliced gcode
- Printers must be running mainsail and klipper
- Go (Golang) installed on your system.
- Access to the printer API endpoints.
