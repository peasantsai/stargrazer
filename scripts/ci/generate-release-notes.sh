#!/usr/bin/env bash
# Generate release notes with download table and changelog.
# Usage: ./generate-release-notes.sh <version> <release_files_dir> <changelog_file> [output_file]
set -euo pipefail

VERSION="$1"
FILES_DIR="$2"
CHANGELOG="$3"
OUTPUT="${4:-release-notes.md}"

cat > "$OUTPUT" << HEADER
## Stargrazer ${VERSION}

### Downloads

| Platform | Architecture | File | Size |
|----------|-------------|------|------|
HEADER

for f in "$FILES_DIR"/*; do
  [ ! -f "$f" ] && continue
  fname=$(basename "$f")
  size=$(du -h "$f" | cut -f1)
  case "$fname" in
    *linux-amd64*.tar.gz)       echo "| Linux | x86_64 | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *linux-arm64*.tar.gz)       echo "| Linux | ARM64 | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *windows-amd64*setup*)      echo "| Windows | x64 (Installer) | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *windows-amd64*.zip)        echo "| Windows | x64 (Portable) | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *windows-arm64*setup*)      echo "| Windows | ARM64 (Installer) | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *windows-arm64*.zip)        echo "| Windows | ARM64 (Portable) | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *macos-arm64*)              echo "| macOS | Apple Silicon | \`$fname\` | $size |" >> "$OUTPUT" ;;
    *macos-amd64*)              echo "| macOS | Intel | \`$fname\` | $size |" >> "$OUTPUT" ;;
  esac
done

echo "" >> "$OUTPUT"
cat "$CHANGELOG" >> "$OUTPUT"

cat >> "$OUTPUT" << FOOTER

### What's included
- Bundled Ungoogled Chromium with stealth configuration
- Chrome DevTools Protocol (CDP) automation
- Social media session management (6 platforms)
- Cookie-based auth with auto keep-alive scheduler
- Content upload workflows with scheduling
- Dark and light themes

### Requirements
- Windows 10+ / macOS 12+ / Linux (glibc 2.31+)
- WebView2 runtime (Windows, auto-installed)

---
*Built with [Wails v2](https://wails.io) + Go + React + TypeScript*
FOOTER

echo "Release notes written to $OUTPUT"
