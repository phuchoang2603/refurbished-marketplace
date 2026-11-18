#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

mapfile -t module_dirs < <(go work edit -json | jq -r '.Use[].DiskPath')

layout() {
  local stray
  stray=$(
    find . -name '*.sql' -not -path './.devenv/*' -not -path './services/*/db/*'
    find . -name '*.proto' -not -path './.devenv/*' -not -path './shared/proto/*/v1/*'
  )
  if [ -n "$stray" ]; then
    echo "SQL belongs under services/<service>/db/ and protobuf under shared/proto/<domain>/v1/:" >&2
    echo "$stray" >&2
    exit 1
  fi
}

golangci() {
  golangci-lint run "${module_dirs[@]/%//...}"
}

charts() {
  local environment chart name values rendered
  for environment in dev prod; do
    for chart in infra/charts/*; do
      name=$(basename "$chart")
      values=()
      if [ -f "$chart/values-${environment}.yaml" ]; then
        values=(-f "$chart/values-${environment}.yaml")
      fi
      helm lint "$chart" "${values[@]}"
      rendered=$(helm template "$name" "$chart" --namespace ecommerce "${values[@]}")
      conftest test --combine --policy policy --namespace images --namespace doppler - <<<"$rendered"
      if [ "$name" = refurbished-marketplace ]; then
        conftest test --combine --policy policy --namespace mesh - <<<"$rendered"
      fi
    done
    helm lint infra/argocd/app-of-apps
    rendered=$(helm template root infra/argocd/app-of-apps --set namePrefix="$environment")
    conftest test --combine --policy policy --namespace argo --namespace doppler - <<<"$rendered"
  done
}

layout
golangci
charts
