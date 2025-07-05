# kwatch - Build and Installation Makefile

.PHONY: build install uninstall clean test deps dev help

# Variables
BINARY_NAME := kwatch
MAIN_FILE := main.go
BUILD_DIR := build
INSTALL_DIR := /usr/local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

# Default target
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Install globally (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully!"
	@echo "Usage: $(BINARY_NAME) ."

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled."

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Development build (with race detection)
dev: deps
	@echo "Building development version..."
	@mkdir -p $(BUILD_DIR)
	go build -race -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(MAIN_FILE)
	@echo "Development binary built: $(BUILD_DIR)/$(BINARY_NAME)-dev"

# Cross-compile for different platforms
build-all: deps
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	
	@echo "Cross-compilation complete!"

# Run tests
test: deps
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Quick install (local user only, no sudo required)
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@chmod +x ~/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed to ~/bin/"
	@echo "Make sure ~/bin is in your PATH"
	@echo "Add this to your shell config: export PATH=\"\$$HOME/bin:\$$PATH\""

# Check if installed correctly
check: 
	@echo "Checking installation..."
	@which $(BINARY_NAME) || echo "$(BINARY_NAME) not found in PATH"
	@$(BINARY_NAME) --help 2>/dev/null || echo "$(BINARY_NAME) not executable"

# Development server (runs in current directory)
run: build
	@echo "Starting $(BINARY_NAME) in current directory..."
	@$(BUILD_DIR)/$(BINARY_NAME) .

# Package for distribution
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(BUILD_DIR)/dist
	
	# Create tar.gz for Unix systems
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	@tar -czf $(BUILD_DIR)/dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	
	# Create zip for Windows
	@zip -j $(BUILD_DIR)/dist/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	
	@echo "Distribution packages created in $(BUILD_DIR)/dist/"

# Help
help:
	@echo "kwatch - Build and Installation Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the binary"
	@echo "  make install      - Install globally (requires sudo)"
	@echo "  make install-user - Install to ~/bin (no sudo required)"
	@echo "  make uninstall    - Uninstall globally"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make dev          - Build development version"
	@echo "  make run          - Build and run in current directory"
	@echo "  make build-all    - Cross-compile for all platforms"
	@echo "  make package      - Create distribution packages"
	@echo "  make check        - Check if installed correctly"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  make install      # Install globally"
	@echo "  kwatch .          # Start monitoring current directory"
	@echo "  kwatch . status   # Get status"