.PHONY: all build test clean dev install help

all: build

install:
	@echo "Installing dependencies..."
	cd frontend && npm install
	go mod download

frontend:
	@echo "Building frontend..."
	cd frontend && npm run build

build: frontend
	@echo "Building application..."
	go build -o ../build/arm-emulator-gui

dev:
	@echo "Starting development server..."
	wails dev

test: frontend
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	cd frontend && npm test -- --run

test-watch:
	cd frontend && npm test

clean:
	@echo "Cleaning build artifacts..."
	rm -rf frontend/dist
	rm -rf ../build/arm-emulator-gui

help:
	@echo "Available targets:"
	@echo "  install     - Install all dependencies"
	@echo "  build       - Build production binary"
	@echo "  dev         - Start development server"
	@echo "  test        - Run all tests"
	@echo "  test-watch  - Run frontend tests in watch mode"
	@echo "  clean       - Remove build artifacts"
