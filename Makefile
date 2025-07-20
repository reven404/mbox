# Makefile for mdriver

# Build variables
BINARY_NAME=mdriver
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILT_BY ?= make

# Go build flags
LDFLAGS=-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=$(BUILT_BY)
BUILD_FLAGS=-trimpath -ldflags="$(LDFLAGS)"

# Platform targets
PLATFORMS=linux/amd64 linux/arm64 windows/amd64 windows/386 darwin/amd64 darwin/arm64

.PHONY: all build clean test fmt vet deps release-test help

all: build

## Build the binary for the current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

## Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		export GOOS=$$(echo $$platform | cut -d/ -f1); \
		export GOARCH=$$(echo $$platform | cut -d/ -f2); \
		export CGO_ENABLED=1; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		if [ "$$GOOS" = "windows" ] && [ "$$GOARCH" = "386" ]; then \
			go build $(BUILD_FLAGS) -tags=legacy,windowsxp -o $(BINARY_NAME)-$$GOOS-$$GOARCH.exe .; \
		elif [ "$$GOOS" = "windows" ]; then \
			go build $(BUILD_FLAGS) -o $(BINARY_NAME)-$$GOOS-$$GOARCH.exe .; \
		else \
			go build $(BUILD_FLAGS) -o $(BINARY_NAME)-$$GOOS-$$GOARCH .; \
		fi; \
	done

## Build Windows XP compatible version
build-xp:
	@echo "Building Windows XP compatible version..."
	GOOS=windows GOARCH=386 CGO_ENABLED=1 go build $(BUILD_FLAGS) -tags=legacy,windowsxp -o $(BINARY_NAME)-windows-xp.exe .

## Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-*
	@go clean

## Run tests
test:
	@echo "Running tests..."
	go test -race -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

## Run linter
vet:
	@echo "Running vet..."
	go vet ./...

## Lint with golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

## Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

## Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	install -m 755 $(BINARY_NAME) $(DESTDIR)/usr/local/bin/$(BINARY_NAME)

## Test release process with GoReleaser
release-test:
	@echo "Testing release process..."
	goreleaser release --snapshot --clean

## Run GoReleaser
release:
	@echo "Creating release..."
	goreleaser release --clean

## Show version
version:
	@if [ -f $(BINARY_NAME) ]; then \
		./$(BINARY_NAME) -version; \
	else \
		echo "Binary not built. Run 'make build' first."; \
	fi

## Run platform info
platform-info: build
	@echo "Platform capabilities:"
	./$(BINARY_NAME) -platform-info

## Development workflow
dev: clean fmt vet test build
	@echo "Development build completed"

## CI workflow
ci: deps fmt vet lint test build-all
	@echo "CI build completed"

## Help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help