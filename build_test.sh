#!/bin/bash
# Build and test ARM Emulator with version information

set -e

echo "Building and testing with clean cache..."
echo ""

# Get version from git tags, or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0-dev")

# Get commit hash
COMMIT=$(git log -1 --format=%H 2>/dev/null || echo "unknown")

# Get build date
DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo "Building with version info..."
echo "  Version: $VERSION"
echo "  Commit:  ${COMMIT:0:8}"
echo "  Date:    $DATE"
echo ""

# Build with ldflags to inject version info
go build -ldflags "-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT' -X 'main.Date=$DATE'" -o arm-emulator

echo "✓ Build complete"
echo ""
echo "Cleaning test cache and running tests..."
go clean -testcache
go test ./...
