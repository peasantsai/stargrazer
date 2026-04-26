#!/usr/bin/env bash
# Remove non-essential files from the Chromium bundle to reduce size.
# Usage: ./clean-chromium.sh <bundle_dir>
set -euo pipefail

DIR="$1"
cd "$DIR"

echo "Cleaning Chromium bundle in $DIR..."

# Remove unnecessary directories
rm -rf swiftshader VisualElements default_apps MEIPreload Dictionaries
rm -rf "First Run" IwaKeyDistribution

# Remove unnecessary executables and files
rm -f chrome_200_percent.pak
rm -f notification_helper* elevation_service* chrome_proxy*
rm -f chrome_pwa_launcher* chrome_wer* elevated_tracing_service*
rm -f *.debug *.pdb

# Keep only English locale
if [ -d locales ]; then
  find locales -type f ! -name 'en-US.pak' -delete 2>/dev/null || true
fi

echo "=== Bundle contents ==="
ls -la
echo "=== Size ==="
du -sh .
