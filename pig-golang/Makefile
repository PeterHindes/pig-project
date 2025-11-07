.PHONY: all build run test clean install help dev

# Default target
all: build

# Build the server
build:
	@echo "Building pig-server..."
	go build -o pig-server main.go
	@echo "Build complete!"

# Run the server
run:
	@echo "Starting pig-server..."
	go run main.go

# Run with custom port
run-custom:
	@echo "Starting pig-server on custom port..."
	go run main.go -port=8080 -wsurl=ws://localhost:8080

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f pig-server
	rm -f pig-server-*
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Development mode (with file watching would require additional tools)
dev:
	@echo "Starting development server..."
	go run main.go -port=8080 -wsurl=ws://localhost:8080

# Build for multiple platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -o pig-server-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o pig-server-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o pig-server-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o pig-server-windows-amd64.exe main.go
	@echo "Build complete for all platforms!"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete!"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Check for security issues (requires gosec)
security:
	@echo "Checking for security issues..."
	gosec ./...

# Run the example client (opens in browser)
client:
	@echo "Opening example client in browser..."
	@if command -v open > /dev/null; then \
		open example-client.html; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open example-client.html; \
	elif command -v start > /dev/null; then \
		start example-client.html; \
	else \
		echo "Please open example-client.html in your browser manually"; \
	fi

# Help target
help:
	@echo "Pig Game Server - Available targets:"
	@echo ""
	@echo "  make build          - Build the server binary"
	@echo "  make run            - Run the server"
	@echo "  make run-custom     - Run the server with custom settings"
	@echo "  make install        - Install dependencies"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make dev            - Run in development mode"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code (requires golangci-lint)"
	@echo "  make vet            - Vet code"
	@echo "  make security       - Check for security issues (requires gosec)"
	@echo "  make client         - Open example client in browser"
	@echo "  make help           - Show this help message"
	@echo ""
