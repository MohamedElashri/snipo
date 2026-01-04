.PHONY: all build run test test-coverage test-short coverage coverage-func lint clean docker docker-multiarch docker-run docker-stop dev migrate migrate-down vendor-install vendor-sync vendor-check vendor-update vendor-update-major

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-w -s -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

all: build

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/snipo ./cmd/server

run: build
	./bin/snipo serve

dev:
	go run ./cmd/server serve

test:
	go test -v -race ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	@echo "\n=== Coverage Summary ==="
	@go tool cover -func=coverage.out | tail -1

test-short:
	go test -short ./...

coverage:
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

coverage-func:
	go tool cover -func=coverage.out

lint:
	golangci-lint run

clean:
	rm -rf bin/ coverage.out coverage.html data/

docker:
	docker build -t snipo:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) .

docker-multiarch:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		-t snipo:$(VERSION) \
		--load .

docker-run:
	docker compose up -d

docker-stop:
	docker compose down

migrate:
	go run ./cmd/server migrate

migrate-down:
	go run ./cmd/server migrate down

vendor-install:
	@echo "Installing npm dependencies..."
	npm install
	@echo "Dependencies installed"

vendor-sync:
	@echo "Syncing vendor files..."
	npm run vendor:sync

vendor-check:
	@echo "Checking for outdated packages..."
	npm outdated || true

vendor-update:
	@echo "Updating vendor libraries (minor/patch)..."
	npm update
	npm run vendor:sync
	@echo "Vendor libraries updated"

vendor-update-major:
	@echo "Updating vendor libraries (including major versions)..."
	npx npm-check-updates -u
	npm install
	npm run vendor:sync
	@echo "Vendor libraries updated (major versions included)"

help:
	@echo ""
	@echo "Use 'make <command>' to execute any command."
	@echo ""
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  dev            - Run in development mode"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run short tests"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker         - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose"
	@echo "  migrate        - Run database migrations"
	@echo "  vendor-install - Install npm dependencies"
	@echo "  vendor-sync    - Sync vendor files"
	@echo "  vendor-check   - Check for outdated packages"
	@echo "  vendor-update  - Update vendor libraries (minor/patch)"
	@echo "  vendor-update-major - Update vendor libraries (including major versions)"
	@echo "  help           - Show this help message"
