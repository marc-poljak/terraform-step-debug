.PHONY: build clean test lint install ensure-build-dir

# Binary name
BINARY_NAME=terraform-step-debug
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: test lint build

# Create build directory if it doesn't exist
ensure-build-dir:
	mkdir -p $(BUILD_DIR)

# Build the application
build: ensure-build-dir
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/terraform-step-debug

# Install the application
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin"
	@if [ -z "$(GOPATH)" ]; then \
		echo "GOPATH is not set. Please set it or use go install directly."; \
		exit 1; \
	fi
	@mkdir -p $(GOPATH)/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete. Make sure $(GOPATH)/bin is in your PATH."

# Clean build files
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Run linter
lint: build
	golangci-lint run -E gocyclo -E staticcheck

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Package for distribution
dist: ensure-build-dir
	mkdir -p $(BUILD_DIR)/dist
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/dist/$(BINARY_NAME)_darwin_amd64 ./cmd/terraform-step-debug
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/dist/$(BINARY_NAME)_darwin_arm64 ./cmd/terraform-step-debug
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/dist/$(BINARY_NAME)_linux_amd64 ./cmd/terraform-step-debug

# Help target
help:
	@echo "Available targets:"
	@echo "  build   - Build the application into the $(BUILD_DIR) directory"
	@echo "  install - Install the application to your GOPATH"
	@echo "  clean   - Clean build files"
	@echo "  test    - Run tests"
	@echo "  lint    - Run linter"
	@echo "  run     - Build and run the application"
	@echo "  dist    - Package for distribution into $(BUILD_DIR)/dist"