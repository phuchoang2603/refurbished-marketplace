## MODIFIED Requirements

### Requirement: Shared dependency test fan-out

The CI workflow SHALL expand path filters so changes under shared modules trigger tests for dependent service modules according to this map:

- `shared/proto/**` → users, products, inventory, orders, cart, payment, web, search
- `shared/auth/**` → users, web
- `shared/messaging/**` → products, inventory, orders, payment, search
- `shared/err/dberr/**` → users, inventory, orders, payment
- `shared/err/grpcerr/**` → users, products, inventory, orders, cart, payment, search
- `shared/runtime/**` → users, products, inventory, orders, cart, payment, web, search
- `shared/observe/log/**` → users, products, inventory, orders, cart, payment, web, search
- `shared/observe/trace/**` → products, inventory, orders, payment, web, search
- `shared/testutil/postgres/**` → users, inventory, orders, payment
- `shared/testutil/kafka/**` → products, inventory, orders, payment, search
- `shared/testutil/redis/**` → cart
- `shared/testutil/mongo/**` → products
- `shared/testutil/meilisearch/**` → search

#### Scenario: Shared proto change

- **WHEN** a pull request modifies files under `shared/proto/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, payment, web, and search

#### Scenario: Shared messaging change

- **WHEN** a pull request modifies files under `shared/messaging/**`
- **THEN** CI runs tests for products, inventory, orders, payment, and search and does not run tests for cart solely due to that change

#### Scenario: Shared observe/trace change

- **WHEN** a pull request modifies files under `shared/observe/trace/**`
- **THEN** CI runs tests for products, inventory, orders, payment, web, and search

#### Scenario: Shared observe/log change

- **WHEN** a pull request modifies files under `shared/observe/log/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, payment, web, and search

#### Scenario: Shared mongo testutil change

- **WHEN** a pull request modifies files under `shared/testutil/mongo/**`
- **THEN** CI runs tests for products and does not run tests for unrelated services solely due to that change

#### Scenario: Shared meilisearch testutil change

- **WHEN** a pull request modifies files under `shared/testutil/meilisearch/**`
- **THEN** CI runs tests for search and does not run tests for unrelated services solely due to that change

## ADDED Requirements

### Requirement: Search participates in service CI

CI SHALL include search in the lint module list, service path filters and outputs, test matrix, and vulnerability-scan matrix using the same triggers and shared dependency rules as other Go services.

#### Scenario: Search changes are checked

- **WHEN** search code or its shared dependencies change in a pull request
- **THEN** search lint, tests, and vulnerability scanning SHALL run

#### Scenario: Scheduled vulnerability scan includes search

- **WHEN** the scheduled full vulnerability scan runs
- **THEN** search SHALL be scanned alongside the other services
