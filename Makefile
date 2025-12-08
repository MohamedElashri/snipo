.PHONY: all build run test test-coverage test-short coverage coverage-func lint clean docker docker-run docker-stop dev migrate migrate-down

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

docker-run:
	docker compose up -d

docker-stop:
	docker compose down

migrate:
	go run ./cmd/server migrate

migrate-down:
	go run ./cmd/server migrate down
