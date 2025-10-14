# WebGPU Triangle - Go WASM Build Makefile

# Variables
GOOS=js
GOARCH=wasm
GOROOT=$(shell go env GOROOT)
WASM_EXEC_JS=$(GOROOT)/misc/wasm/wasm_exec.js
BUILD_DIR=build
ASSETS_DIR=assets
OUTPUT_DIR=dist
PORT=8080

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build clean dev serve test deps help

# Default target
all: deps build

# Help target
help:
	@echo "$(BLUE)WebGPU Triangle - Go WASM Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@echo "  $(GREEN)build$(NC)     - Compile Go to WebAssembly"
	@echo "  $(GREEN)dev$(NC)       - Build and start development server"
	@echo "  $(GREEN)serve$(NC)    - Start HTTP server for testing"
	@echo "  $(GREEN)test$(NC)     - Run Go tests"
	@echo "  $(GREEN)clean$(NC)    - Clean build artifacts"
	@echo "  $(GREEN)deps$(NC)     - Fetch WASM runtime and setup dependencies"
	@echo "  $(GREEN)help$(NC)     - Show this help message"
	@echo ""

# Setup dependencies and fetch WASM runtime
deps:
	@echo "$(BLUE)Setting up dependencies...$(NC)"
	@mkdir -p $(BUILD_DIR) $(OUTPUT_DIR)
	@echo "$(YELLOW)Fetching WASM runtime from Go installation...$(NC)"
	@if [ -f "$(WASM_EXEC_JS)" ]; then \
		cp "$(WASM_EXEC_JS)" $(ASSETS_DIR)/wasm_exec.js; \
		echo "$(GREEN)✓ Copied wasm_exec.js from $(WASM_EXEC_JS)$(NC)"; \
	else \
		echo "$(RED)✗ wasm_exec.js not found at $(WASM_EXEC_JS)$(NC)"; \
		echo "$(YELLOW)Trying alternative location...$(NC)"; \
		find $(GOROOT) -name "wasm_exec.js" -exec cp {} $(ASSETS_DIR)/wasm_exec.js \; 2>/dev/null || \
		echo "$(RED)✗ Could not find wasm_exec.js in Go installation$(NC)"; \
	fi
	@echo "$(GREEN)✓ Dependencies setup complete$(NC)"

# Build the WebAssembly binary
build: deps
	@echo "$(BLUE)Building WebAssembly binary...$(NC)"
	@echo "$(YELLOW)Compiling Go to WASM with GOOS=js GOARCH=wasm$(NC)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/main.wasm ./cmd/webgpu-triangle
	@if [ -f "$(BUILD_DIR)/main.wasm" ]; then \
		echo "$(GREEN)✓ WebAssembly binary built successfully$(NC)"; \
		ls -lh $(BUILD_DIR)/main.wasm; \
	else \
		echo "$(RED)✗ Build failed$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Copying assets to output directory...$(NC)"
	@cp -r $(ASSETS_DIR)/* $(OUTPUT_DIR)/ 2>/dev/null || true
	@cp $(BUILD_DIR)/main.wasm $(OUTPUT_DIR)/
	@echo "$(GREEN)✓ Build complete - files in $(OUTPUT_DIR)/$(NC)"

# Development build with verbose output
dev: clean deps
	@echo "$(BLUE)Development build...$(NC)"
	@echo "$(YELLOW)Building with debug information...$(NC)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(BUILD_DIR)/main.wasm ./cmd/webgpu-triangle
	@cp -r $(ASSETS_DIR)/* $(OUTPUT_DIR)/ 2>/dev/null || true
	@cp $(BUILD_DIR)/main.wasm $(OUTPUT_DIR)/
	@echo "$(GREEN)✓ Development build complete$(NC)"
	@echo "$(YELLOW)Starting development server...$(NC)"
	@echo "$(BLUE)Open http://localhost:$(PORT) in your browser$(NC)"
	@cd $(OUTPUT_DIR) && python3 -m http.server $(PORT)

# Start HTTP server for testing
serve:
	@echo "$(BLUE)Starting HTTP server...$(NC)"
	@if [ ! -f "$(OUTPUT_DIR)/main.wasm" ] || [ ! -d "$(OUTPUT_DIR)" ]; then \
		echo "$(YELLOW)No build found, building first...$(NC)"; \
		$(MAKE) build; \
	fi
	@echo "$(GREEN)✓ Server starting at http://localhost:$(PORT)$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@cd $(OUTPUT_DIR) && python3 -m http.server $(PORT)

# Run tests
test:
	@echo "$(BLUE)Running Go tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(OUTPUT_DIR)
	@echo "$(GREEN)✓ Clean complete$(NC)"

# Show build information
info:
	@echo "$(BLUE)Build Information:$(NC)"
	@echo "  Go version: $(shell go version)"
	@echo "  GOROOT: $(GOROOT)"
	@echo "  GOOS: $(GOOS)"
	@echo "  GOARCH: $(GOARCH)"
	@echo "  WASM runtime: $(WASM_EXEC_JS)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Output directory: $(OUTPUT_DIR)"

# Quick build and serve
quick: build serve

# Production build
prod: clean deps
	@echo "$(BLUE)Production build...$(NC)"
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o $(BUILD_DIR)/main.wasm ./cmd/webgpu-triangle
	@cp -r $(ASSETS_DIR)/* $(OUTPUT_DIR)/ 2>/dev/null || true
	@cp $(BUILD_DIR)/main.wasm $(OUTPUT_DIR)/
	@echo "$(GREEN)✓ Production build complete (optimized)$(NC)"
	@ls -lh $(OUTPUT_DIR)/main.wasm
