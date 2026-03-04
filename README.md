# Keystone Proxy Service

A lightweight macOS background service that emulates the Corelation device identification service (normally a Windows-only background process). This allows Mac users to log in to Keystone via the real web app using Kerberos/AD authentication.

**Single binary, zero dependencies.**

## Install

### Option 1: Download the .pkg (recommended)

1. Download the latest `.pkg` from [Releases](../../releases/latest)
2. Double-click to install
3. The service starts automatically and will auto-start on every login

### Option 2: Download the .dmg

1. Download the latest `.dmg` from [Releases](../../releases/latest)
2. Open the DMG and double-click the `.pkg` inside

## First-time browser setup

After installing, you need to trust the self-signed certificate once:

1. Open **https://127.0.0.1:51763** in your browser
2. Accept the self-signed certificate
3. Navigate to your Keystone URL (e.g. `https://keystonedev.revfcu.com:8443/Development/`)
4. If your device isn't registered, check the **"Insert New Device"** checkbox when prompted

## How it works

Keystone's web app calls `https://127.0.0.1:51763/GetDeviceInformation` to identify the local machine. On Windows, this is handled by a Corelation background service. This binary provides that same endpoint on Mac.

## Logs

    cat ~/.keystone-proxy-service/service.log

## Uninstall

    launchctl unload ~/Library/LaunchAgents/com.revfcu.keystone-proxy-service.plist
    rm ~/Library/LaunchAgents/com.revfcu.keystone-proxy-service.plist
    sudo rm /usr/local/bin/keystone-proxy-service
    rm -rf ~/.keystone-proxy-service
    sudo pkgutil --forget com.revfcu.keystone-proxy-service

## Development

### Build locally

    go build -o keystone-proxy-service

### Build release artifacts (.pkg + .dmg with universal binary)

    ./scripts/build-release.sh 1.0.0

### Create a release

Push a version tag to trigger the GitHub Actions build:

    git tag v1.0.0
    git push origin v1.0.0

The workflow builds a universal binary (Apple Silicon + Intel), packages it as a `.pkg` and `.dmg`, and attaches them to a GitHub Release.
