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

# Create non-root user
RUN useradd -m appuser

# Copy the binary from the builder stage
COPY --from=builder /app/map-cache /usr/local/bin/map-cache

# Set ownership and switch to non-root user
RUN chown appuser:appuser /usr/local/bin/map-cache
USER appuser

# Expose the app's port
EXPOSE 8080

# Run the app
ENTRYPOINT ["map-cache"]
