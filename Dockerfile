# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY ./app /app

# Build the Go application
RUN go build -o printer-monitor

# Expose the port for communication (if needed)
EXPOSE 8080

# Run the printer-monitor binary when the container launches
CMD ["./printer-monitor"]
