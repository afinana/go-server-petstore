## Build Stage
FROM golang:1.25-alpine AS build

WORKDIR /app

# Install ca-certificates for secure outbound calls
RUN apk --no-cache add ca-certificates

# Cache dependencies first (improves build caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY petstore/ ./petstore/
COPY main.go .

# Build static binary with stripped debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go-server-petstore main.go

## Production Stage (Minimal & Secure)
FROM scratch

# Import ca-certificates from builder
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=build /go-server-petstore /go-server-petstore

# Run as non-root user (nobody:nobody)
USER 65534:65534

# Expose port
EXPOSE 8080

ENTRYPOINT ["/go-server-petstore"]