load('ext://namespace', 'namespace_create')
namespace_create('ecommerce')
namespace_create('operators')

### External Secrets Operators ###
local_resource(
  'eso-operator-install',
  'helm repo add external-secrets https://charts.external-secrets.io && \
   helm repo update && \
   helm upgrade --install external-secrets external-secrets/external-secrets \
     --namespace operators --create-namespace --version 2.6.0',
)

k8s_yaml([
  'infra/k8s/doppler-token.secret.yaml',
  'infra/k8s/cluster-secret-store.yaml',
])

### Cloudnative-pg Operators ###
k8s_kind('Cluster', pod_readiness='wait')
local_resource(
  'cnpg-operator-install',
  'helm repo add cnpg https://cloudnative-pg.io/charts/ && \
   helm repo update && \
   helm upgrade --install cnpg cnpg/cloudnative-pg \
   --namespace operators --create-namespace --version 0.28.3',
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
  'helm upgrade --install strimzi-cluster-operator oci://quay.io/strimzi-helm/strimzi-kafka-operator \
  --namespace operators --create-namespace \
  --set watchAnyNamespace=true \
  --version 1.0.0',
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
docker_build(
  'users-migrator',
  '.',
  dockerfile='./infra/docker/users-migrator.Dockerfile',
  only=[
    './services/users/db/migrations',
  ],
)

docker_build(
  'users',
  '.',
  dockerfile='./infra/docker/users.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('users-db', extra_pod_selectors=[{'cnpg.io/cluster': 'users-db'}], port_forwards=['5432:5432'], resource_deps=['cnpg-operator-install'], labels=['users'])
k8s_resource('users-migrate', resource_deps=['users-db'], labels='users')
k8s_resource('users', port_forwards=['9091:9091'], resource_deps=['users-db'], labels='users')

### Products Service ###
docker_build(
  'products-migrator',
  '.',
  dockerfile='./infra/docker/products-migrator.Dockerfile',
  only=[
    './services/products/db/migrations',
  ],
)

docker_build(
  'products',
  '.',
  dockerfile='./infra/docker/products.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('products-db', extra_pod_selectors=[{'cnpg.io/cluster': 'products-db'}], port_forwards=['5433:5432'], resource_deps=['cnpg-operator-install'], labels='products')
k8s_resource('products-migrate', resource_deps=['products-db'], labels='products')
k8s_resource('products', port_forwards=['9092:9092'], resource_deps=['products-db'], labels='products')

### Orders Service ###
docker_build(
  'orders-migrator',
  '.',
  dockerfile='./infra/docker/orders-migrator.Dockerfile',
  only=[
    './services/orders/db/migrations',
  ],
)

docker_build(
  'orders',
  '.',
  dockerfile='./infra/docker/orders.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('orders-db', extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}], port_forwards=['5434:5432'], resource_deps=['cnpg-operator-install'], labels='orders')
k8s_resource('orders-migrate', resource_deps=['orders-db'], labels='orders')
k8s_resource('orders', port_forwards=['9093:9093'], resource_deps=['orders-db'], labels='orders')

### Cart Service ###
docker_build(
  'cart',
  '.',
  dockerfile='./infra/docker/cart.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('cart', port_forwards=['9094:9094'],  labels='cart')

### Payment Service ###
docker_build(
  'payment-migrator',
  '.',
  dockerfile='./infra/docker/payment-migrator.Dockerfile',
  only=[
    './services/payment/db/migrations',
  ],
)

docker_build(
  'payment',
  '.',
  dockerfile='./infra/docker/payment.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('payment-db', extra_pod_selectors=[{'cnpg.io/cluster': 'payment-db'}], port_forwards=['5436:5432'], resource_deps=['cnpg-operator-install'], labels='payment')
k8s_resource('payment-migrate', resource_deps=['payment-db'], labels='payment')
k8s_resource('payment', port_forwards=['9096:9096'], resource_deps=['payment-db'], labels='payment')

### Payment Simulator Service ###
docker_build(
  'payment-gateway-simulator',
  '.',
  dockerfile='./infra/docker/payment-gateway-simulator.Dockerfile',
  only=[
    './go.work',
    './go.work.sum',
    './shared',
    './services',
    './tools',
  ],
)

k8s_resource('payment-gateway-simulator', port_forwards=['8097:8097'], resource_deps=['payment'], labels='payment')
