load('ext://namespace', 'namespace_create')
namespace_create('ecommerce')
namespace_create('operators')

### External Secrets Operators ###
local_resource(
  'eso-operator-install',
  'helm upgrade --install external-secrets ./infra/charts/operators/external-secrets \
     --namespace operators --create-namespace',
)

k8s_yaml('infra/k8s/doppler-token.dev.secret.yaml')

### Cloudnative-pg Operators ###
k8s_kind('Cluster', pod_readiness='wait')
local_resource(
  'cnpg-operator-install',
  'helm upgrade --install cnpg ./infra/charts/operators/cnpg \
   --namespace operators --create-namespace',
)

### Our helm charts
k8s_yaml(helm(
    './infra/charts/refurbished-marketplace-infra',
    name='refurbished-marketplace-infra',
    namespace='ecommerce',
    values=['./infra/charts/refurbished-marketplace/values.yaml']
))

app_yaml = helm(
    './infra/charts/refurbished-marketplace',
    name='refurbished-marketplace',
    namespace='ecommerce',
    values=['./infra/charts/refurbished-marketplace/values.yaml']
)
k8s_yaml(app_yaml)

### Kafka Cluster and Topics ###
local_resource(
  'kafka-cluster-install',
  'helm upgrade --install strimzi-cluster-operator ./infra/charts/operators/strimzi \
  --namespace operators --create-namespace',
)

k8s_yaml(helm(
  './infra/charts/kafka',
  name='ecommerce-kafka-cluster',
  namespace='ecommerce',
  values=['./infra/charts/kafka/values.yaml']
))

k8s_resource(
    new_name='kafka-cluster', 
    objects=[
        'ecommerce-kafka-cluster:kafka',
        'ecommerce-kafka-cluster-dual-role:kafkanodepool'
    ],
    resource_deps=['kafka-cluster-install'],
    labels=['kafka']
)

k8s_resource(
    new_name='debezium-connect',
    objects=['ecommerce-connect-cluster:kafkaconnect'],
    resource_deps=['kafka-cluster'],
    labels=['kafka']
)
k8s_resource('kafka-ui', port_forwards=['8081:8080'], resource_deps=['kafka-cluster'], labels='kafka')

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

### Web Service ###
local_resource(
    "templ-watch",
    serve_cmd='cd services/web && templ generate --watch  --proxy="http://localhost:8080"  --open-browser=false',
    labels=["web"]
)

local_resource(
    "tailwind-watch",
    serve_cmd='cd services/web && tailwindcss -c tailwind.config.js -i tailwind.css -o static/app.css --watch=always',
    labels=["web"]
)

docker_build(
  'web',
  '.',
  dockerfile='./infra/docker/web.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)
k8s_resource('web', port_forwards=['8080:8080'], labels=["web"])

### Users Service ###
goose_migrator('users', 'services/users/db/migrations')
go_service('users', './services/users/cmd/users', 9091)

k8s_resource('users-db', extra_pod_selectors=[{'cnpg.io/cluster': 'users-db'}], port_forwards=['5432:5432'], resource_deps=['cnpg-operator-install'], labels=['users'])
k8s_resource('users-migrate', resource_deps=['users-db'], labels='users')
k8s_resource('users', port_forwards=['9091:9091'], resource_deps=['users-db'], labels='users')

### Products Service ###
goose_migrator('products', 'services/products/db/migrations')
go_service('products', './services/products/cmd/products', 9092)

k8s_resource('products-db', extra_pod_selectors=[{'cnpg.io/cluster': 'products-db'}], port_forwards=['5433:5432'], resource_deps=['cnpg-operator-install'], labels='products')
k8s_resource('products-migrate', resource_deps=['products-db'], labels='products')
k8s_resource('products', port_forwards=['9092:9092'], resource_deps=['products-db'], labels='products')

### Orders Service ###
goose_migrator('orders', 'services/orders/db/migrations')
go_service('orders', './services/orders/cmd/orders', 9093)

k8s_resource('orders-db', extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}], port_forwards=['5434:5432'], resource_deps=['cnpg-operator-install'], labels='orders')
k8s_resource('orders-migrate', resource_deps=['orders-db'], labels='orders')
k8s_resource('orders', port_forwards=['9093:9093'], resource_deps=['orders-db'], labels='orders')

### Cart Service ###
go_service('cart', './services/cart/cmd/cart', 9094)

k8s_resource('cart', port_forwards=['9094:9094'],  labels='cart')

### Payment Service ###
goose_migrator('payment', 'services/payment/db/migrations')
go_service('payment', './services/payment/cmd/payment', 9096)

k8s_resource('payment-db', extra_pod_selectors=[{'cnpg.io/cluster': 'payment-db'}], port_forwards=['5436:5432'], resource_deps=['cnpg-operator-install'], labels='payment')
k8s_resource('payment-migrate', resource_deps=['payment-db'], labels='payment')
k8s_resource('payment', port_forwards=['9096:9096'], resource_deps=['payment-db'], labels='payment')

### Payment Simulator Service ###
go_service('payment-gateway-simulator', './tools/payment-gateway-simulator', 8097)

k8s_resource('payment-gateway-simulator', port_forwards=['8097:8097'], resource_deps=['payment'], labels='payment')
