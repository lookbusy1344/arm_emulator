# Version API Implementation Summary

## Overview
Added a `/api/v1/version` endpoint to the backend and integrated it into the Swift GUI's About dialog.

## Backend Changes

### 1. API Models (`api/models.go`)
- Added `VersionResponse` struct with `version`, `commit`, and `date` fields

### 2. API Handlers (`api/handlers.go`)
- Added `handleGetVersion()` handler for `GET /api/v1/version`
- Returns JSON with version information from the server

### 3. Server (`api/server.go`)
- Updated `Server` struct to include `version`, `commit`, `date` fields
- Added `NewServerWithVersion()` constructor
- Maintained backwards compatibility with `NewServer()` (returns "dev", "unknown", "unknown")
- Registered `/api/v1/version` route

### 4. Main Entry Point (`main.go`)
- Updated to pass `Version`, `Commit`, `Date` variables to API server
- These variables can be set at build time using `-ldflags`

### 5. Build System
- **Makefile**: Already configured with version injection via ldflags
  ```bash
  make build  # Builds with git version info
  make version  # Shows version info that will be embedded
  ```
- **Build Script**: Created `build_with_version.sh` for convenience
  - Automatically extracts version from git tags
  - Injects commit hash and build date
  - Verifies build by calling `-version` flag
- **Test Script**: Updated `build_test.sh` to build with version info before running tests

### 6. Unit Tests (`tests/unit/api/version_test.go`)
Comprehensive test coverage including:
- ✅ Version endpoint with different version formats
- ✅ Method validation (405 for POST/PUT/DELETE/PATCH)
- ✅ CORS headers for localhost/127.0.0.1/file://
- ✅ JSON format validation
- ✅ Backwards compatibility with `NewServer()`
- **Total: 5 test functions with 13 sub-tests, all passing**

## Frontend Changes (Swift)

### 1. API Client (`APIClient.swift`)
- Added `getVersion()` method to fetch backend version
- Added `BackendVersion` struct matching API response

### 2. About View (`Views/AboutView.swift`)
- New SwiftUI view displaying:
  - App icon and name
  - Backend version number
  - Commit hash (truncated to 8 chars)
  - Build date
- Handles loading states and errors gracefully
- Uses `DebugLog` for error tracking

### 3. App Menu (`ARMEmulatorApp.swift`)
- Added "About ARM Emulator" menu item (replaces default About)
- Shows AboutView in a modal sheet

## API Endpoint

### Request
```bash
GET /api/v1/version
```

### Response
```json
{
  "version": "v1.1.2-123-g1e713a3-dirty",
  "commit": "1e713a3006ca790974eb44d22691a192f2ab98c1",
  "date": "2026-01-07T09:34:45Z"
}
```

## Build Instructions

### Backend (with version info)
```bash
# Option 1: Use Makefile (recommended)
make build

# Option 2: Use build script
./build_with_version.sh

# Option 3: Manual with git info
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git log -1 --format=%H)
DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
go build -ldflags "-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT' -X 'main.Date=$DATE'" -o arm-emulator
```

### Swift App
```bash
cd swift-gui
xcodegen generate  # If project.yml changed
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build
```

The app automatically copies the backend binary during build. Ensure the backend is built first.

## Testing

### Backend Tests
```bash
# Run all API unit tests
go test -v ./tests/unit/api/...

# Run specific version tests
go test -v ./tests/unit/api/version_test.go
```

### Swift Tests
```bash
cd swift-gui
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS'
```

### Manual Testing
1. Build backend: `make build`
2. Start API server: `./arm-emulator -api-server`
3. Test endpoint: `curl http://localhost:8080/api/v1/version`
4. Or open Swift app and click "About ARM Emulator" in app menu

## Documentation Updates

- Updated `CLAUDE.md` with build instructions emphasizing version injection
- Added warning to Swift app build section about building backend first

## Verification

All tests passing:
- ✅ Backend: 5 new version API tests (13 sub-tests)
- ✅ All existing API tests continue to pass
- ✅ Swift: All 12 tests pass
- ✅ Linting: 0 violations (Go and Swift)
- ✅ Manual testing: Version displays correctly in About dialog

## Future Enhancements

Potential improvements (not implemented):
- Cache version info in Swift app to avoid repeated API calls
- Add GUI version/build info alongside backend version
- Display more detailed build information (Go version, OS, architecture)
- Add version check on startup with warning if outdated
