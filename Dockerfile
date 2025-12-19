# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Final stage
FROM scratch

# Copy certificates for TLS connections (database, external APIs)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /app/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
