#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

status=0
while read -r dir; do
  govulncheck -C "$dir" ./... || status=1
done < <(go work edit -json | jq -r '.Use[].DiskPath')
exit "$status"
