#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

services=(users products search inventory orders checkout cart web)

if [ $# -eq 0 ]; then
  set -- "${services[@]}"
fi

for service in "$@"; do
  if [[ ! " ${services[*]} " =~ " ${service} " ]]; then
    echo "unknown service: ${service} (expected one of: ${services[*]})" >&2
    exit 2
  fi
done

for service in "$@"; do
  go test "./services/${service}/..."
done
