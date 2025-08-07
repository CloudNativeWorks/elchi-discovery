# Claude Instructions

This repository contains the elchi-discovery service for discovering Kubernetes cluster nodes and endpoints.

## Development Commands

When working on this codebase, please use the following commands:

### Build and Test
- `go build .` - Build the elchi-discovery binary
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up module dependencies
- `go mod download` - Download dependencies

### Development
- `go run .` - Run the elchi-discovery service locally
- `docker build -t elchi-discovery .` - Build Docker image
- `helm install elchi-discovery ./helm` - Deploy to Kubernetes

## Code Standards

- Follow existing Go conventions and patterns
- Use the internal packages for common functionality (logger, config, context)
- Ensure all tests pass before committing
- Run `go mod tidy` before submitting changes
- Use Go 1.21+ features when appropriate

## Project Structure

This is a standalone Kubernetes node discovery service:

```
elchi-discovery/
├── main.go                  ← Main application entry point
├── api/                     ← API client for sending discovery results
│   └── client.go
├── discovery/               ← Core discovery logic
│   ├── discovery.go
│   └── types.go
├── internal/                ← Internal packages
│   ├── config/             ← Configuration management
│   ├── context/            ← Context utilities
│   └── logger/             ← Logging utilities
├── helm/                    ← Helm chart for Kubernetes deployment
├── config.yaml             ← Configuration file
└── Dockerfile              ← Container image definition
```

## Environment Variables

- `CLUSTER_NAME` - Required: Name of the Kubernetes cluster
- `DISCOVERY_INTERVAL` - Discovery interval in seconds (default: 30)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)
- `LOG_FORMAT` - Log format: text or json (default: text)
- `ELCHI_TOKEN` - Optional: Authentication token for API
- `ELCHI_API_ENDPOINT` - Optional: API endpoint to send discovery results
- `ELCHI_INSECURE_SKIP_VERIFY` - Skip TLS verification for API calls (default: false)

## Running in Kubernetes

This service is designed to run inside a Kubernetes cluster. It uses in-cluster configuration to connect to the Kubernetes API and requires appropriate RBAC permissions to list nodes.