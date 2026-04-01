# Create namespace if not exist
load('ext://namespace', 'namespace_create')
namespace_create('ecommerce')
namespace_create('cnpg-system')

# Deploy cnpg operator
load('ext://helm_resource', 'helm_resource', 'helm_repo')
k8s_kind('Cluster', pod_readiness='wait')
helm_repo('cnpg-repo', 'https://cloudnative-pg.github.io/charts')
helm_resource(
    'cnpg-operator-install',
    'cnpg-repo/cloudnative-pg', 
    namespace='cnpg-system',
)

# Deploy everything
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
    './services/web',
    './shared',
    './go.mod',
    './go.sum',
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
    './go.mod',
    './go.sum',
  ],
)

docker_build(
  'refurbished-marketplace/users',
  '.',
  dockerfile='./infra/development/docker/users.Dockerfile',
  only=[
    './services/users',
    './shared',
    './go.mod',
    './go.sum',
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
    './go.mod',
    './go.sum',
  ],
)

docker_build(
  'refurbished-marketplace/products',
  '.',
  dockerfile='./infra/development/docker/products.Dockerfile',
  only=[
    './services/products',
    './shared',
    './go.mod',
    './go.sum',
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
    './go.mod',
    './go.sum',
  ],
)

docker_build(
  'refurbished-marketplace/orders',
  '.',
  dockerfile='./infra/development/docker/orders.Dockerfile',
  only=[
    './services/orders',
    './shared',
    './go.mod',
    './go.sum',
  ],
)

k8s_resource('orders-db', extra_pod_selectors=[{'cnpg.io/cluster': 'orders-db'}], port_forwards=['5434:5432'], resource_deps=['cnpg-operator-install'], labels='orders')
k8s_resource('orders-migrate', resource_deps=['orders-db'], labels='orders')
k8s_resource('orders', port_forwards=['9093:9093'], resource_deps=['orders-db'], labels='orders')
