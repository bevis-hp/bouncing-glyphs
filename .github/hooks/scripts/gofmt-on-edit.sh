#!/bin/sh
set -eu

workspace_root=$(pwd -P)

payload_file=$(mktemp)
paths_file=$(mktemp)
targets_file=$(mktemp)
trap 'rm -f "$payload_file" "$paths_file" "$targets_file"' EXIT HUP INT TERM

cat >"$payload_file"

if ! command -v gofmt >/dev/null 2>&1; then
  printf '%s\n' '{"continue":true,"systemMessage":"gofmt-on-edit: gofmt not found; skipped formatting."}'
  exit 0
fi

if ! command -v git >/dev/null 2>&1; then
  printf '%s\n' '{"continue":true}'
  exit 0
fi

# Only act on successful edit-like tool payloads instead of formatting after reads.
if ! grep -Eq '"toolName"[[:space:]]*:[[:space:]]*"(apply_patch|create_file|vscode_renameSymbol)"|"recipient_name"[[:space:]]*:[[:space:]]*"functions\.(apply_patch|create_file|vscode_renameSymbol)"|\*\*\* (Add|Update) File: .*\.go' "$payload_file"; then
  printf '%s\n' '{"continue":true}'
  exit 0
fi

LC_ALL=C grep -Eo '(/[^"[:space:]]+\.go|[[:alnum:]_.-]+(/[[:alnum:]_.-]+)*\.go)' "$payload_file" | sort -u >"$paths_file" || true

if [ ! -s "$paths_file" ]; then
  printf '%s\n' '{"continue":true}'
  exit 0
fi

while IFS= read -r path; do
  case "$path" in
    /*)
      candidate="$path"
      ;;
    *)
      candidate="$workspace_root/$path"
      ;;
  esac

  if [ ! -f "$candidate" ]; then
    continue
  fi

  candidate_dir=$(dirname "$candidate")
  candidate_base=$(basename "$candidate")
  resolved_dir=$(cd "$candidate_dir" 2>/dev/null && pwd -P) || continue
  resolved_path="$resolved_dir/$candidate_base"

  case "$resolved_path" in
    "$workspace_root"/*.go)
      rel_path=${resolved_path#"$workspace_root"/}
      if [ -n "$(git status --porcelain --untracked-files=all -- "$rel_path")" ]; then
        printf '%s\n' "$resolved_path" >>"$targets_file"
      fi
      ;;
  esac
done <"$paths_file"

if [ ! -s "$targets_file" ]; then
  printf '%s\n' '{"continue":true}'
  exit 0
fi

sort -u "$targets_file" | xargs gofmt -w
count=$(sort -u "$targets_file" | wc -l | tr -d ' ')
printf '{"continue":true,"systemMessage":"gofmt-on-edit: formatted %s changed Go file(s)."}\n' "$count"