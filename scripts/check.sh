#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

services=(users products search inventory orders checkout cart web)

go_packages=(
  ./services/users/...
  ./services/products/...
  ./services/search/...
  ./services/inventory/...
  ./services/orders/...
  ./services/checkout/...
  ./services/cart/...
  ./services/web/...
  ./shared/auth/...
  ./shared/err/dberr/...
  ./shared/err/grpcerr/...
  ./shared/messaging/...
  ./shared/observe/metric/...
  ./shared/observe/log/...
  ./shared/observe/trace/...
  ./shared/proto/...
  ./shared/runtime/...
  ./shared/testutil/postgres/...
  ./shared/testutil/kafka/...
  ./shared/testutil/meilisearch/...
  ./shared/testutil/redis/...
)

lint() {
  golangci-lint run --config=.golangci.yml "${go_packages[@]}"
}

charts() {
  local environment chart values
  for environment in dev prod; do
    for chart in infra/charts/*; do
      helm lint "$chart"
      values=()
      if [ -f "$chart/values-${environment}.yaml" ]; then
        values=(-f "$chart/values-${environment}.yaml")
      fi
      helm lint "$chart" "${values[@]}"
      helm template marketplace "$chart" --namespace ecommerce "${values[@]}" >/dev/null
    done
    helm lint infra/argocd/app-of-apps
    helm template root infra/argocd/app-of-apps --set namePrefix="$environment" >/dev/null
  done
}

test_service() {
  go test "./services/$1/..."
}

test_all() {
  local service
  for service in "${services[@]}"; do
    test_service "$service"
  done
}

require_service() {
  local service
  for service in "${services[@]}"; do
    [ "$service" = "$1" ] && return 0
  done
  echo "unknown service: $1 (expected one of: ${services[*]})" >&2
  exit 2
}

case "${1:-all}" in
  all)
    lint
    charts
    test_all
    ;;
  lint) lint ;;
  charts) charts ;;
  *)
    require_service "$1"
    test_service "$1"
    ;;
esac
