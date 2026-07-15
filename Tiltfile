# Hybrid local DX:
# - Tilt owns refurbished-marketplace (images + Helm + templ/tailwind watches + debug PFs)
# - Argo CD (installed here) owns operators, Istio, Kafka, apps-only observability, Cloudflare Tunnel

local_resource(
  'connect-debezium',
  'docker build -t connect-debezium:latest -f infra/docker/connect-debezium.Dockerfile .',
  deps=['infra/docker/connect-debezium.Dockerfile'],
)

### Argo CD + infra apps ###

local_resource(
  'argocd-install',
  '''
  helm repo add argo https://argoproj.github.io/argo-helm || true
  helm upgrade --install argocd argo/argo-cd \
    --namespace argo-cd --create-namespace \
    --set fullnameOverride=argocd \
    --wait --timeout 10m
  ''',
  labels=['argocd'],
)

local_resource(
  'gateway-api-crds',
  'kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml',
)

local_resource(
  'doppler-secret',
  '''
  kubectl apply -f - <<'EOF'
apiVersion: v1
kind: Namespace
metadata:
  name: operators
EOF
  kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml
  ''',
)

# Apply only local-root via kubectl — never k8s_yaml Application CRs.
# Tilt would own them and force-delete on update; Argo finalizers then hang the deploy.
local_resource(
  'argocd-local-root',
  '''
  set -euo pipefail
  REV="$(git branch --show-current)"
  # Pin targetRevision to the current branch (Argo children inherit via $ARGOCD_APP_SOURCE_TARGET_REVISION).
  sed "s|^    targetRevision:.*|    targetRevision: ${REV}|" infra/argocd/local/root.yaml | kubectl apply -f -
  ''',
  resource_deps=['argocd-install', 'gateway-api-crds', 'doppler-secret'],
  labels=['argocd'],
)

# Marketplace Helm needs CNPG + ESO CRDs from Argo-managed operators before apply.
local_resource(
  'argocd-operators-ready',
  '''
  set -euo pipefail
  kubectl wait --for=condition=Established --timeout=10m crd/clusters.postgresql.cnpg.io
  kubectl wait --for=condition=Established --timeout=10m crd/externalsecrets.external-secrets.io
  kubectl wait --for=condition=Established --timeout=10m crd/clustersecretstores.external-secrets.io
  ''',
  resource_deps=['argocd-local-root'],
  labels=['argocd'],
)

local_resource(
  'argocd-ui',
  serve_cmd='kubectl -n argo-cd port-forward svc/argocd-server 8088:443',
  resource_deps=['argocd-install'],
  links=['https://localhost:8088'],
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

k8s_resource(
  new_name='marketplace-secrets',
  objects=[
    'users-app:externalsecret',
    'products-app:externalsecret',
    'orders-app:externalsecret',
    'payment-app:externalsecret',
    'users-auth:externalsecret',
  ],
  resource_deps=['argocd-operators-ready'],
)

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
k8s_resource('web', port_forwards=['8080:8080'], resource_deps=['argocd-operators-ready'], labels=['web'])

### Users ###

goose_migrator('users', 'services/users/db/migrations')
go_service('users', './services/users/cmd/users', 9091)

k8s_resource(
  'users-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'users-db'}],
  port_forwards=['5432:5432'],
  resource_deps=['argocd-operators-ready', 'marketplace-secrets'],
  labels=['users'],
)
k8s_resource('users-migrate', resource_deps=['users-db'], labels=['users'])
k8s_resource('users', port_forwards=['9091:9091'], resource_deps=['users-db'], labels=['users'])

### Products ###

goose_migrator('products', 'services/products/db/migrations')
go_service('products', './services/products/cmd/products', 9092)

k8s_resource(
  'products-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'products-db'}],
  port_forwards=['5433:5432'],
  resource_deps=['argocd-operators-ready', 'marketplace-secrets'],
  labels=['products'],
)
k8s_resource('products-migrate', resource_deps=['products-db'], labels=['products'])
k8s_resource('products', port_forwards=['9092:9092'], resource_deps=['products-db'], labels=['products'])

### Orders ###

goose_migrator('orders', 'services/orders/db/migrations')
go_service('orders', './services/orders/cmd/orders', 9093)

k8s_resource(
  'orders-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}],
  port_forwards=['5434:5432'],
  resource_deps=['argocd-operators-ready', 'marketplace-secrets'],
  labels=['orders'],
)
k8s_resource('orders-migrate', resource_deps=['orders-db'], labels=['orders'])
k8s_resource('orders', port_forwards=['9093:9093'], resource_deps=['orders-db'], labels=['orders'])

### Cart ###

go_service('cart', './services/cart/cmd/cart', 9094)
k8s_resource('cart', port_forwards=['9094:9094'], resource_deps=['argocd-operators-ready'], labels=['cart'])

### Payment ###

goose_migrator('payment', 'services/payment/db/migrations')
go_service('payment', './services/payment/cmd/payment', 9096)

k8s_resource(
  'payment-db',
  extra_pod_selectors=[{'cnpg.io/cluster': 'payment-db'}],
  port_forwards=['5436:5432'],
  resource_deps=['argocd-operators-ready', 'marketplace-secrets'],
  labels=['payment'],
)
k8s_resource('payment-migrate', resource_deps=['payment-db'], labels=['payment'])
k8s_resource('payment', port_forwards=['9096:9096'], resource_deps=['payment-db'], labels=['payment'])

go_service('payment-gateway-simulator', './tools/payment-gateway-simulator', 8097)
k8s_resource('payment-gateway-simulator', port_forwards=['8097:8097'], resource_deps=['argocd-operators-ready'], labels=['payment'])
