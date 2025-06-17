# Stage 1: Build using standard Golang image
FROM golang:1.24 as builder

# Disable CGO for static binary
ENV CGO_ENABLED=0
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o map-cache ./cmd/cache/main.go

# Stage 2: Minimal Ubuntu image for runtime
FROM ubuntu:22.04

# Copy the binary from the builder stage
COPY --from=builder /app/map-cache /usr/local/bin/map-cache

# Run as root for port 80
USER root

# Expose the app's port
EXPOSE 80

# Listen in port 80
ENV LISTEN_ADDRESS=":80"

# Set the default log format to JSON
ENV LOG_FORMAT="json"

# Run the app
ENTRYPOINT ["map-cache"]
