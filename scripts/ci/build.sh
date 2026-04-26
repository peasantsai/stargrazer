#!/usr/bin/env bash
# Build the Wails application for a target platform.
# Follows Wails v2 platform guides:
#   - Windows: -webview2 embed, -nsis for installer
#   - Linux: -tags webkit2_41 for Ubuntu 24.04+
#   - macOS: -clean for universal/fresh builds
#   - All: -obfuscated with garble for release builds
#
# Usage: ./build.sh <goos> <platform> [version] [--obfuscated]
set -euo pipefail

GOOS="$1"
PLATFORM="$2"
VERSION="${3:-}"
OBFUSCATED="${4:-}"

FLAGS="-clean"
TAGS=""
LDFLAGS=""

# Platform-specific flags
case "$GOOS" in
  linux)
    TAGS="-tags webkit2_41"
    ;;
  windows)
    # Embed WebView2 bootstrapper so users on Win10 get auto-install prompt
    FLAGS="$FLAGS -webview2 embed -nsis"
    ;;
  darwin)
    # macOS universal binary if building for darwin/universal
    ;;
esac

# Version injection via ldflags
if [ -n "$VERSION" ] && [ "$VERSION" != "dry-run" ]; then
  LDFLAGS="-ldflags \"-X main.version=$VERSION\""
fi

# Obfuscation (garble) for release builds
if [ "$OBFUSCATED" = "--obfuscated" ]; then
  FLAGS="$FLAGS -obfuscated"
fi

# Output filename
if [[ "$GOOS" == "windows" ]]; then
  OUTPUT="-o stargrazer.exe"
else
  OUTPUT="-o stargrazer"
fi

echo "=== Building stargrazer ==="
echo "Platform: $PLATFORM"
echo "Flags: $FLAGS $TAGS $OUTPUT"
[ -n "$VERSION" ] && echo "Version: $VERSION"

wails build -platform "$PLATFORM" $FLAGS $TAGS $OUTPUT

echo "Build complete."
ls -lh build/bin/
