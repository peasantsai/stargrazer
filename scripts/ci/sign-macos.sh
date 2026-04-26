#!/usr/bin/env bash
# Sign a macOS .app bundle for App Store or notarization.
# Requires: Apple Developer certificates in Keychain, provisioning profile.
# Usage: ./sign-macos.sh <app_path> <app_cert_name> <installer_cert_name> [provisioning_profile]
#
# References:
#   https://wails.io/docs/guides/mac-appstore
set -euo pipefail

APP_PATH="$1"
APP_CERT="$2"
INSTALLER_CERT="$3"
PROFILE="${4:-}"

APP_NAME=$(basename "$APP_PATH" .app)
ENTITLEMENTS="build/darwin/entitlements.plist"

echo "=== macOS Code Signing ==="
echo "App: $APP_PATH"
echo "App cert: $APP_CERT"
echo "Installer cert: $INSTALLER_CERT"

# Embed provisioning profile if provided
if [ -n "$PROFILE" ] && [ -f "$PROFILE" ]; then
  echo "Embedding provisioning profile..."
  cp "$PROFILE" "$APP_PATH/Contents/embedded.provisionprofile"
fi

# Check entitlements file exists
if [ ! -f "$ENTITLEMENTS" ]; then
  echo "Warning: $ENTITLEMENTS not found. Creating minimal entitlements..."
  mkdir -p build/darwin
  cat > "$ENTITLEMENTS" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
</dict>
</plist>
PLIST
fi

# Sign the application
echo "Signing application..."
codesign --timestamp --options=runtime \
  -s "$APP_CERT" \
  -v --entitlements "$ENTITLEMENTS" \
  "$APP_PATH"

# Build the signed installer package
echo "Building installer package..."
productbuild --sign "$INSTALLER_CERT" \
  --component "$APP_PATH" /Applications \
  "./${APP_NAME}.pkg"

echo "=== Signing complete ==="
echo "Installer: ./${APP_NAME}.pkg"
