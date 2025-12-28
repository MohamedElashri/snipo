.PHONY: build install clean test run help

BINARY_NAME=snipo-tui
BUILD_DIR=bin
INSTALL_PATH=/usr/local/bin

help:
	@echo "Snipo CLI - Makefile commands:"
	@echo ""
	@echo "  make build      Build the binary"
	@echo "  make install    Install to $(INSTALL_PATH)"
	@echo "  make run        Run without installing"
	@echo "  make clean      Remove build artifacts"
	@echo "  make test       Run tests"
	@echo "  make deps       Download dependencies"
	@echo ""

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/snipo-tui
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✓ Installed successfully!"
	@echo "Run '$(BINARY_NAME) config' to configure"

run:
	@echo "Running $(BINARY_NAME)..."
	go run ./cmd/snipo-tui

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	go test -v ./...

.DEFAULT_GOAL := help
