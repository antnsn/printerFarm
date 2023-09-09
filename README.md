# Project Name: PrinterFarm - 3D Printer File Distribution

Description:
PrinterFarm is a specialized Flask-based application designed to simplify the distribution of 3D printing files to a fleet of 3D printers. The project's core objective is to automate the selection of an available 3D printer that is in the "ready" state and send the print job to it. Key features of the project include:

## 1. Auto Printer Selection:

PrinterFarm intelligently selects an available 3D printer that is in the "ready" state from a pool of configured printers. This automation eliminates the need for users to manually choose a printer for each print job.
## 2. File Upload Endpoint:

Users can easily send 3D model or G-code files to PrinterFarm using a dedicated API endpoint. The application takes care of file handling and distribution to the selected printer.

## 3. Real-time Printer Status:

PrinterFarm provides real-time status information about each connected 3D printer. Users can monitor the readiness and states of their printers through the web interface.

## 4. Error Handling and Notifications:

The application includes robust error handling to manage cases where printers are unavailable or file uploads encounter issues. Users receive clear notifications about the success or failure of their print jobs.
Getting Started:
To start using PrinterFarm, users need to have Docker installed on their systems. The process involves cloning the project repository, building the Docker image, running a Docker container, and accessing the PrinterFarm web interface through a web browser.

Contributing:
PrinterFarm is an open-source project that welcomes contributions from the community. Contributors can follow standard Git and GitHub practices to fork the repository, create branches for new features or bug fixes, and submit pull requests to the main repository.