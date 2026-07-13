#!/usr/bin/env bash
# Build marketplace images into the current Docker context (Colima) and restart pods.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NS_ECOM="${NS_ECOM:-ecommerce}"
NS_KAFKA="${NS_KAFKA:-kafka}"
ONLY="${1:-}"

build() {
  local image="$1"
  shift
  if [[ -n "$ONLY" && "$ONLY" != "$image" ]]; then
    return 0
  fi
  echo "==> docker build -t $image"
  docker build -t "$image" "$@"
}

build web -f infra/docker/web.Dockerfile .

build users \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./services/users/cmd/users \
  --build-arg BUILD_BIN=users \
  --build-arg EXPOSE_PORT=9091 \
  .

build users-migrator \
  -f infra/docker/goose-migrator.Dockerfile \
  --build-arg MIGRATIONS_DIR=services/users/db/migrations \
  .

build products \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./services/products/cmd/products \
  --build-arg BUILD_BIN=products \
  --build-arg EXPOSE_PORT=9092 \
  .

build products-migrator \
  -f infra/docker/goose-migrator.Dockerfile \
  --build-arg MIGRATIONS_DIR=services/products/db/migrations \
  .

build orders \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./services/orders/cmd/orders \
  --build-arg BUILD_BIN=orders \
  --build-arg EXPOSE_PORT=9093 \
  .

build orders-migrator \
  -f infra/docker/goose-migrator.Dockerfile \
  --build-arg MIGRATIONS_DIR=services/orders/db/migrations \
  .

build cart \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./services/cart/cmd/cart \
  --build-arg BUILD_BIN=cart \
  --build-arg EXPOSE_PORT=9094 \
  .

build payment \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./services/payment/cmd/payment \
  --build-arg BUILD_BIN=payment \
  --build-arg EXPOSE_PORT=9096 \
  .

build payment-migrator \
  -f infra/docker/goose-migrator.Dockerfile \
  --build-arg MIGRATIONS_DIR=services/payment/db/migrations \
  .

build payment-gateway-simulator \
  -f infra/docker/go-service.Dockerfile \
  --build-arg BUILD_PKG=./tools/payment-gateway-simulator \
  --build-arg BUILD_BIN=payment-gateway-simulator \
  --build-arg EXPOSE_PORT=8097 \
  .

build connect-debezium -f infra/docker/connect-debezium.Dockerfile .

restart_deploy() {
  local ns="$1" name="$2"
  if kubectl -n "$ns" get deploy "$name" >/dev/null 2>&1; then
    echo "==> rollout restart deploy/$name ($ns)"
    kubectl -n "$ns" rollout restart "deploy/$name"
  fi
}

if [[ -z "$ONLY" || "$ONLY" == web ]]; then restart_deploy "$NS_ECOM" web; fi
if [[ -z "$ONLY" || "$ONLY" == users ]]; then restart_deploy "$NS_ECOM" users; fi
if [[ -z "$ONLY" || "$ONLY" == products ]]; then restart_deploy "$NS_ECOM" products; fi
if [[ -z "$ONLY" || "$ONLY" == orders ]]; then restart_deploy "$NS_ECOM" orders; fi
if [[ -z "$ONLY" || "$ONLY" == cart ]]; then restart_deploy "$NS_ECOM" cart; fi
if [[ -z "$ONLY" || "$ONLY" == payment ]]; then restart_deploy "$NS_ECOM" payment; fi
if [[ -z "$ONLY" || "$ONLY" == payment-gateway-simulator ]]; then restart_deploy "$NS_ECOM" payment-gateway-simulator; fi

if [[ -z "$ONLY" || "$ONLY" == connect-debezium ]]; then
  if kubectl -n "$NS_KAFKA" get kafkaconnect ecommerce-connect-cluster >/dev/null 2>&1; then
    echo "==> delete connect pods to pick up connect-debezium image"
    kubectl -n "$NS_KAFKA" delete pod -l strimzi.io/cluster=ecommerce-connect-cluster --ignore-not-found
  fi
fi

echo "Done. Optional: build-images <image> (or ./tools/build-images.sh <image>) for one target."
