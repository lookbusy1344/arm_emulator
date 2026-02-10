# Phase 10: Platform Integration - Testing Guide

This document outlines how to test all Phase 10 platform integration features.

## Completed Features

### 1. Backend Binary Bundling ✅

**What it does:**
- Automatically copies the Go backend binary to the output directory during build
- Platform-specific bundling for macOS (.app/Contents/Resources), Windows (alongside exe), and Linux

**How to test:**

```bash
# Build the project
dotnet build

# Verify backend binary is in output directory
ls -la ARMEmulator/bin/Debug/net10.0/arm-emulator  # macOS/Linux
dir ARMEmulator\bin\Debug\net10.0\arm-emulator.exe  # Windows

# Run the app - backend should auto-start
dotnet run --project ARMEmulator
```

**Expected result:** App launches and successfully connects to backend on http://localhost:8080

### 2. Platform Theme Detection ✅

**What it does:**
- Detects system theme (light/dark mode) on macOS, Windows, and Linux
- Provides reactive updates when system theme changes

**How to test:**

```bash
# Run tests
dotnet test --filter "FullyQualifiedName~PlatformThemeDetectorTests"

# Manual test on macOS
defaults read -g AppleInterfaceStyle  # Should print "Dark" or return error (light mode)

# In app: Change system theme and verify app updates
# macOS: System Settings > Appearance > Light/Dark
# Windows: Settings > Personalization > Colors > Dark/Light
# Linux: Depends on desktop environment
```

**Expected result:** All tests pass, app theme changes when system theme changes (if AppSettings.Theme is set to Auto)

### 3. Theme Service ✅

**What it does:**
- Integrates platform theme detection with Avalonia's theme system
- Supports Light, Dark, and Auto modes
- Reactive theme switching

**How to test:**

```bash
# Run tests
dotnet test --filter "FullyQualifiedName~ThemeServiceTests"

# Manual test: Launch app and open Preferences
# Set theme to Auto - should match system theme
# Set theme to Light - should use light theme
# Set theme to Dark - should use dark theme
```

**Expected result:** All tests pass, theme changes immediately when preference is changed

### 4. macOS Packaging ✅

**What it does:**
- Creates .app bundle with proper structure
- Bundles backend binary in Contents/Resources
- Generates DMG installer

**How to test:**

```bash
cd packaging/macos
./build-dmg.sh

# Verify .app structure
ls -la ../../dist/macos/ARMEmulator.app/Contents/Resources/arm-emulator
ls -la ../../dist/macos/ARMEmulator.app/Contents/Info.plist

# Test .app bundle
open ../../dist/macos/ARMEmulator.app

# Test DMG
open ../../dist/macos/ARMEmulator-1.0.0-osx-arm64.dmg
```

**Expected results:**
- .app bundle launches successfully
- Backend binary is present and executable
- Info.plist contains correct bundle metadata
- DMG mounts and shows ARMEmulator.app

**Troubleshooting:**
- If you get "App is damaged", run: `xattr -cr ARMEmulator.app`
- If backend doesn't start, check: `ls -la Contents/Resources/arm-emulator`

### 5. Windows Packaging ✅

**What it does:**
- Creates single-file executable with bundled dependencies
- Prepares MSIX manifest for Windows Store distribution

**How to test (requires Windows):**

```powershell
cd packaging\windows
.\build-msix.ps1

# Verify structure
dir ..\..\dist\windows\app\ARMEmulator.exe
dir ..\..\dist\windows\app\arm-emulator.exe

# Test executable
..\..\dist\windows\app\ARMEmulator.exe
```

**Expected results:**
- Single-file exe is created
- Backend binary is bundled
- App launches and connects to backend

**Creating MSIX (requires Windows SDK):**

```powershell
# Requires Windows SDK installed
MakeAppx.exe pack /d "dist\windows\app" /p "dist\windows\ARMEmulator-1.0.0-win-x64.msix"
```

### 6. Linux Packaging ✅

**What it does:**
- Creates AppImage directory structure
- Includes .desktop file and launcher

**How to test (requires Linux):**

