#!/usr/bin/env bash
# Shared image catalog for local Colima builds and GHCR release CI.
#
# Usage:
#   ./tools/build-images.sh              # build all + restart local workloads
#   ./tools/build-images.sh web          # build one
#   ./tools/build-images.sh --list       # names only
#   ./tools/build-images.sh --matrix     # JSON for GitHub Actions matrix.include
#
# Env:
#   NS_ECOM / NS_KAFKA  restart namespaces (defaults: ecommerce / kafka)
#   SKIP_RESTART=1      build only
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NS_ECOM="${NS_ECOM:-ecommerce}"
NS_KAFKA="${NS_KAFKA:-kafka}"

# name|dockerfile|build_args (comma-separated KEY=VAL)|restart
# restart: deploy:<name> | connect | -
IMAGES=(
  'web|infra/docker/web.Dockerfile|-|deploy:web'
  'users|infra/docker/go-service.Dockerfile|BUILD_PKG=./services/users/cmd/users,BUILD_BIN=users,EXPOSE_PORT=9091|deploy:users'
  'users-migrator|infra/docker/goose-migrator.Dockerfile|MIGRATIONS_DIR=services/users/db/migrations|-'
  'products|infra/docker/go-service.Dockerfile|BUILD_PKG=./services/products/cmd/products,BUILD_BIN=products,EXPOSE_PORT=9092|deploy:products'
  'products-migrator|infra/docker/goose-migrator.Dockerfile|MIGRATIONS_DIR=services/products/db/migrations|-'
  'orders|infra/docker/go-service.Dockerfile|BUILD_PKG=./services/orders/cmd/orders,BUILD_BIN=orders,EXPOSE_PORT=9093|deploy:orders'
  'orders-migrator|infra/docker/goose-migrator.Dockerfile|MIGRATIONS_DIR=services/orders/db/migrations|-'
  'cart|infra/docker/go-service.Dockerfile|BUILD_PKG=./services/cart/cmd/cart,BUILD_BIN=cart,EXPOSE_PORT=9094|deploy:cart'
  'payment|infra/docker/go-service.Dockerfile|BUILD_PKG=./services/payment/cmd/payment,BUILD_BIN=payment,EXPOSE_PORT=9096|deploy:payment'
  'payment-migrator|infra/docker/goose-migrator.Dockerfile|MIGRATIONS_DIR=services/payment/db/migrations|-'
  'payment-gateway-simulator|infra/docker/go-service.Dockerfile|BUILD_PKG=./tools/payment-gateway-simulator,BUILD_BIN=payment-gateway-simulator,EXPOSE_PORT=8097|deploy:payment-gateway-simulator'
  'connect-debezium|infra/docker/connect-debezium.Dockerfile|-|connect'
)

split_fields() {
  local row="$1"
  IFS='|' read -r IMAGE DOCKERFILE BUILD_ARGS RESTART <<<"$row"
}

json_escape() {
  # Escape a string for JSON (no surrounding quotes).
  local s=${1//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  printf '%s' "$s"
}

build_args_to_docker() {
  local raw="$1"
  local -a out=()
  if [[ -z "$raw" || "$raw" == "-" ]]; then
    return 0
  fi
  local IFS=,
  local part
  for part in $raw; do
    out+=(--build-arg "$part")
  done
  printf '%s\n' "${out[@]}"
}

build_args_to_gha() {
  # Multiline KEY=VAL for docker/build-push-action.
  local raw="$1"
  if [[ -z "$raw" || "$raw" == "-" ]]; then
    printf ''
    return 0
  fi
  local IFS=,
  local part
  local first=1
  for part in $raw; do
    if [[ $first -eq 1 ]]; then
      first=0
    else
      printf '\n'
    fi
    printf '%s' "$part"
  done
}

emit_matrix() {
  local first=1
  printf '['
  local row
  for row in "${IMAGES[@]}"; do
    split_fields "$row"
    local args
    args="$(build_args_to_gha "$BUILD_ARGS")"
    if [[ $first -eq 1 ]]; then
      first=0
    else
      printf ','
    fi
    printf '{"image":"%s","dockerfile":"%s","build_args":"%s"}' \
      "$(json_escape "$IMAGE")" \
      "$(json_escape "$DOCKERFILE")" \
      "$(json_escape "$args")"
  done
  printf ']\n'
}

list_names() {
  local row
  for row in "${IMAGES[@]}"; do
    split_fields "$row"
    printf '%s\n' "$IMAGE"
  done
}

restart_for() {
  local spec="$1" name="$2"
  case "$spec" in
    -) ;;
    deploy:*)
      local deploy="${spec#deploy:}"
      if kubectl -n "$NS_ECOM" get deploy "$deploy" >/dev/null 2>&1; then
        echo "==> rollout restart deploy/$deploy ($NS_ECOM)"
        kubectl -n "$NS_ECOM" rollout restart "deploy/$deploy"
      fi
      ;;
    connect)
      if kubectl -n "$NS_KAFKA" get kafkaconnect ecommerce-connect-cluster >/dev/null 2>&1; then
        echo "==> delete connect pods to pick up $name"
        kubectl -n "$NS_KAFKA" delete pod -l strimzi.io/cluster=ecommerce-connect-cluster --ignore-not-found
      fi
      ;;
    *)
      echo "unknown restart spec: $spec" >&2
      return 1
      ;;
  esac
}

build_one() {
  local row="$1"
  split_fields "$row"
  local -a args=()
  local line
  while IFS= read -r line; do
    [[ -n "$line" ]] && args+=("$line")
  done < <(build_args_to_docker "$BUILD_ARGS")

  echo "==> docker build -t $IMAGE"
  docker build -t "$IMAGE" -f "$DOCKERFILE" "${args[@]}" .

  if [[ "${SKIP_RESTART:-}" != "1" ]]; then
    restart_for "$RESTART" "$IMAGE"
  fi
}

find_row() {
  local want="$1" row
  for row in "${IMAGES[@]}"; do
    split_fields "$row"
    if [[ "$IMAGE" == "$want" ]]; then
      printf '%s\n' "$row"
      return 0
    fi
  done
  return 1
}

usage() {
  cat <<'EOF'
Usage: build-images.sh [--list|--matrix|<image>|]
  (no args)   build all images into the current Docker context and restart pods
  <image>     build one image from the catalog
  --list      print image names
  --matrix    print GitHub Actions matrix.include JSON
EOF
}

main() {
  local arg="${1:-}"
  case "$arg" in
    -h|--help)
      usage
      ;;
    --list)
      list_names
      ;;
    --matrix)
      emit_matrix
      ;;
    "")
      local row
      for row in "${IMAGES[@]}"; do
        build_one "$row"
      done
      echo "Done. Tip: build-images <name> for one image."
      ;;
    *)
      local row
      if ! row="$(find_row "$arg")"; then
        echo "unknown image: $arg (try --list)" >&2
        exit 1
      fi
      build_one "$row"
      echo "Done."
      ;;
  esac
}

main "$@"
