#!/usr/bin/env bash
# Install platform-specific build dependencies.
# Follows Wails v2 platform guides.
# Usage: ./install-deps.sh <goos>
set -euo pipefail

GOOS="$1"

case "$GOOS" in
  linux)
    echo "Installing Linux build dependencies..."
    sudo apt-get update
    sudo apt-get install -y \
      libgtk-3-dev \
      libwebkit2gtk-4.1-dev \
      pkg-config \
      gstreamer1.0-plugins-good
    ;;
  windows)
    echo "Installing Windows build dependencies..."
    choco install nsis -y
    ;;
  darwin)
    echo "macOS: no additional system deps needed for build."
    echo "Note: For App Store submission, install Xcode CLI tools and signing certificates."
    ;;
  *)
    echo "Unknown platform: $GOOS"
    exit 1
    ;;
esac

echo "Dependencies installed for $GOOS."
