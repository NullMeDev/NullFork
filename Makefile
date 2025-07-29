BINARY_NAME=gateway-scraper
VERSION=$(shell cat VERSION 2>/dev/null || echo "v1.2.0")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR=./dist
CMD_DIR=./cmd/gateway-scraper
GO_LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

.PHONY: all build build-cli build-gui clean test run-cli run-gui help install

# Default target
all: clean build

# Build both CLI and GUI versions
build:
	@echo "🔨 Building Enhanced Gateway Scraper $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	@echo "📦 Building CLI version..."
	@go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-cli $(CMD_DIR)
	@echo "📦 Building GUI version..."
	@go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-gui $(CMD_DIR)
	@echo "🔗 Creating compatibility binaries..."
	@cp $(BUILD_DIR)/$(BINARY_NAME)-cli $(BUILD_DIR)/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-cli $(BUILD_DIR)/scraper
	@chmod +x $(BUILD_DIR)/*
	@echo "✅ Build complete! Files created in $(BUILD_DIR)"

# Build only CLI version
build-cli:
	@echo "📦 Building CLI version only..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-cli $(CMD_DIR)
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-cli
	@echo "✅ CLI version built: $(BUILD_DIR)/$(BINARY_NAME)-cli"

# Build only GUI version  
build-gui:
	@echo "📦 Building GUI version only..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-gui $(CMD_DIR)
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-gui
	@echo "✅ GUI version built: $(BUILD_DIR)/$(BINARY_NAME)-gui"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Run CLI version
run-cli: build-cli
	@echo "🚀 Starting Gateway Scraper (CLI Mode)..."
	@$(BUILD_DIR)/$(BINARY_NAME)-cli

# Run GUI version
run-gui: build-gui
	@echo "🚀 Starting Gateway Scraper (GUI Mode)..."
	@echo "📱 Web interface will be available at: http://localhost:8081"
	@$(BUILD_DIR)/$(BINARY_NAME)-gui -gui

# Install to system (requires sudo)
install: build
	@echo "📦 Installing to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME)-cli /usr/local/bin/$(BINARY_NAME)-cli
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME)-gui /usr/local/bin/$(BINARY_NAME)-gui
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installation complete"

# Show help
help:
	@echo "Enhanced Gateway Scraper - Build System"
	@echo "======================================="
	@echo ""
	@echo "Available commands:"
	@echo "  make build      - Build both CLI and GUI versions"
	@echo "  make build-cli  - Build only CLI version"
	@echo "  make build-gui  - Build only GUI version"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make run-cli    - Build and run CLI version"
	@echo "  make run-gui    - Build and run GUI version"
	@echo "  make install    - Install to system (requires sudo)"
	@echo "  make help       - Show this help"
	@echo ""
	@echo "Quick usage after building:"
	@echo "  CLI: $(BUILD_DIR)/$(BINARY_NAME)-cli"
	@echo "  GUI: $(BUILD_DIR)/$(BINARY_NAME)-gui -gui"
