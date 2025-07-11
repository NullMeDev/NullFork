# Dockerfile for Enhanced Gateway Scraper

# Use a minimal base image with Go installed
FROM golang:1.18-alpine

# Set environment variables for maximum efficiency
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install necessary packages
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files for dependency resolution
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source
COPY . .

# Build the application
RUN go build -o /gateway-scraper ./cmd/scraper/main.go

# Expose necessary ports
EXPOSE 8080
EXPOSE 9090

# Set the entry point
CMD ["/gateway-scraper"]

