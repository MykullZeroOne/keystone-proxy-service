#!/bin/bash

PLIST_NAME="com.revfcu.keystone-proxy-service"
PLIST_PATH="$HOME/Library/LaunchAgents/$PLIST_NAME.plist"
INSTALL_DIR="$HOME/.keystone-proxy-service"

launchctl unload "$PLIST_PATH" 2>/dev/null
rm -f "$PLIST_PATH"
rm -rf "$INSTALL_DIR"

echo "Keystone Proxy Service uninstalled."
