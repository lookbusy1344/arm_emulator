# Platform Packaging Scripts

Build scripts for creating distributable packages on each platform.

## Prerequisites

### All Platforms
- .NET SDK 10.0+
- ARM Emulator backend binary built for target platform

### macOS
- macOS 13+ (Ventura or later)
- Xcode Command Line Tools: `xcode-select --install`

### Windows
- Windows 10/11
- Windows SDK (for MakeAppx.exe and signing tools)
- PowerShell 5.1+

### Linux
- Linux with AppImage tools
- `appimagetool` - Download from https://appimage.github.io/appimagetool/

## Building Packages

### macOS DMG

```bash
cd packaging/macos
./build-dmg.sh
```

Output:
- `dist/macos/ARMEmulator.app` - Application bundle
- `dist/macos/ARMEmulator-1.0.0-osx-arm64.dmg` - DMG installer

The script automatically:
- Builds the .NET app for the current architecture
- Creates .app bundle structure
- Bundles the backend binary in Contents/Resources
- Creates Info.plist with app metadata
- Generates a DMG installer

### Windows MSIX

```powershell
cd packaging\windows
.\build-msix.ps1
```

Output:
- `dist\windows\app\` - Application files
- `dist\windows\manifest\` - MSIX manifest

To create the MSIX package:
```powershell
# Using Windows SDK MakeAppx tool
MakeAppx.exe pack /d "dist\windows\app" /p "dist\windows\ARMEmulator-1.0.0-win-x64.msix"

# Sign the package (required for installation)
SignTool.exe sign /fd SHA256 /a /f YourCertificate.pfx "dist\windows\ARMEmulator-1.0.0-win-x64.msix"
```

### Linux AppImage

```bash
cd packaging/linux
./build-appimage.sh
```

Output:
- `dist/linux/AppDir/` - AppImage directory structure

To create the AppImage:
```bash
# Download appimagetool if not installed
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage

# Create AppImage
ARCH=x86_64 ./appimagetool-x86_64.AppImage dist/linux/AppDir dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage
```

## Backend Binary Requirements

The backend binary must be built for the target platform before packaging:

### macOS
```bash
# From project root
cd ..
make build  # Builds arm-emulator for macOS
```

### Windows
```bash
# Cross-compile from macOS/Linux
GOOS=windows GOARCH=amd64 go build -o arm-emulator.exe

# Or build on Windows
go build -o arm-emulator.exe
```

### Linux
```bash
# Cross-compile from macOS
GOOS=linux GOARCH=amd64 go build -o arm-emulator

# Or build on Linux
go build -o arm-emulator
```

## Distribution

### macOS
- Distribute the DMG file
- Users drag ARMEmulator.app to Applications folder
- Gatekeeper may require "Open Anyway" on first launch (unsigned builds)

### Windows
- Distribute the MSIX file
- Users double-click to install
- Requires developer mode or certificate signing

### Linux
- Distribute the AppImage file
- Users make it executable and run: `chmod +x ARMEmulator-*.AppImage && ./ARMEmulator-*.AppImage`
- No installation required

## Code Signing

### macOS (optional, recommended for distribution)
```bash
# Sign the .app bundle
codesign --force --deep --sign "Developer ID Application: Your Name" ARMEmulator.app

# Notarize with Apple (requires Apple Developer account)
xcrun notarytool submit ARMEmulator-1.0.0-osx-arm64.dmg --apple-id your@email.com --team-id TEAMID --wait
```

### Windows (required for MSIX)
```powershell
# Create a self-signed certificate (development only)
New-SelfSignedCertificate -Type Custom -Subject "CN=ARMEmulator" -KeyUsage DigitalSignature -FriendlyName "ARMEmulator" -CertStoreLocation "Cert:\CurrentUser\My"

# Sign the MSIX
SignTool.exe sign /fd SHA256 /a /f Certificate.pfx ARMEmulator-1.0.0-win-x64.msix
```

## Testing Packages

### macOS
```bash
open dist/macos/ARMEmulator.app
# Or install from DMG
open dist/macos/ARMEmulator-1.0.0-osx-arm64.dmg
```

### Windows
```powershell
# Enable developer mode or sideload the app
Add-AppxPackage dist\windows\ARMEmulator-1.0.0-win-x64.msix
```

### Linux
```bash
chmod +x dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage
./dist/linux/ARMEmulator-1.0.0-linux-x64.AppImage
```

## Troubleshooting

### macOS: "App is damaged and can't be opened"
```bash
xattr -cr ARMEmulator.app
```

### Windows: "Windows protected your PC"
- Click "More info" → "Run anyway"
- Or properly sign the MSIX with a trusted certificate

### Linux: AppImage won't run
```bash
# Check architecture
file ARMEmulator-1.0.0-linux-x64.AppImage

# Try FUSE-less mode
./ARMEmulator-1.0.0-linux-x64.AppImage --appimage-extract-and-run
```

## CI/CD Integration

These scripts can be integrated into GitHub Actions or other CI systems:

```yaml
# Example: GitHub Actions
- name: Build macOS DMG
  run: |
    cd avalonia-gui/packaging/macos
    ./build-dmg.sh

- name: Upload DMG
  uses: actions/upload-artifact@v3
  with:
    name: macos-dmg
    path: avalonia-gui/dist/macos/*.dmg
```
