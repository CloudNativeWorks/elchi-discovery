# Multi-stage build for elchi-discovery
FROM golang:1.23.1-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git ca-certificates

# Get build arguments
ARG PROJECT_VERSION
ARG TARGETARCH=amd64

# Set working directory
WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${PROJECT_VERSION}" \
    -o /elchi-discovery \
    .

# Final minimal image
FROM scratch

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /elchi-discovery /elchi-discovery

# Use non-root user ID
USER 65534

# Set entrypoint
ENTRYPOINT ["/elchi-discovery"]