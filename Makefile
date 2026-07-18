.PHONY: all build run run-test demo test test-coverage test-short coverage coverage-func lint govulncheck clean docker docker-multiarch docker-run docker-stop dev migrate vendor vendor-update vendor-clean chrome firefox extension-build

VERSION ?= $(shell grep 'const Current =' internal/version/version.go | cut -d '"' -f 2)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-w -s -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

all: build

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/snipo ./cmd/server

run: build
	./bin/snipo serve

run-test: build
	SNIPO_DISABLE_AUTH=true SNIPO_DB_PATH=./snipo.db ./bin/snipo serve

demo: build
	@echo "Starting Snipo in demo mode (password: demo)..."
	SNIPO_DEMO_MODE=true SNIPO_DEMO_RESET_INTERVAL=15m SNIPO_DB_PATH=./demo-snipo.db ./bin/snipo serve

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

govulncheck:
	govulncheck ./...

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

migrate-test:
	SNIPO_DISABLE_AUTH=true SNIPO_DB_PATH=./snipo.db go run ./cmd/server migrate

# ── Vendor Library Management ────────────────────────────────────────
# All frontend JS/CSS libs are served locally (no CDN) for privacy.
# npm postinstall automatically syncs files to internal/web/static/vendor/.

vendor:
	npm install --no-audit --no-fund

vendor-update:
	npm update --no-audit --no-fund
	@echo "\nInstalled versions:"
	@node scripts/sync-vendor.js --status

vendor-clean:
	@node scripts/sync-vendor.js --cleanup

extension-build:
	@echo "Building extension packages..."
	@cd extension && ./build.sh all

chrome:
	@echo "Building Chrome extension package..."
	@cd extension && ./build.sh chrome

firefox:
	@echo "Building Firefox extension package..."
	@cd extension && ./build.sh firefox

help:
	@echo ""
	@echo "Use 'make <command>' to execute any command."
	@echo ""
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  run-test       - Run the application (no auth, test db)"
	@echo "  demo           - Run the application in demo mode (password: demo)"
	@echo "  dev            - Run in development mode"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run short tests"
	@echo "  lint           - Run linter"
	@echo "  govulncheck    - Run vulnerability check"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker         - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-test   - Run database migrations (no auth, test db)"
	@echo "  vendor         - Install npm deps & sync vendor files (auto via postinstall)"
	@echo "  vendor-update  - Update vendor libs to latest (respecting semver)"
	@echo "  vendor-clean   - Remove orphaned files from vendor/"
	@echo "  chrome         - Build Chrome extension zip"
	@echo "  firefox        - Build Firefox extension zip + source archive"
	@echo "  extension-build - Build Chrome and Firefox extension packages"
	@echo "  help           - Show this help message"
