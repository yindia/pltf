#!/usr/bin/env bash
set -euo pipefail

root_dir="${1:-modules}"

if [[ ! -d "$root_dir" ]]; then
  echo "modules root not found: $root_dir" >&2
  exit 1
fi

for dir in "$root_dir"/*; do
  if [[ ! -d "$dir" ]]; then
    continue
  fi
  if ! find "$dir" -maxdepth 1 -type f -name '*.tf' -print -quit | grep -q .; then
    continue
  fi
  echo "Generating module.yaml in $dir"
  go run ./main.go module init --path "$dir" --force
 done
