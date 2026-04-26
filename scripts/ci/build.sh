#!/usr/bin/env bash
# Build the Wails application for a target platform.
# Usage: ./build.sh <goos> <platform> [version]
set -euo pipefail

GOOS="$1"
PLATFORM="$2"
VERSION="${3:-}"

TAGS=""
if [[ "$GOOS" == "linux" ]]; then
  TAGS="-tags webkit2_41"
fi

echo "Building stargrazer for $PLATFORM..."

if [[ "$GOOS" == "windows" ]]; then
  wails build -platform "$PLATFORM" -nsis -o stargrazer.exe $TAGS
else
  wails build -platform "$PLATFORM" -o stargrazer $TAGS
fi

echo "Build complete."
