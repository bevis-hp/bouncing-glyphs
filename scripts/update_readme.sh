#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
readme="$repo_root/README.md"

if [[ ! -f "$readme" ]]; then
  echo "README.md not found at repository root" >&2
  exit 1
fi

help_text="$(go run . -h 2>&1 || true)"
help_text="$(printf '%s\n' "$help_text" | sed -E '1s|^Usage of .*:|Usage of glyphfall:|')"

block="$(cat <<EOF
<!-- BEGIN AUTO-CLI -->
\`\`\`text
$help_text
\`\`\`
<!-- END AUTO-CLI -->
EOF
)"

tmp="$(mktemp)"
block_file="$(mktemp)"
printf '%s\n' "$block" > "$block_file"

if grep -q "<!-- BEGIN AUTO-CLI -->" "$readme" && grep -q "<!-- END AUTO-CLI -->" "$readme"; then
  awk -v block_file="$block_file" '
    /<!-- BEGIN AUTO-CLI -->/ {
      while ((getline line < block_file) > 0) {
        print line
      }
      close(block_file)
      skip = 1
      next
    }
    /<!-- END AUTO-CLI -->/ {
      if (skip == 1) {
        skip = 0
        next
      }
    }
    skip != 1 { print }
  ' "$readme" > "$tmp"
else
  cat "$readme" > "$tmp"
  cat >> "$tmp" <<EOF

## CLI Reference (Auto-generated)

This section is refreshed by scripts/update_readme.sh.

$block
EOF
fi

if ! cmp -s "$readme" "$tmp"; then
  mv "$tmp" "$readme"
else
  rm -f "$tmp"
fi

rm -f "$block_file"
