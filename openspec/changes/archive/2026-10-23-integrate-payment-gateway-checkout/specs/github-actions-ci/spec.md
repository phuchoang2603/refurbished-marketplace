## MODIFIED Requirements

### Requirement: Shared dependency test fan-out

The CI workflow SHALL expand path filters so changes under shared modules trigger tests for dependent service modules according to this map:

- `shared/proto/**` → users, products, inventory, orders, cart, web, search
- `shared/auth/**` → users, web
- `shared/messaging/**` → products, inventory, orders, search
- `shared/err/dberr/**` → users, inventory, orders
- `shared/err/grpcerr/**` → users, products, inventory, orders, cart, search
- `shared/runtime/**` → users, products, inventory, orders, cart, web, search
- `shared/observe/log/**` → users, products, inventory, orders, cart, web, search
- `shared/observe/trace/**` → products, inventory, orders, web, search
- `shared/testutil/postgres/**` → users, inventory, orders
- `shared/testutil/kafka/**` → products, inventory, orders, search
- `shared/testutil/redis/**` → cart
- `shared/testutil/mongo/**` → products
- `shared/testutil/meilisearch/**` → search

#### Scenario: Shared proto change

- **WHEN** a pull request modifies files under `shared/proto/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, web, and search

#### Scenario: Shared messaging change

- **WHEN** a pull request modifies files under `shared/messaging/**`
- **THEN** CI runs tests for products, inventory, orders, and search and does not run tests for cart solely due to that change

#### Scenario: Shared observe/trace change

- **WHEN** a pull request modifies files under `shared/observe/trace/**`
- **THEN** CI runs tests for products, inventory, orders, web, and search

#### Scenario: Shared observe/log change

- **WHEN** a pull request modifies files under `shared/observe/log/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, web, and search

#### Scenario: Shared mongo testutil change

- **WHEN** a pull request modifies files under `shared/testutil/mongo/**`
- **THEN** CI runs tests for products and does not run tests for unrelated services solely due to that change

#### Scenario: Shared meilisearch testutil change

- **WHEN** a pull request modifies files under `shared/testutil/meilisearch/**`
- **THEN** CI runs tests for search and does not run tests for unrelated services solely due to that change

## ADDED Requirements

### Requirement: Release builds pin production chart image tags

After a successful image build, the release workflow SHALL write the built commit SHA into the prod marketplace and Kafka chart overlays and commit it, so production Argo CD applications roll to that SHA without a manual tag edit. It MUST NOT pin from forked pull requests and MUST skip the pin when the target branch head has moved past the built SHA.

#### Scenario: Images build on a branch push

- **WHEN** all images build successfully for a commit on a branch
- **THEN** the workflow commits the SHA into both prod overlay `global.imageTag` values with a skip-CI marker

#### Scenario: Branch moved during the build

- **WHEN** the branch head no longer matches the built SHA
- **THEN** the workflow leaves both overlays unchanged
