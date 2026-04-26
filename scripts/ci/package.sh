#!/usr/bin/env bash
# Package the build output into a distributable archive.
# Usage: ./package.sh <goos> <artifact_name> [version]
set -euo pipefail

GOOS="$1"
ARTIFACT_NAME="$2"
VERSION="${3:-}"

if [ -n "$VERSION" ]; then
  ARTIFACT="${ARTIFACT_NAME}-${VERSION}"
else
  ARTIFACT="$ARTIFACT_NAME"
fi

mkdir -p dist

echo "Packaging $ARTIFACT for $GOOS..."

case "$GOOS" in
  windows)
    cp build/bin/*installer* "dist/${ARTIFACT}-setup.exe" 2>/dev/null || true
    cd build/bin && 7z a "$GITHUB_WORKSPACE/dist/${ARTIFACT}.zip" . -r
    ;;
  darwin)
    cd build/bin && zip -r "$GITHUB_WORKSPACE/dist/${ARTIFACT}.zip" .
    ;;
  linux)
    cd build/bin && tar czf "$GITHUB_WORKSPACE/dist/${ARTIFACT}.tar.gz" .
    ;;
  *)
    echo "Unknown platform: $GOOS"
    exit 1
    ;;
esac

echo "Package created in dist/"
ls -lh "$GITHUB_WORKSPACE/dist/"
