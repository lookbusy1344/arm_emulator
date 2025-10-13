#!/bin/bash

echo "Building and testing with clean cache..."
go build -o arm-emulator
go clean -testcache
go test ./...
