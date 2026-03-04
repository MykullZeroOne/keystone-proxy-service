# Keystone Proxy Service

A lightweight macOS background service that emulates the Corelation device identification service (normally a Windows-only background process). This allows Mac users to log in to Keystone via the real web app using Kerberos/AD authentication.

**Single binary, zero dependencies.**

## How it works

Keystone's web app calls `https://127.0.0.1:51763/GetDeviceInformation` to identify the local machine. On Windows, this is handled by a background service. This binary provides that same endpoint on Mac.

## Quick start

### Build

    go build -o keystone-proxy-service

### Install as a background service (auto-starts on login)

    ./install.sh

### First-time browser setup

1. Open **https://127.0.0.1:51763** in your browser
2. Accept the self-signed certificate (one time only)
3. Navigate to your Keystone URL (e.g. `https://keystonedev.revfcu.com:8443/Development/`)
4. If your device isn't registered, check the **"Insert New Device"** checkbox when prompted

### Uninstall

    ./uninstall.sh

## Distributing to other Mac users

Give them two files:
- `keystone-proxy-service` (the binary)
- `install.sh`

They run `./install.sh` — that's it. No Go, Node, or any other runtime needed.

### Cross-compile for Apple Silicon and Intel

    GOOS=darwin GOARCH=arm64 go build -o keystone-proxy-service-arm64
    GOOS=darwin GOARCH=amd64 go build -o keystone-proxy-service-amd64

Or build a universal binary:

    GOOS=darwin GOARCH=arm64 go build -o keystone-proxy-service-arm64
    GOOS=darwin GOARCH=amd64 go build -o keystone-proxy-service-amd64
    lipo -create -output keystone-proxy-service keystone-proxy-service-arm64 keystone-proxy-service-amd64

## Logs

    cat ~/.keystone-proxy-service/service.log
