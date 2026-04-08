# Create namespace if not exist
load('ext://namespace', 'namespace_create')
namespace_create('ecommerce')
namespace_create('cnpg-system')
namespace_create('kafka-system')

# Deploy cnpg operator
k8s_kind('Cluster', pod_readiness='wait')
local_resource(
  'cnpg-operator-install',
  'helm repo add cnpg https://cloudnative-pg.github.io/charts && helm repo update && helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg',
  )

# Deploy Kafka cluster
local_resource(
  'kafka-cluster-install',
  'helm upgrade --install strimzi-cluster-operator oci://quay.io/strimzi-helm/strimzi-kafka-operator --namespace kafka-system --create-namespace --set watchAnyNamespace=true',
)

# Deploy our application
k8s_yaml('./infra/development/k8s/secrets.yaml')
app_yaml = helm(
    './infra/charts/refurbished-marketplace',
    name='refurbished-marketplace',
    namespace='ecommerce',
    values=['./infra/development/k8s/dev-helm-values.yaml']
)
k8s_yaml(app_yaml)

### Web Service ###
docker_build(
  'refurbished-marketplace/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
  only=[
    './shared',
    './services/web',
  ],
)

k8s_resource('web', port_forwards=['8080:8080'])

### Users Service ###
docker_build(
  'refurbished-marketplace/users-migrator',
  '.',
  dockerfile='./infra/development/docker/users-migrator.Dockerfile',
  only=[
    './services/users/db/migrations',
  ],
)

docker_build(
  'refurbished-marketplace/users',
  '.',
  dockerfile='./infra/development/docker/users.Dockerfile',
  only=[
    './services/users',
    './shared',
  ],
)

k8s_resource('users-db', extra_pod_selectors=[{'cnpg.io/cluster': 'users-db'}], port_forwards=['5432:5432'], resource_deps=['cnpg-operator-install'], labels=['users'])
k8s_resource('users-migrate', resource_deps=['users-db'], labels='users')
k8s_resource('users', port_forwards=['9091:9091'], resource_deps=['users-db'], labels='users')

### Products Service ###
docker_build(
  'refurbished-marketplace/products-migrator',
  '.',
  dockerfile='./infra/development/docker/products-migrator.Dockerfile',
  only=[
    './services/products/db/migrations',
  ],
)

docker_build(
  'refurbished-marketplace/products',
  '.',
  dockerfile='./infra/development/docker/products.Dockerfile',
  only=[
    './services/products',
    './shared',
  ],
)

k8s_resource('products-db', extra_pod_selectors=[{'cnpg.io/cluster': 'products-db'}], port_forwards=['5433:5432'], resource_deps=['cnpg-operator-install'], labels='products')
k8s_resource('products-migrate', resource_deps=['products-db'], labels='products')
k8s_resource('products', port_forwards=['9092:9092'], resource_deps=['products-db'], labels='products')

### Orders Service ###
docker_build(
  'refurbished-marketplace/orders-migrator',
  '.',
  dockerfile='./infra/development/docker/orders-migrator.Dockerfile',
  only=[
    './services/orders/db/migrations',
  ],
)

docker_build(
  'refurbished-marketplace/orders',
  '.',
  dockerfile='./infra/development/docker/orders.Dockerfile',
  only=[
    './services/orders',
    './shared',
  ],
)

k8s_resource('orders-db', extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}], port_forwards=['5434:5432'], resource_deps=['cnpg-operator-install'], labels='orders')
k8s_resource('orders-migrate', resource_deps=['orders-db'], labels='orders')
k8s_resource('orders', port_forwards=['9093:9093'], resource_deps=['orders-db'], labels='orders')

### Cart Service ###
docker_build(
  'refurbished-marketplace/cart',
  '.',
  dockerfile='./infra/development/docker/cart.Dockerfile',
  only=[
    './services/cart',
    './shared',
  ],
)

k8s_resource('cart', port_forwards=['9094:9094'],  labels='cart')

### Inventory Service ###
docker_build(
  'refurbished-marketplace/inventory-migrator',
  '.',
  dockerfile='./infra/development/docker/inventory-migrator.Dockerfile',
  only=[
    './services/inventory/db/migrations',
  ],
)

docker_build(
  'refurbished-marketplace/inventory',
  '.',
  dockerfile='./infra/development/docker/inventory.Dockerfile',
  only=[
    './services/inventory',
    './shared',
  ],
)

k8s_resource('inventory-db', extra_pod_selectors=[{'cnpg.io/cluster': 'inventory-db'}], port_forwards=['5435:5432'], resource_deps=['cnpg-operator-install'], labels='inventory')
k8s_resource('inventory-migrate', resource_deps=['inventory-db'], labels='inventory')
k8s_resource('inventory', port_forwards=['9095:9095'], resource_deps=['inventory-db'], labels='inventory')

### Kafka Cluster and Topics ###
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
