.PHONY: build run test clean docker-build docker-run docker-setup help

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with docker-compose"
	@echo "  docker-stop  - Stop docker-compose services"
	@echo "  docker-setup - Setup Docker environment"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"

# Build the application
build:
	go build -o bin/ipssl-client .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t ipssl-client .

# Setup Docker environment
docker-setup:
	@echo "Setting up Docker environment..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# Run with docker-compose
docker-run:
	@echo "Starting Docker services..."
	@cd docker && docker-compose up -d

# Stop docker-compose services
docker-stop:
	@echo "Stopping Docker services..."
	@cd docker && docker-compose down

# Create required directories (legacy - use docker-setup instead)
setup-dirs:
	mkdir -p docker/data/caddy/{data,config,logs,webroot,ipssl}
	mkdir -p docker/config/caddy
	mkdir -p bin

# Install development dependencies
install-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Full setup (legacy - use docker-setup instead)
setup: setup-dirs deps
	@echo "Setup complete. Copy docker/env.example to docker/.env and configure your settings."
