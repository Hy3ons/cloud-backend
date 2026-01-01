# VM Controller

This is a VM Controller backend service built with Go and Gin.

## Prerequisites

- Go 1.20+
- Kubernetes cluster (or a configured kubeconfig)

## Project Structure

- `cmd/server`: Application entry point.
- `internal/api`: API handlers and routes.
- `internal/config`: Configuration management.
- `internal/services`: Business logic and external services (e.g., K8s).

## Getting Started

1. **Clone the repository** (if applicable).
2. **Install dependencies**:
   ```bash
   go mod tidy
   ```
3. **Setup Environment**:
   Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
4. **Run the server**:
   ```bash
   go run cmd/server/main.go
   ```

## Endpoints

- `GET /health`: Health check and K8s connectivity status.

## Kubernetes Connection

The service attempts to connect to a Kubernetes cluster using In-Cluster config, which works seamlessly if deployed within a cluster. For local development, ensure your `KUBECONFIG` is set correctly or modify the service to support local kubeconfig files explicitly if needed.
