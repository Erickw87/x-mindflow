.PHONY: build test lint run clean docker-build docker-run help

BINARY_NAME=mock-server
DOCKER_IMAGE=mock-service
VERSION?=v1.0.0

help:
	@echo "Mock Service Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the binary"
	@echo "  make test           Run tests"
	@echo "  make lint           Run linter"
	@echo "  make run            Run the service locally"
	@echo "  make clean          Clean build artifacts"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-run     Run in Docker"

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) cmd/mock-server/main.go

test:
	@echo "Running tests..."
	go test -v -cover ./...

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1)
	golangci-lint run ./...

run: build
	@echo "Running $(BINARY_NAME)..."
	./bin/$(BINARY_NAME) --config configs/user-service-v1.1.yaml

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

docker-run:
	@echo "Running in Docker..."
	docker run -p 8080:8081 -v $(PWD)/configs:/app/configs $(DOCKER_IMAGE):latest --config /app/configs/user-service-v1.1.yaml

.PHONY: init
init:
	@echo "Initializing project..."
	go mod tidy
	go mod download
