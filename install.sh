#!/bin/bash
set -e

INSTALL_DIR="$HOME/.keystone-proxy-service"
PLIST_NAME="com.revfcu.keystone-proxy-service"
PLIST_PATH="$HOME/Library/LaunchAgents/$PLIST_NAME.plist"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$SCRIPT_DIR/keystone-proxy-service"

if [ ! -f "$BINARY" ]; then
  echo "Error: binary not found at $BINARY"
  echo "Build it first: go build -o keystone-proxy-service"
  exit 1
fi

# Unload existing service if present
launchctl unload "$PLIST_PATH" 2>/dev/null || true

# Install binary
mkdir -p "$INSTALL_DIR"
cp "$BINARY" "$INSTALL_DIR/keystone-proxy-service"
chmod +x "$INSTALL_DIR/keystone-proxy-service"

# Create LaunchAgent plist
mkdir -p "$HOME/Library/LaunchAgents"
cat > "$PLIST_PATH" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$PLIST_NAME</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/keystone-proxy-service</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$INSTALL_DIR/service.log</string>
    <key>StandardErrorPath</key>
    <string>$INSTALL_DIR/service.log</string>
</dict>
</plist>
EOF

# Load and start the service
launchctl load "$PLIST_PATH"

echo ""
echo "=== Keystone Proxy Service Installed ==="
echo ""
echo "The service is now running and will auto-start on login."
echo ""
echo "FIRST TIME SETUP:"
echo "  1. Open https://127.0.0.1:51763 in your browser"
echo "  2. Accept the self-signed certificate"
echo "  3. Navigate to your Keystone URL (e.g. https://keystonedev.revfcu.com:8443/Development/)"
echo ""
echo "Logs: $INSTALL_DIR/service.log"
echo "Uninstall: run uninstall.sh"
