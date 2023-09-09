# Stage 1: Build the application
# Use the official Golang image as a parent image
FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .


# Stage 2: Create a smaller runtime image
# Build the Go application
RUN go build -o go-printerfarm .
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-printerfarm .
EXPOSE 8081

# Start the Go application
CMD ["./go-printerfarm"]
