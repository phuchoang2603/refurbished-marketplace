## 1. Catalog Boundary Consolidation

- [x] 1.1 Extend `services/products` database ownership to include inventory, reservation, and supporting inbox/outbox tables needed by the unified catalog boundary.
- [x] 1.2 Move or recreate the inventory reservation service logic inside `services/products/internal` while keeping product, stock, and reservation code separated by module.
- [x] 1.3 Update service startup so the unified catalog runtime owns the existing reservation Kafka consumer behavior.

## 2. Data And API Unification

- [x] 2.1 Replace legacy products migrations with a fresh products-side schema baseline that includes inventory and reservation tables.
- [x] 2.2 Update the product-facing gRPC contract so detail/admin stock-aware reads come from the unified catalog service.
- [x] 2.3 Update the create-product path so product creation requires initial stock and initializes stock in one logical catalog operation.
- [x] 2.4 Keep list-oriented reads lighter than detail/admin stock reads in the first merged phase.

## 3. Runtime Cutover

- [x] 3.1 Update any order or payment integration points that currently assume reservation handling lives in a separate inventory runtime.
- [x] 3.2 Remove standalone inventory service deployment wiring after the unified catalog runtime owns inventory behavior.
- [x] 3.3 Defer `services/web` caller migration and capture any temporary compatibility needs without changing web code in this phase.

## 4. Verification

- [x] 4.1 Add or update tests for stock-aware product reads and unified listing creation with initial stock.
- [x] 4.2 Add or update tests for reservation behavior after it moves under the unified catalog boundary, including Kafka-driven order/payment handling.
- [x] 4.3 Add validation covering the fresh schema baseline, unified reservation behavior, and stock-aware catalog behavior.
- [x] 4.4 Run the relevant test suites for `services/products`, `services/orders`, `services/payment`, and shared contracts, addressing any failures needed for this change.
