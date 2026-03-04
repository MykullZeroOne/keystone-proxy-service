#!/bin/bash
set -e

VERSION="${1:-1.0.0}"
BUILD_DIR="build"

echo "=== Building Keystone Proxy Service v${VERSION} ==="

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build universal macOS binary
echo "Building arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "$BUILD_DIR/keystone-proxy-service-arm64" .

echo "Building amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "$BUILD_DIR/keystone-proxy-service-amd64" .

echo "Creating universal binary..."
lipo -create \
  -output "$BUILD_DIR/keystone-proxy-service" \
  "$BUILD_DIR/keystone-proxy-service-arm64" \
  "$BUILD_DIR/keystone-proxy-service-amd64"

# Create pkg payload
echo "Packaging .pkg..."
mkdir -p "$BUILD_DIR/pkg-root/usr/local/bin"
cp "$BUILD_DIR/keystone-proxy-service" "$BUILD_DIR/pkg-root/usr/local/bin/"
chmod +x "$BUILD_DIR/pkg-root/usr/local/bin/keystone-proxy-service"

pkgbuild \
  --root "$BUILD_DIR/pkg-root" \
  --identifier "com.revfcu.keystone-proxy-service" \
  --version "$VERSION" \
  --scripts "scripts/pkg" \
  --install-location "/" \
  "$BUILD_DIR/keystone-proxy-service-${VERSION}.pkg"

# Create DMG containing the pkg
echo "Creating .dmg..."
DMG_DIR="$BUILD_DIR/dmg-contents"
mkdir -p "$DMG_DIR"
cp "$BUILD_DIR/keystone-proxy-service-${VERSION}.pkg" "$DMG_DIR/"

# Add a readme visible when the DMG is opened
cat > "$DMG_DIR/READ ME FIRST.txt" << 'EOF'
Keystone Proxy Service
======================

Double-click the .pkg file to install.

The service will start automatically and run in the background.

FIRST TIME SETUP:
  1. Open https://127.0.0.1:51763 in your browser
  2. Accept the self-signed certificate (one time only)
  3. Navigate to your Keystone URL
     (e.g. https://keystonedev.revfcu.com:8443/Development/)

To uninstall, run in Terminal:
  sudo pkgutil --forget com.revfcu.keystone-proxy-service
  launchctl unload ~/Library/LaunchAgents/com.revfcu.keystone-proxy-service.plist
  rm ~/Library/LaunchAgents/com.revfcu.keystone-proxy-service.plist
  sudo rm /usr/local/bin/keystone-proxy-service
  rm -rf ~/.keystone-proxy-service
EOF

hdiutil create \
  -volname "Keystone Proxy Service ${VERSION}" \
  -srcfolder "$DMG_DIR" \
  -ov \
  -format UDZO \
  "$BUILD_DIR/keystone-proxy-service-${VERSION}.dmg"

# Clean up intermediates
rm -rf "$BUILD_DIR/pkg-root" "$DMG_DIR"
rm -f "$BUILD_DIR/keystone-proxy-service-arm64" "$BUILD_DIR/keystone-proxy-service-amd64"

echo ""
echo "=== Build complete ==="
echo "  PKG: $BUILD_DIR/keystone-proxy-service-${VERSION}.pkg"
echo "  DMG: $BUILD_DIR/keystone-proxy-service-${VERSION}.dmg"
echo "  BIN: $BUILD_DIR/keystone-proxy-service (universal)"
