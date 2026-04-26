#!/usr/bin/env bash
# Generate a changelog between two git refs.
# Usage: ./generate-changelog.sh <prev_tag> [output_file]
set -euo pipefail

PREV_TAG="${1:-}"
OUTPUT="${2:-changelog.md}"

echo "## Changelog" > "$OUTPUT"
echo "" >> "$OUTPUT"

if [ -n "$PREV_TAG" ]; then
  echo "Changes since \`$PREV_TAG\`:" >> "$OUTPUT"
  echo "" >> "$OUTPUT"
  git log "$PREV_TAG"..HEAD --pretty=format:"- %s (\`%h\`)" --no-merges >> "$OUTPUT"
else
  echo "Initial release." >> "$OUTPUT"
  echo "" >> "$OUTPUT"
  git log --pretty=format:"- %s (\`%h\`)" --no-merges >> "$OUTPUT"
fi

echo "" >> "$OUTPUT"
echo "Changelog written to $OUTPUT"
