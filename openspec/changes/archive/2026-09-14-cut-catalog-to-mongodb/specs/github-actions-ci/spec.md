## MODIFIED Requirements

### Requirement: Shared dependency test fan-out

The CI workflow SHALL expand path filters so changes under shared modules trigger tests for dependent service modules according to this map:

- `shared/proto/**` → users, products, inventory, orders, cart, payment, web
- `shared/auth/**` → users, web
- `shared/messaging/**` → products, inventory, orders, payment
- `shared/err/dberr/**` → users, inventory, orders, payment
- `shared/err/grpcerr/**` → users, products, inventory, orders, cart, payment
- `shared/runtime/**` → users, products, inventory, orders, cart, payment, web
- `shared/observe/log/**` → users, products, inventory, orders, cart, payment, web
- `shared/observe/trace/**` → products, inventory, orders, payment, web
- `shared/testutil/postgres/**` → users, inventory, orders, payment
- `shared/testutil/kafka/**` → products, inventory, orders, payment
- `shared/testutil/redis/**` → cart
- `shared/testutil/mongo/**` → products

#### Scenario: Shared proto change

- **WHEN** a pull request modifies files under `shared/proto/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, payment, and web

#### Scenario: Shared messaging change

- **WHEN** a pull request modifies files under `shared/messaging/**`
- **THEN** CI runs tests for products, inventory, orders, and payment and does not run tests for cart solely due to that change

#### Scenario: Shared observe/trace change

- **WHEN** a pull request modifies files under `shared/observe/trace/**`
- **THEN** CI runs tests for products, inventory, orders, payment, and web

#### Scenario: Shared observe/log change

- **WHEN** a pull request modifies files under `shared/observe/log/**`
- **THEN** CI runs tests for users, products, inventory, orders, cart, payment, and web

#### Scenario: Shared mongo testutil change

- **WHEN** a pull request modifies files under `shared/testutil/mongo/**`
- **THEN** CI runs tests for products and does not run tests for unrelated services solely due to that change
