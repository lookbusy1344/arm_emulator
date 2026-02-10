#!/bin/bash
set -e

# Linux AppImage builder for ARM Emulator
# Requires: appimagetool (download from https://appimage.github.io/appimagetool/)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$PROJECT_DIR/dist/linux"
APP_NAME="ARMEmulator"
VERSION="1.0.0"
RID="linux-x64"

echo "Building ARM Emulator for Linux..."
echo "Project directory: $PROJECT_DIR"
echo "Output directory: $OUTPUT_DIR"

# Clean previous builds
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build the .NET application
echo "Publishing .NET application..."
cd "$PROJECT_DIR"
dotnet publish ARMEmulator/ARMEmulator.csproj \
    --configuration Release \
    --runtime "$RID" \
    --self-contained true \
    -p:PublishSingleFile=false \
    --output "$OUTPUT_DIR/AppDir/usr/bin"

# Copy backend binary
if [ -f "$PROJECT_DIR/../arm-emulator" ]; then
    echo "Copying backend binary..."
    cp "$PROJECT_DIR/../arm-emulator" "$OUTPUT_DIR/AppDir/usr/bin/"
    chmod +x "$OUTPUT_DIR/AppDir/usr/bin/arm-emulator"
else
    echo "WARNING: Backend binary not found at $PROJECT_DIR/../arm-emulator"
    echo "You'll need to build it for Linux and place it in the AppDir manually."
fi

# Create AppImage directory structure
mkdir -p "$OUTPUT_DIR/AppDir/usr/share/applications"
mkdir -p "$OUTPUT_DIR/AppDir/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$OUTPUT_DIR/AppDir/usr/share/metainfo"

# Create .desktop file
cat > "$OUTPUT_DIR/AppDir/usr/share/applications/$APP_NAME.desktop" << EOF
[Desktop Entry]
Name=ARM Emulator
Exec=ARMEmulator
Icon=armemulator
Type=Application
Categories=Development;Education;
Comment=ARM Assembly Language Emulator
Terminal=false
StartupWMClass=ARMEmulator
EOF

# Create AppRun script
cat > "$OUTPUT_DIR/AppDir/AppRun" << 'EOF'
#!/bin/bash
SELF_DIR="$(dirname "$(readlink -f "$0")")"
export LD_LIBRARY_PATH="$SELF_DIR/usr/bin:$LD_LIBRARY_PATH"
exec "$SELF_DIR/usr/bin/ARMEmulator" "$@"
EOF

chmod +x "$OUTPUT_DIR/AppDir/AppRun"

# Copy desktop file to root of AppDir
cp "$OUTPUT_DIR/AppDir/usr/share/applications/$APP_NAME.desktop" "$OUTPUT_DIR/AppDir/"

echo "AppImage structure created successfully"
echo ""
echo "Build complete!"
echo "AppDir: $OUTPUT_DIR/AppDir"
echo ""
echo "To create an AppImage, download appimagetool and run:"
echo "  appimagetool '$OUTPUT_DIR/AppDir' '$OUTPUT_DIR/ARMEmulator-$VERSION-$RID.AppImage'"
echo ""
echo "Or use this command if appimagetool is in PATH:"
echo "  ARCH=x86_64 appimagetool '$OUTPUT_DIR/AppDir' '$OUTPUT_DIR/ARMEmulator-$VERSION-$RID.AppImage'"
