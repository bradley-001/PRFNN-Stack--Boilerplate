# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=api
BINARY_UNIX=$(BINARY_NAME)_unix

.PHONY: all build clean test coverage deps dev prod tools generate-models

all: test build

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v main.go

# Clean build files
clean:
	$(GOCLEAN)
	rm -rf bin/

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run in development mode with hot reload (requires air)
dev:
	air

# Run in production mode
prod: build
	./bin/$(BINARY_NAME)

# Build tools
tools:
	$(GOBUILD) -o bin/tools ./src/cmd/tools/main.go

# Generate models from database
generate-models: tools
	./bin/tools -generate-models

# Install development dependencies
install-dev:
	go install github.com/air-verse/air@latest
	air init

# Start the application
run:
	$(GOCMD) run main.go

# Show help
help:
	@echo "Available commands:"
	@echo "  build          Build the application"
	@echo "  clean          Clean build files"
	@echo "  test           Run tests"
	@echo "  deps           Download dependencies"
	@echo "  dev            Run in development mode (requires air, run install-dev first if haven't already)"
	@echo "  prod           Build and run in production mode"
	@echo "  tools          Build CLI tools"
	@echo "  generate-models Generate models from database"
	@echo "  install-dev    Install development dependencies"
	@echo "  run            Run the application (no hot-reload)"
	@echo "  help           Show this help message"
