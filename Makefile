.PHONY: all build run run-test test test-coverage test-short coverage coverage-func lint govulncheck clean docker docker-multiarch docker-run docker-stop dev migrate migrate-down vendor vendor-install vendor-sync vendor-verify vendor-cleanup vendor-check vendor-status vendor-update vendor-update-major

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

migrate-down:
	go run ./cmd/server migrate down

# ── Vendor Library Management ────────────────────────────────────────
# All frontend JS/CSS libs are served locally (no CDN) for privacy.
# node_modules/ is gitignored; only internal/web/static/vendor/ is committed.

# Full one-shot setup: install + sync + verify
vendor: vendor-install vendor-sync vendor-verify

vendor-install:
	@echo "Installing npm dependencies..."
	npm install --no-audit --no-fund
	@echo "Dependencies installed"

vendor-sync: vendor-install
	@echo "Syncing vendor files..."
	@node scripts/sync-vendor.js

# Verify all expected vendor files exist
vendor-verify:
	@echo "Verifying vendor files..."
	@node scripts/verify-vendor.js
	@echo "All vendor files present"

# Remove orphaned files not in sync config
vendor-cleanup:
	@echo "Cleaning orphaned vendor files..."
	@node scripts/verify-vendor.js --cleanup
	@echo "Orphaned files removed"

# Check for outdated packages with colored summary
vendor-check:
	@echo "Checking for outdated packages..."
	@npm outdated 2>/dev/null || echo "  All packages up to date"

# Show current vendor versions + update availability
vendor-status:
	@echo "Vendor library versions:"
	@node scripts/verify-vendor.js --status

# Update minor/patch versions + sync + verify
vendor-update: vendor-install
	@echo "Updating vendor libraries (minor/patch)..."
	@npm update --no-audit --no-fund
	@echo "Syncing updated files..."
	@node scripts/sync-vendor.js
	@echo "Verifying..."
	@node scripts/verify-vendor.js
	@echo "Update summary:"
	@node scripts/verify-vendor.js --status
	@echo "Vendor libraries updated"

# Update including major versions (installs npm-check-updates first)
vendor-update-major:
	@echo "Updating vendor libraries (including major versions)..."
	@if ! npx npm-check-updates --version >/dev/null 2>&1; then \
		echo "  Installing npm-check-updates..."; \
		npm install -D npm-check-updates --no-audit --no-fund; \
	fi
	@npx npm-check-updates -u
	@npm install --no-audit --no-fund
	@echo "Syncing updated files..."
	@node scripts/sync-vendor.js
	@echo "Verifying..."
	@node scripts/verify-vendor.js
	@echo "Update summary:"
	@node scripts/verify-vendor.js --status
	@echo "Vendor libraries updated (major versions included)"

help:
	@echo ""
	@echo "Use 'make <command>' to execute any command."
	@echo ""
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  run-test       - Run the application (no auth, test db)"
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
	@echo "  vendor         - Full setup: install + sync + verify"
	@echo "  vendor-install - Install npm dependencies"
	@echo "  vendor-sync    - Sync vendor files from node_modules"
	@echo "  vendor-verify  - Check all expected vendor files exist"
	@echo "  vendor-cleanup - Remove orphaned vendor files"
	@echo "  vendor-check   - Check for outdated packages"
	@echo "  vendor-status  - Show current vendor versions"
	@echo "  vendor-update  - Update vendor libs (minor/patch)"
	@echo "  vendor-update-major - Update vendor libs (incl. major)"
	@echo "  help           - Show this help message"
