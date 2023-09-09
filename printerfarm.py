import os
import json
import requests
from flask import Flask, request, jsonify, render_template
import time  # Import the time module

app = Flask(__name__)

# Define the array of printer URLs
# Read printer URLs from the environment variable
printerURLs = os.environ.get("PRINTER_URLS", "").split(",")


class PrinterStatus:
    def __init__(self, state, state_message, hostname, klipper_path, python_path, log_file, config_file, software_version, cpu_info):
        self.state = state
        self.state_message = state_message
        self.hostname = hostname
        self.klipper_path = klipper_path
        self.python_path = python_path
        self.log_file = log_file
        self.config_file = config_file
        self.software_version = software_version
        self.cpu_info = cpu_info

# Function to check if a printer is ready
def is_printer_ready(printer_url):
    try:
        response = requests.get(f"{printer_url}/printer/info")
        response.raise_for_status()
        printer_info = response.json().get("result", {})  # Check the "result" field
        printer_status = PrinterStatus(**printer_info)
        return printer_status.state == "ready"
    except Exception as e:
        print(f"Error checking printer status: {e}")
        return False

# Function to check the state of a printer
def get_printer_state(printer_url):
    try:
        response = requests.get(f"{printer_url}/printer/info")
        response.raise_for_status()
        printer_info = response.json().get("result", {})  # Check the "result" field
        printer_status = PrinterStatus(**printer_info)
        return printer_status.state
    except Exception as e:
        print(f"Error checking printer status: {e}")
        return "Error"

# Function to upload a file to a printer
def upload_file_to_printer(file_content, printer_url, filename):
    try:
        # Generate a unique multipart boundary
        boundary = "---------------------------" + str(int(time.time() * 1000))

        # Construct the headers with the proper boundary
        headers = {
            "Content-Type": f"multipart/form-data; boundary={boundary}"
        }

        # Construct the multipart form-data body as bytes
        data = (
            f"--{boundary}\r\n"
            f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'
            f"Content-Type: application/octet-stream\r\n\r\n"
        )

        end_data = f"\r\n--{boundary}--\r\n"

        # Combine all the parts into a single byte array
        body = bytes(data, "utf-8") + file_content + bytes(end_data, "utf-8")

        # Send the request
        response = requests.post(
            f"{printer_url}/server/files/upload",
            headers=headers,
            data=body,
        )
        response.raise_for_status()
        print(f"File uploaded to printer: {printer_url}")
        return True
    except Exception as e:
        print(f"Error uploading file to printer: {e}")
        return False

@app.route('/upload', methods=['POST'])
def upload_file():
    try:
        uploaded_file = request.files['file']
        if uploaded_file:
            # Get the provided filename
            filename = uploaded_file.filename

            # Create the metadata dictionary
            metadata = {
                "item": {
                    "path": filename,  # Use the provided filename as the path
                    "root": "gcodes",
                    "size": uploaded_file.content_length,  # Get the file size
                    "permissions": "rw"
                },
                "print": "true",
                "action": "create_file"
            }

            # Check if any printer is ready
            ready_printer_url = None
            for printer_url in printerURLs:
                if is_printer_ready(printer_url):
                    ready_printer_url = printer_url
                    break

            if ready_printer_url:
                # Upload the file to the ready printer with metadata
                file_content = uploaded_file.read()
                upload_success = upload_file_to_printer(file_content, ready_printer_url, filename)

                if upload_success:
                    return jsonify({"message": "File uploaded to a printer.", "printer_url": ready_printer_url}), 200
                else:
                    return jsonify({"message": "Failed to upload file to printer.", "printer_url": ready_printer_url}), 500
            else:
                return jsonify({"message": "No printer is currently available.", "printer_states": get_printer_states()}), 400
        else:
            return jsonify({"message": "No file was received."}), 400
    except Exception as e:
        return jsonify({"message": f"Error: {str(e)}"}), 500

# Function to get the states of all printers
def get_printer_states():
    printer_states = {}
    for printer_url in printerURLs:
        state = get_printer_state(printer_url)
        printer_states[printer_url] = state
    return printer_states

@app.route('/printer_states', methods=['GET'])
def printer_states():
    return jsonify(get_printer_states())

@app.route('/')
def homepage():
    return render_template('index.html', printerURLs=printerURLs)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
