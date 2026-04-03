#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
meta_file="$repo_root/.github/repo-metadata.env"

if [[ ! -f "$meta_file" ]]; then
  echo "Metadata file not found: $meta_file" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$meta_file"

origin_url="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "$origin_url" ]]; then
  echo "No origin remote found; skipping metadata sync"
  exit 0
fi

slug=""
if [[ "$origin_url" =~ github.com[:/]([^/]+)/([^/.]+)(\\.git)?$ ]]; then
  slug="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
fi

if [[ -z "$slug" ]]; then
  echo "Could not parse GitHub repository slug from origin URL: $origin_url" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "GitHub CLI (gh) is not installed; skipping metadata sync"
  exit 0
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "GitHub CLI is not authenticated; skipping metadata sync"
  exit 0
fi

gh repo edit "$slug" --description "$DESCRIPTION" >/dev/null

for topic in $TOPICS; do
  gh repo edit "$slug" --add-topic "$topic" >/dev/null
done

echo "Synced GitHub description/topics for $slug"
