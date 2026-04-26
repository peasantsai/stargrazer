#!/usr/bin/env bash
# Copy Chromium bundle and cookies extension into the build output.
# Usage: ./bundle-assets.sh <build_bin_dir>
set -euo pipefail

BUILD_DIR="$1"

echo "Bundling assets into $BUILD_DIR..."

mkdir -p "$BUILD_DIR/assets/chromium-bundle"
cp -R assets/chromium-bundle/* "$BUILD_DIR/assets/chromium-bundle/" 2>/dev/null || true
cp -R assets/cookies-extension "$BUILD_DIR/assets/cookies-extension" 2>/dev/null || true

echo "Assets bundled."
