.PHONY: proto build run clean test

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@mkdir -p api/proto/user
	@protoc --proto_path=proto \
		--go_out=api/proto --go_opt=paths=source_relative \
		--go-grpc_out=api/proto --go-grpc_opt=paths=source_relative \
		proto/user/user.proto
	@echo "Protobuf code generated successfully"

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/distributed-service main.go
	@echo "Build completed"

# Run the application
run:
	@echo "Starting application..."
	@go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...
	@echo "Tests completed"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "Linting completed"

# Generate swagger docs
swagger:
	@echo "Generating swagger documentation..."
	@swag init
	@echo "Swagger documentation generated"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t distributed-service:latest .
	@echo "Docker image built"

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker-compose up -d
	@echo "Docker container started"

# Docker stop
docker-stop:
	@echo "Stopping Docker container..."
	@docker-compose down
	@echo "Docker container stopped"

# Help
help:
	@echo "Available commands:"
	@echo "  proto        - Generate protobuf code"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  swagger      - Generate swagger docs"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-stop  - Stop Docker container"
	@echo "  help         - Show this help message" 