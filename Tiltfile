local_resource(
  'cnpg-operator-install',
  'helm repo add cnpg https://cloudnative-pg.github.io/charts && helm repo update && helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg',
)

k8s_yaml('./infra/development/k8s/secrets.yaml')
k8s_yaml(helm('./infra/chart', namespace='ecommerce', values=['./infra/development/k8s/dev-helm-values.yaml']))

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
    './go.mod',
    './go.sum',
  ],
)

k8s_resource('users-migrate', resource_deps=['cnpg-operator-install'], labels='migrate')
k8s_resource('users', port_forwards=['9091:9091'], resource_deps=['cnpg-operator-install'], labels='services')
### End Users Service ###

### Web Service ###
docker_build(
  'refurbished-marketplace/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
  only=[
    './services/web',
    './services/users/proto',
    './go.mod',
    './go.sum',
  ],
)

k8s_resource('web', port_forwards=['8080:8080'], resource_deps=['cnpg-operator-install'], labels='services')
### End Web Service ###

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
