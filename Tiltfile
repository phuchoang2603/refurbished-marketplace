local_resource(
  'cnpg-operator-install',
  'helm repo add cnpg https://cloudnative-pg.github.io/charts && helm repo update && helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg',
)

k8s_yaml(helm('./infra/development/helm/refurbished-marketplace'))

### Users Service ###
local_resource(
  'users-compile',
  'mkdir -p build && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/users ./services/users/cmd/users',
  deps=['./services/users', './go.mod'], labels='compiles')

docker_build(
  'refurbished-marketplace/users',
  '.',
  dockerfile='./infra/development/docker/users.Dockerfile',
  only=[
    './build/users',
  ],
)

k8s_resource('users', port_forwards=['8081:8081'], resource_deps=['users-compile', 'cnpg-operator-install'], labels='services')
### End Users Service ###

### Products Service ###
local_resource(
  'products-compile',
  'mkdir -p build && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/products ./services/products/cmd/products',
  deps=['./services/products', './go.mod'], labels='compiles')

docker_build(
  'refurbished-marketplace/products',
  '.',
  dockerfile='./infra/development/docker/products.Dockerfile',
  only=[
    './build/products',
  ],
)

k8s_resource('products', port_forwards=['8082:8082'], resource_deps=['products-compile', 'cnpg-operator-install'], labels='services')
### End Products Service ###

### Orders Service ###
local_resource(
  'orders-compile',
  'mkdir -p build && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/orders ./services/orders/cmd/orders',
  deps=['./services/orders', './go.mod'], labels='compiles')

docker_build(
  'refurbished-marketplace/orders',
  '.',
  dockerfile='./infra/development/docker/orders.Dockerfile',
  only=[
    './build/orders',
  ],
)

k8s_resource('orders', port_forwards=['8083:8083'], resource_deps=['orders-compile', 'cnpg-operator-install'], labels='services')
### End Orders Service ###
