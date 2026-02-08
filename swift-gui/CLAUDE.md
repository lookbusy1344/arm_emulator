# Swift GUI Development Guide

macOS-native GUI for ARM Emulator using SwiftUI and MVVM architecture.

## Platform Requirements

**IMPORTANT:** This Swift project targets modern platforms only - no backward compatibility required.

- **macOS:** 26.2
- **Swift:** 6.2
- **Xcode:** 26.2

Always use the latest SwiftUI APIs and Swift language features. Target Apple Silicon only.

## Prerequisites

```bash
brew install xcodegen swiftlint swiftformat xcbeautify
```

## Quick Reference

### Building

```bash
# Generate Xcode project (after modifying project.yml)
xcodegen generate

# Build (requires Go backend built first: cd .. && make build)
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# Clean build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build | xcbeautify
```

### Testing

```bash
# Run all tests
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# Run specific test
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -only-testing:TestClass/testMethod -destination 'platform=macOS'
```

### Code Quality (MANDATORY before commit)

```bash
# Format code
swiftformat .

# Lint code (must have 0 violations)
swiftlint

# Fix auto-correctable lint issues
swiftlint --fix

# Strict mode (warnings as errors)
swiftlint --strict
```

**CRITICAL:** SwiftLint must show **0 violations** before committing. Use `swiftformat .` first, then verify with `swiftlint`.

### Running the App

```bash
# After building, run the app
open ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*/Build/Products/Debug/ARMEmulator.app

# Or find and run
find ~/Library/Developer/Xcode/DerivedData -name "ARMEmulator.app" -type d -exec open {} \;
```

## Architecture

- **Pattern:** MVVM with SwiftUI
- **Backend Connection:** HTTP REST API + WebSocket to Go backend (port 8080)
- **Features:** Modern Swift 6.2 features (async/await, actors, structured concurrency)
- **Target:** Apple Silicon only, macOS 26.2+

**⚠️ CRITICAL: API SYNCHRONIZATION**
- Go backend API is shared by **both Swift GUI and Avalonia GUI**
- **DO NOT make breaking API changes** - only additive changes allowed
- Any API modifications must work with both frontends
- Test both GUIs after backend changes

## Complete Build & Test Pipeline

```bash
#!/bin/bash
set -e

PROJECT="ARMEmulator.xcodeproj"
SCHEME="ARMEmulator"
DESTINATION="platform=macOS"

echo "🧹 Cleaning..."
xcodebuild clean -project "$PROJECT" -scheme "$SCHEME" > /dev/null

echo "✨ Formatting code..."
swiftformat .

echo "🔍 Linting..."
swiftlint --strict

echo "🔨 Building..."
xcodebuild build -project "$PROJECT" -scheme "$SCHEME" -destination "$DESTINATION" | xcbeautify

echo "🧪 Testing..."
xcodebuild test -project "$PROJECT" -scheme "$SCHEME" -destination "$DESTINATION" | xcbeautify

echo "✅ All checks passed!"
```

## Common Pitfalls and Solutions

### 1. Scheme Not Found

**Problem:** `xcodebuild: error: The project does not contain a scheme named 'ARMEmulator'`

**Solution:**
- Schemes must be marked as "Shared" to be visible to command-line tools
- In Xcode: Product → Scheme → Manage Schemes → Check "Shared"
- Or manually edit `.xcodeproj/xcshareddata/xcschemes/`

```bash
# Verify scheme exists and is shared
ls -la ARMEmulator.xcodeproj/xcshareddata/xcschemes/
```

### 2. Derived Data Issues

**Problem:** Stale build artifacts causing mysterious errors.

**Solution:**
```bash
# Clean derived data
rm -rf ~/Library/Developer/Xcode/DerivedData

# Or use specific derived data path
xcodebuild build -derivedDataPath ./build/DerivedData -project ARMEmulator.xcodeproj -scheme ARMEmulator
```

### 3. SwiftLint/SwiftFormat Version Mismatches

**Problem:** Different tool versions causing inconsistent results.

**Solution:**
```bash
# Check versions
swiftlint version
swiftformat --version
```

### 4. Build Errors Hard to Read

**Problem:** Raw xcodebuild output is verbose and hard to parse.

**Solution:**
```bash
# Always pipe through xcbeautify for readable output
xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator 2>&1 | xcbeautify

# Save formatted output while displaying it
xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator 2>&1 | tee build.log | xcbeautify
```

## XcodeGen Usage

This project uses XcodeGen to generate the Xcode project from `project.yml`.

```bash
# After modifying project.yml, regenerate the project
xcodegen generate

# The generated ARMEmulator.xcodeproj should not be committed
# Only commit project.yml changes
```

## Pre-commit Checklist

Before every commit, run:

```bash
# 1. Format code
swiftformat .

# 2. Lint code (must pass with 0 violations)
swiftlint

# 3. Build successfully
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# 4. Tests pass
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify
```

## DEVELOPER_DIR Setup

Instead of setting `DEVELOPER_DIR` before every command:

```bash
# Check current developer directory
xcode-select -p

# If it's not /Applications/Xcode.app/Contents/Developer, set it:
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

## Fast Iteration During Development

```bash
# Quick build without cleaning
xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator -quiet | xcbeautify --simple

# Quick test of specific test
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -only-testing:ARMEmulatorTests/SpecificTest -quiet
```

## Additional Documentation

- **Full CLI Automation Guide:** `../docs/SWIFT_CLI_AUTOMATION.md`
- **MCP UI Debugging:** `../docs/MCP_UI_DEBUGGING.md`
- **Architecture Planning:** `../SWIFT_GUI_PLANNING.md`
- **Instructions & Syscalls:** `../docs/INSTRUCTIONS.md`

## Project Structure

```
swift-gui/
├── ARMEmulator/
│   ├── Models/          - Data models
│   ├── ViewModels/      - MVVM view models
│   ├── Views/           - SwiftUI views
│   ├── Services/        - API/WebSocket clients
│   └── Utilities/       - Helpers and extensions
├── ARMEmulatorTests/    - Unit tests
├── project.yml          - XcodeGen project definition
└── CLAUDE.md           - This file
```

## Development Workflow

1. **Start Go Backend:** `cd .. && ./arm-emulator` (or use HTTP API mode)
2. **Generate Project:** `xcodegen generate` (if project.yml changed)
3. **Format Code:** `swiftformat .`
4. **Build:** `xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify`
5. **Test:** `xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`
6. **Lint:** `swiftlint` (0 violations required)
7. **Run:** `open ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*/Build/Products/Debug/ARMEmulator.app`

## Notes

- **No backward compatibility** - use latest Swift/SwiftUI features freely
- **Apple Silicon only** - no Intel support required
- **Modern Swift 6.2** - leverage all new language features
- **SwiftUI-first** - avoid UIKit/AppKit unless absolutely necessary
- **0 SwiftLint violations** - non-negotiable before commit