```bash
cd packaging/linux
./build-appimage.sh

# Verify structure
ls -la ../../dist/linux/AppDir/usr/bin/ARMEmulator
ls -la ../../dist/linux/AppDir/usr/bin/arm-emulator
cat ../../dist/linux/AppDir/ARMEmulator.desktop

# Create AppImage (requires appimagetool)
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage
ARCH=x86_64 ./appimagetool-x86_64.AppImage ../../dist/linux/AppDir ../../dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage

# Test AppImage
chmod +x ../../dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage
../../dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage
```

**Expected results:**
- AppDir structure is correct
- AppImage launches successfully
- Backend starts automatically

## Integration Tests

### Test Backend Auto-Start

```bash
# Kill any running backend
pkill arm-emulator

# Launch app
dotnet run --project ARMEmulator

# Verify backend is running
ps aux | grep arm-emulator
curl http://localhost:8080/health
```

**Expected:** Backend process starts automatically, health check returns 200 OK

### Test Backend Health Check

```bash
# Build and run
dotnet run --project ARMEmulator

# Check logs for health check
# Should see "Backend status: Running" or similar

# Manually kill backend while app is running
pkill arm-emulator

# App should detect backend is down and show error state
```

**Expected:** App detects backend failure and updates UI state

### Test Cross-Platform File Dialogs

File dialogs use Avalonia's built-in cross-platform dialogs which work on all platforms without additional configuration.

```bash
# Run app
dotnet run --project ARMEmulator

# Test File > Open (Ctrl+O)
# Test File > Save (Ctrl+S)
# Test File > Save As (Ctrl+Shift+S)
# Test File > Open Example (Ctrl+Shift+E)
```

**Expected:** Native-looking file dialogs appear on each platform

## Platform-Specific Features

### macOS

**Menu Bar Integration:**
- macOS automatically uses native menu bar
- Avalonia handles this without special configuration

**System Theme:**
```bash
# Test auto theme switching
# 1. Set app theme to Auto
# 2. Change system theme: System Settings > Appearance
# 3. App theme should update immediately
```

### Windows

**Task Bar:**
- Windows automatically shows app in taskbar
- Icon and title from app manifest

**Theme:**
```powershell
# Test auto theme switching
# 1. Set app theme to Auto
# 2. Settings > Personalization > Colors > Choose your mode
# 3. App should update (may require app restart)
```

### Linux

**Desktop Integration:**
```bash
# .desktop file is in AppDir
cat dist/linux/AppDir/ARMEmulator.desktop

# Test AppImage integration
./ARMEmulator-1.0.0-linux-x64.AppImage
```

## Regression Tests

After Phase 10 changes, verify previous phases still work:

```bash
# Run all tests
dotnet test

# Verify UI still works
dotnet run --project ARMEmulator

# Test key features:
# - Load program (examples)
# - Run/Step/Pause
# - View registers, memory, stack
# - Set breakpoints
# - Expression evaluator
```

**Expected:** All tests pass (100% of previous functionality intact)

## Performance Tests

### Startup Time

```bash
# Measure app startup time
time dotnet run --project ARMEmulator --no-build
```

**Expected:** App launches in < 3 seconds on modern hardware

### Backend Startup

```bash
# Measure backend startup within app
# Check logs for "Backend started" timestamp
```

**Expected:** Backend starts in < 1 second

### Memory Usage

```bash
# Monitor memory usage
dotnet run --project ARMEmulator &
sleep 5
ps aux | grep ARMEmulator
```

**Expected:** Reasonable memory usage (< 200MB for GUI + backend)

## Documentation

All packaging documentation is in `packaging/README.md`:
- Build instructions for each platform
- Code signing guidance
- Distribution instructions
- Troubleshooting tips

## Known Issues

None currently identified.

## Future Enhancements

Potential future improvements:
- Code signing for macOS (requires Apple Developer account)
- Windows Store submission (requires Publisher ID)
- Linux Flatpak packaging (alternative to AppImage)
- Auto-update mechanism
- Crash reporting integration

## Summary

Phase 10 Platform Integration is complete with the following delivered:

✅ Backend binary auto-bundling (all platforms)
✅ Platform theme detection (macOS, Windows, Linux)
✅ Theme service with Auto/Light/Dark modes
✅ macOS .app bundle and DMG packaging
✅ Windows single-file exe and MSIX manifest
✅ Linux AppImage structure
✅ Cross-platform file dialogs (Avalonia built-in)
✅ Comprehensive documentation
✅ Test coverage for all new features

All features are production-ready and tested.
