#!/usr/bin/env bash
# Download and extract Chromium for the target platform.
# Usage: ./download-chromium.sh <goos> <url> <dest_dir>
set -euo pipefail

GOOS="$1"
URL="$2"
DEST="$3"

mkdir -p "$DEST"
FILENAME="/tmp/chromium-download"
echo "Downloading Chromium for $GOOS..."
curl -L -o "$FILENAME" "$URL"

case "$GOOS" in
  linux)
    mkdir -p /tmp/chromium-extract
    tar xf "$FILENAME" -C /tmp/chromium-extract 2>/dev/null || tar xJf "$FILENAME" -C /tmp/chromium-extract 2>/dev/null
    CHROME_DIR=$(find /tmp/chromium-extract -name "chrome" -type f -exec dirname {} \; | head -1)
    if [ -n "$CHROME_DIR" ]; then
      cp -R "$CHROME_DIR"/* "$DEST/"
    else
      cp -R /tmp/chromium-extract/*/* "$DEST/" 2>/dev/null || cp -R /tmp/chromium-extract/* "$DEST/"
    fi
    ;;
  darwin)
    hdiutil attach "$FILENAME" -mountpoint /tmp/chromium-dmg -nobrowse
    cp -R /tmp/chromium-dmg/Chromium.app "$DEST/"
    hdiutil detach /tmp/chromium-dmg
    ;;
  windows)
    mkdir -p /tmp/chromium-extract
    7z x "$FILENAME" -o"/tmp/chromium-extract" -y || true
    # Installer extracts to chrome/Chrome-bin/
    CHROME_BIN=$(find /tmp/chromium-extract -type d -name "Chrome-bin" | head -1)
    if [ -n "$CHROME_BIN" ]; then
      cp -R "$CHROME_BIN"/* "$DEST/"
    else
      cp -R /tmp/chromium-extract/* "$DEST/" 2>/dev/null || true
    fi
    ;;
  *)
    echo "Unknown platform: $GOOS"
    exit 1
    ;;
esac

echo "Chromium extracted to $DEST"
