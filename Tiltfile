local_resource(
  'cnpg-operator-install',
  'helm repo add cnpg https://cloudnative-pg.github.io/charts && helm repo update && helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg',
)

k8s_yaml('./infra/development/helm/secrets.yaml')

k8s_yaml(helm('./infra/development/helm/refurbished-marketplace'))

### Users Service ###
docker_build(
  'refurbished-marketplace/users',
  '.',
  dockerfile='./infra/development/docker/users.Dockerfile',
  only=[
    './services/users',
    './go.mod',
    './go.sum',
  ],
)

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

k8s_resource('users', port_forwards=['8081:8081'], resource_deps=['cnpg-operator-install'], labels='services')
### End Users Service ###

### Products Service ###
docker_build(
  'refurbished-marketplace/products',
  '.',
  dockerfile='./infra/development/docker/products.Dockerfile',
  only=[
    './services/products',
    './go.mod',
    './go.sum',
  ],
)

k8s_resource('products', port_forwards=['8082:8082'], resource_deps=['cnpg-operator-install'], labels='services')
### End Products Service ###

### Orders Service ###
docker_build(
  'refurbished-marketplace/orders',
  '.',
  dockerfile='./infra/development/docker/orders.Dockerfile',
  only=[
    './services/orders',
    './go.mod',
    './go.sum',
  ],
)

k8s_resource('orders', port_forwards=['8083:8083'], resource_deps=['cnpg-operator-install'], labels='services')
### End Orders Service ###
