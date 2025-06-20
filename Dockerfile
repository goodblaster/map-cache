# Stage 1: Build using standard Golang image
FROM golang:1.24 as builder

ENV CGO_ENABLED=0
WORKDIR /app

# Build arguments
ARG VERSION
ARG COMMIT
ARG DATE

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags "\
    -X 'your/module/path/build.Version=${VERSION}' \
    -X 'your/module/path/build.Commit=${COMMIT}' \
    -X 'your/module/path/build.Date=${DATE}'" \
    -o map-cache ./cmd/cache/main.go

# Stage 2: Runtime image
FROM ubuntu:22.04
COPY --from=builder /app/map-cache /usr/local/bin/map-cache

USER root
EXPOSE 80

ENV LISTEN_ADDRESS=":80"
ENV LOG_FORMAT="json"

ENTRYPOINT ["map-cache"]
