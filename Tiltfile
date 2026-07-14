# Hybrid local DX:
# - Tilt owns refurbished-marketplace (images + Helm + templ/tailwind watches + debug PFs)
# - Argo CD (installed here) owns operators, Istio, Kafka, observability, Cloudflare Tunnel

load('ext://namespace', 'namespace_create')

namespace_create('ecommerce')
namespace_create('operators')

GATEWAY_API_VERSION = 'v1.3.0'
DOPPLER_SECRET = 'infra/k8s/doppler-token.dev.secret.yaml'
ARGO_NS = 'argo-cd'

### Argo CD + infra Applications (no resource_deps) ###

local_resource(
  'argocd-install',
  '''
  set -euo pipefail
  helm repo add argo https://argoproj.github.io/argo-helm >/dev/null 2>&1 || true
  helm repo update argo >/dev/null
  helm upgrade --install argocd argo/argo-cd \
    --namespace argo-cd \
    --create-namespace \
    --set fullnameOverride=argocd \
    --wait \
    --timeout 10m
  ''',
  labels=['argocd'],
)

local_resource(
  'gateway-api-crds',
  '''
  set -euo pipefail
  kubectl get crd gateways.gateway.networking.k8s.io >/dev/null 2>&1 || \
    kubectl apply --server-side -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/%s/standard-install.yaml"
  ''' % GATEWAY_API_VERSION,
  labels=['argocd'],
)

local_resource(
  'doppler-secret',
  '''
  set -euo pipefail
  if [[ ! -f "%s" ]]; then
    echo "Missing %s — copy from .example and paste the Doppler dev token." >&2
    exit 1
  fi
  kubectl get ns operators >/dev/null 2>&1 || kubectl create namespace operators
  kubectl apply -f %s
  ''' % (DOPPLER_SECRET, DOPPLER_SECRET, DOPPLER_SECRET),
  labels=['argocd'],
)

local_resource(
  'argocd-local-apps',
  '''
  set -euo pipefail
  REVISION="${ARGO_REVISION:-$(git branch --show-current)}"
  if [[ -z "$REVISION" || "$REVISION" == "HEAD" ]]; then
    echo "Detached HEAD; set ARGO_REVISION to a pushable branch name." >&2
    exit 1
  fi
  echo "==> Pinning local Argo Applications to $REVISION"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT
  sed "s/^    targetRevision: .*/    targetRevision: ${REVISION}/" \
    infra/argocd/local/root.yaml >"$tmpdir/root.yaml"
  kubectl apply -n argo-cd -f "$tmpdir/root.yaml"
  for _ in $(seq 1 60); do
    count="$(kubectl -n argo-cd get applications -o name 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "$count" -gt 1 ]]; then
      break
    fi
    sleep 2
  done
  kubectl -n argo-cd get applications -o name | while read -r app; do
    kubectl -n argo-cd patch "$app" --type merge \
      -p "{\"spec\":{\"source\":{\"targetRevision\":\"${REVISION}\"}}}"
  done
  kubectl -n argo-cd get applications -o custom-columns='NAME:.metadata.name,REVISION:.spec.source.targetRevision' || true
  ''',
  labels=['argocd'],
)

### Marketplace Helm (Tilt-owned) ###

k8s_kind('Cluster', pod_readiness='wait')

k8s_yaml(helm(
  './infra/charts/refurbished-marketplace',
  name='refurbished-marketplace',
  namespace='ecommerce',
  values=['./infra/charts/refurbished-marketplace/values.yaml'],
))

GO_WORKSPACE_ONLY = [
  './go.work',
  './go.work.sum',
  './shared',
  './services',
  './tools',
]

def go_service(name, pkg, port):
  docker_build(
    name,
    '.',
    dockerfile='./infra/docker/go-service.Dockerfile',
    build_args={
      'BUILD_PKG': pkg,
      'BUILD_BIN': name,
      'EXPOSE_PORT': str(port),
    },
    only=GO_WORKSPACE_ONLY,
  )

def goose_migrator(name, migrations_dir):
  docker_build(
    name + '-migrator',
    '.',
    dockerfile='./infra/docker/goose-migrator.Dockerfile',
    build_args={
      'MIGRATIONS_DIR': migrations_dir,
    },
    only=['./' + migrations_dir],
  )

### Web ###

local_resource(
  'templ-watch',
  serve_cmd='cd services/web && templ generate --watch --proxy="http://localhost:8080" --open-browser=false',
  labels=['web'],
)

local_resource(
  'tailwind-watch',
  serve_cmd='cd services/web && tailwindcss -c tailwind.config.js -i tailwind.css -o static/app.css --watch=always',
  labels=['web'],
)

docker_build(
  'web',
  '.',
  dockerfile='./infra/docker/web.Dockerfile',
  only=GO_WORKSPACE_ONLY,
)
k8s_resource('web', port_forwards=['8080:8080'], labels=['web'])

### Users ###

goose_migrator('users', 'services/users/db/migrations')
go_service('users', './services/users/cmd/users', 9091)

k8s_resource(
  'users-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'users-db'}],
  port_forwards=['5432:5432'],
  labels=['users'],
)
k8s_resource('users-migrate', labels=['users'])
k8s_resource('users', port_forwards=['9091:9091'], labels=['users'])

### Products ###

goose_migrator('products', 'services/products/db/migrations')
go_service('products', './services/products/cmd/products', 9092)

k8s_resource(
  'products-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'products-db'}],
  port_forwards=['5433:5432'],
  labels=['products'],
)
k8s_resource('products-migrate', labels=['products'])
k8s_resource('products', port_forwards=['9092:9092'], labels=['products'])

### Orders ###

goose_migrator('orders', 'services/orders/db/migrations')
go_service('orders', './services/orders/cmd/orders', 9093)

k8s_resource(
  'orders-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}],
  port_forwards=['5434:5432'],
  labels=['orders'],
)
k8s_resource('orders-migrate', labels=['orders'])
k8s_resource('orders', port_forwards=['9093:9093'], labels=['orders'])

### Cart ###

go_service('cart', './services/cart/cmd/cart', 9094)
k8s_resource('cart', port_forwards=['9094:9094'], labels=['cart'])

### Payment ###

goose_migrator('payment', 'services/payment/db/migrations')
go_service('payment', './services/payment/cmd/payment', 9096)

k8s_resource(
  'payment-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'payment-db'}],
  port_forwards=['5436:5432'],
  labels=['payment'],
)
k8s_resource('payment-migrate', labels=['payment'])
k8s_resource('payment', port_forwards=['9096:9096'], labels=['payment'])

go_service('payment-gateway-simulator', './tools/payment-gateway-simulator', 8097)
k8s_resource('payment-gateway-simulator', port_forwards=['8097:8097'], labels=['payment'])

### Kafka Connect image (Argo kafka chart; short name for local Colima) ###

docker_build(
  'connect-debezium',
  '.',
  dockerfile='./infra/docker/connect-debezium.Dockerfile',
  only=['./infra/docker/connect-debezium.Dockerfile'],
)
