#!/bin/bash
# Build ARM Emulator backend with version information

set -e

# Get version from git tags, or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0-dev")

# Get commit hash
COMMIT=$(git log -1 --format=%H 2>/dev/null || echo "unknown")

# Get build date
DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo "Building ARM Emulator backend..."
echo "  Version: $VERSION"
echo "  Commit:  ${COMMIT:0:8}"
echo "  Date:    $DATE"
echo ""

# Build with ldflags to inject version info
go build -ldflags "-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT' -X 'main.Date=$DATE'" -o arm-emulator

echo ""
echo "✓ Build complete: ./arm-emulator"
echo ""
echo "Verify version:"
./arm-emulator -version
