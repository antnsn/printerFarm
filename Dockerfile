# Stage 1: Build the application
FROM golang:1.16-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o go-printerfarm

# Stage 2: Create a smaller runtime image
FROM scratch
COPY --from=build /app/go-printerfarm /
EXPOSE 8080
CMD ["/go-printerfarm"]
