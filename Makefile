# ARM Emulator Makefile

# Binary name
BINARY=arm-emulator

# Version from git tag (fallback to "dev" if no tag)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all build clean test fmt lint run install help

all: build

## build: Build the binary with version info
build:
	@echo "Building $(BINARY) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY)

## clean: Remove build artifacts and test cache
clean:
	@echo "Cleaning..."
	go clean
	go clean -testcache
	rm -f $(BINARY)

## test: Run all tests
test:
	@echo "Running tests..."
	go clean -testcache
	go test ./...

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	go fmt ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## run: Build and run with example program
run: build
	./$(BINARY) examples/hello.s

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY) $(VERSION)..."
	go install $(LDFLAGS)

## version: Show version info that would be embedded
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

## help: Show this help message
help:
	@echo "ARM Emulator Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
