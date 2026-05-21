## Context

The product domain has now been collapsed into the products/catalog boundary. Product CRUD, stock rows, reservation rows, and Kafka consumers for order/payment outcomes all live under `services/products`. The web edge is intentionally left untouched in this phase, but the standalone inventory runtime and its separate live contract surface are no longer part of the steady-state architecture.

Two repo details make a merge particularly reasonable here. First, the products schema originally had a `stock` column and later removed it, so reunifying stock with catalog is partly reversing an earlier split rather than inventing a new model from scratch. Second, the truly important complexity is not "inventory as a separate network service" but reservation correctness, and that complexity can remain isolated in its own module and tables inside a unified catalog boundary.

## Goals / Non-Goals

**Goals:**

- Collapse catalog, stock, and reservation persistence into one service boundary and one database boundary.
- Support stock-aware product reads without requiring downstream fan-out, while limiting exact stock exposure to detail and admin-oriented surfaces in the first phase.
- Support product creation with required initial stock in one logical write path instead of a product-first then inventory-later flow.
- Preserve reservation logic and Kafka-driven order/payment handling as a distinct internal concern rather than flattening everything into the `products` table.
- Remove the need for `product.created` inventory bootstrap for this domain boundary.

**Non-Goals:**

- Replacing PostgreSQL with Elasticsearch or another search-oriented primary store.
- Flattening reservation state into product rows or removing reservation tables.
- Changing order, payment, or checkout semantics beyond moving inventory behavior behind the unified catalog boundary.
- Designing a generalized event platform refactor for every service.
- Updating `services/web` to consume the merged API in this first phase.

## Decisions

### Decision: Use the current products service as the unified catalog boundary

The merged catalog boundary lives in `services/products`, which owns inventory tables, reservation logic, and Kafka consumers as part of its steady-state runtime.

Rationale:

- It already owns the listing-facing gRPC surface used by the web edge.
- The products schema historically included stock, so the boundary has precedent in the repo.
- It keeps the merged boundary simple by removing the old inventory runtime instead of preserving a transitional split.

Alternatives considered:

- Use the current inventory service as the base and absorb product CRUD into it. Rejected because product reads and writes are already the primary caller-facing surface.
- Create a brand-new `services/catalog` runtime first. Rejected for now because it adds path churn before the domain shape is settled.

### Decision: Keep separate tables for products, inventory, and reservations inside one database

The unified service will not re-add a single `stock` column to `products`. Instead, it will keep `products`, `inventory`, and `inventory_reservations` as separate tables under one service and one database.

Rationale:

- It preserves the useful separation between listing data, stock totals, and reservation records.
- It keeps reservation mutation logic explicit and compatible with current order/payment workflows.
- It still removes the network boundary that is causing the current awkwardness.

Alternatives considered:

- Merge stock directly back into the `products` table. Rejected because available and reserved quantities plus reservation lifecycle are already richer than a single stock column.

### Decision: Move to one stock-aware product API surface and remove standalone inventory APIs

The unified service exposes one caller-facing catalog surface for product reads and creation, with product reads able to return stock summary on detail and admin-oriented surfaces and product creation requiring initial stock in the same logical operation. The standalone `InventoryService` surface is removed from the live codebase.

Rationale:

- This is what actually removes the web and admin awkwardness the user is reacting to.
- It replaces two downstream calls and two write steps with one read path and one write path.
- Requiring initial stock keeps listing creation explicit and avoids hidden defaults at the new API boundary.

Alternatives considered:

- Keep separate `ProductsService` and `InventoryService` contracts on the same runtime indefinitely. Rejected because it preserves the split mentally and in callers even after the runtime merge.

### Decision: Defer web-service migration until after the service merge is established

This phase focuses on the merged service boundary and internal APIs. `services/web` is intentionally not updated yet.

Rationale:

- It keeps the migration focused on the core domain consolidation.
- It reduces the blast radius while database, runtime, and internal API changes are still settling.

Alternatives considered:

- Migrate web in the same change. Rejected because it mixes caller adaptation with the more fundamental service-boundary change.

### Decision: Keep reservation event handling, but move it into the unified catalog runtime

The Kafka consumer logic that currently lives in inventory for `orders.created` and payment outcome handling will move with reservation behavior into the unified runtime.

Rationale:

- Reservation correctness still matters even if the service split goes away.
- Moving the logic without changing the behavior keeps buyer/order flows stable while simplifying service ownership.

Alternatives considered:

- Remove Kafka and make reservation handling synchronous with orders immediately. Rejected because that is a separate architectural decision from the product/inventory merge.

### Decision: Do not introduce `product.created` outbox bootstrap for inventory

The unified catalog boundary will make product creation plus stock initialization a single service concern, so this change intentionally replaces the async bootstrap direction.

Rationale:

- It removes eventual-consistency delay from listing creation.
- It avoids introducing Kafka machinery only to stitch together a domain that is being merged anyway.

Alternatives considered:

- Keep the split and add `product.created` with an outbox. Rejected because it optimizes the coupling instead of removing it.

## Risks / Trade-offs

- [Combining runtimes can create a larger service with mixed concerns] -> Mitigate by preserving internal module boundaries for product, stock, and reservation code even though the network boundary disappears.
- [Callers may still assume the old inventory runtime exists] -> Mitigate by planning caller migration separately and keeping this change focused on the merged service boundary.
- [Deferring web migration means downstream callers will lag the new contract] -> Mitigate by treating this phase as service consolidation only and planning caller migration separately once the merged boundary is stable.
- [Reservation Kafka handling becomes coupled to the products deployment lifecycle] -> Mitigate by keeping the consumer startup isolated within the unified runtime and preserving current topic semantics.

## Migration Plan

1. Establish a fresh products-side schema baseline that includes inventory and reservation tables plus inbox/outbox support needed for reservation handling.
2. Move reservation Kafka consumers and supporting inventory logic into the unified runtime.
3. Update product-facing gRPC contracts so detail/admin reads can include stock summary and create paths require initial stock in one operation.
4. Remove the standalone inventory runtime and replace live inventory ownership with the unified products runtime.

Rollback strategy:

- Because this implementation uses a fresh baseline rather than a legacy data migration path, rollback is primarily code and deployment rollback within the merged products runtime.
- Caller migration remains a separate concern and should be coordinated independently.

## Open Questions

- None for this phase.
