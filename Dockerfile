# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install CA certificates
RUN apk add --no-cache ca-certificates

# Add go module files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Add source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dyndns-updater

# ---

FROM scratch

# Copy binary and root certs
COPY --from=builder /app/dyndns-updater /dyndns-updater
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/dyndns-updater"]
