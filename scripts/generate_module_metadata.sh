#!/usr/bin/env bash
set -euo pipefail

root_dir="${1:-modules}"

if [[ ! -d "$root_dir" ]]; then
  echo "modules root not found: $root_dir" >&2
  exit 1
fi

go build -o pltf ./main.go

for dir in "$root_dir"/aws_*; do
  if [[ ! -d "$dir" ]]; then
    continue
  fi
  if ! find "$dir" -maxdepth 1 -type f -name '*.tf' -print -quit | grep -q .; then
    continue
  fi
  echo "Generating module.yaml in $dir"
  ./pltf module init --path "$dir" --force --provider aws
 done


for dir in "$root_dir"/azure_*; do
  if [[ ! -d "$dir" ]]; then
    continue
  fi
  if ! find "$dir" -maxdepth 1 -type f -name '*.tf' -print -quit | grep -q .; then
    continue
  fi
  echo "Generating module.yaml in $dir"
  ./pltf module init --path "$dir" --force --provider azure
 done


for dir in "$root_dir"/gcp_*; do
  if [[ ! -d "$dir" ]]; then
    continue
  fi
  if ! find "$dir" -maxdepth 1 -type f -name '*.tf' -print -quit | grep -q .; then
    continue
  fi
  echo "Generating module.yaml in $dir"
  ./pltf module init --path "$dir" --force --provider gcp
 done

./pltf module init --path modules/helm_chart --force --provider helm