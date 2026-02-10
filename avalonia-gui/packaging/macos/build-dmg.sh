#!/bin/bash
set -e

# macOS .app bundle and DMG builder for ARM Emulator
# This script builds the .NET app, creates a .app bundle, and packages it into a DMG

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$PROJECT_DIR/dist/macos"
APP_NAME="ARMEmulator"
VERSION="1.0.0"

echo "Building ARM Emulator for macOS..."
echo "Project directory: $PROJECT_DIR"
echo "Output directory: $OUTPUT_DIR"

# Clean previous builds
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Determine architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    RID="osx-arm64"
else
    RID="osx-x64"
fi

echo "Building for $RID..."

# Build the .NET application
cd "$PROJECT_DIR"
dotnet publish ARMEmulator/ARMEmulator.csproj \
    --configuration Release \
    --runtime "$RID" \
    --self-contained true \
    -p:PublishSingleFile=false \
    -p:IncludeNativeLibrariesForSelfExtract=true \
    --output "$OUTPUT_DIR/$APP_NAME.app/Contents/MacOS"

# Create .app bundle structure
mkdir -p "$OUTPUT_DIR/$APP_NAME.app/Contents/Resources"
mkdir -p "$OUTPUT_DIR/$APP_NAME.app/Contents/MacOS"

# Copy backend binary to Resources
if [ -f "$PROJECT_DIR/../arm-emulator" ]; then
    echo "Copying backend binary to Resources..."
    cp "$PROJECT_DIR/../arm-emulator" "$OUTPUT_DIR/$APP_NAME.app/Contents/Resources/"
    chmod +x "$OUTPUT_DIR/$APP_NAME.app/Contents/Resources/arm-emulator"
else
    echo "WARNING: Backend binary not found at $PROJECT_DIR/../arm-emulator"
    echo "The app will not be able to start the backend automatically."
fi

# Create Info.plist
cat > "$OUTPUT_DIR/$APP_NAME.app/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleDisplayName</key>
    <string>ARM Emulator</string>
    <key>CFBundleExecutable</key>
    <string>ARMEmulator</string>
    <key>CFBundleIdentifier</key>
    <string>com.armemulator.gui</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>ARMEmulator</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>LSMinimumSystemVersion</key>
    <string>13.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
</dict>
</plist>
EOF

echo ".app bundle created successfully"

# Create DMG
DMG_NAME="ARMEmulator-$VERSION-$RID.dmg"
echo "Creating DMG: $DMG_NAME"

# Remove existing DMG if present
rm -f "$OUTPUT_DIR/$DMG_NAME"

# Create temporary DMG mount point
hdiutil create -volname "$APP_NAME" \
    -srcfolder "$OUTPUT_DIR/$APP_NAME.app" \
    -ov -format UDZO \
    "$OUTPUT_DIR/$DMG_NAME"

echo "DMG created successfully: $OUTPUT_DIR/$DMG_NAME"
echo ""
echo "Build complete!"
echo "Application bundle: $OUTPUT_DIR/$APP_NAME.app"
echo "DMG installer: $OUTPUT_DIR/$DMG_NAME"
echo ""
echo "To test the app:"
echo "  open '$OUTPUT_DIR/$APP_NAME.app'"
echo ""
echo "To install from DMG:"
echo "  open '$OUTPUT_DIR/$DMG_NAME'"
