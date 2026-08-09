## Context

Today:

| Path                                          | Flow                                         | Cost                   |
| --------------------------------------------- | -------------------------------------------- | ---------------------- |
| `GET /cart`, qty/remove fragment              | `GetCart` + N× `GetProductByID`              | 1+N                    |
| `POST /cart/checkout` (k lines same merchant) | N× `GetProductByID` + N× `RemoveCartItem`    | 2N (+ order + payment) |
| Cart Redis                                    | `{ product_id, merchant_id, quantity }` only | no display facts       |

Issue #33 direction (already agreed):

1. **Replication** for cart display (stamp name/price on lines at add/set)
2. **Re-validate at checkout** with **one** batch products call
3. **Multi-remove** — implementer’s choice → **batch cart RPC**

### Current vs target RPC shape

```
TODAY — cart paint                          TARGET — cart paint
─────────────────────────                   ─────────────────────────
Web ──GetCart──► Cart                       Web ──GetCart──► Cart
  │ for each line  (N)                        │ name/price from item
  └──GetProductByID──► Products               └── (no Products)


TODAY — checkout (k items)                  TARGET — checkout (k items)
──────────────────────────                  ──────────────────────────
Web ──GetCart──► Cart                       Web ──GetCart──► Cart
  │ k× GetProductByID ► Products              │ 1× GetProductsByIDs ► Products
  │ CreateOrder ► Orders                      │ CreateOrder ► Orders
  │ k× RemoveCartItem ► Cart                  │ 1× RemoveCartItems ► Cart
  └── CreateHostedPaymentSession              └── CreateHostedPaymentSession


TARGET — add item
─────────────────
Web ──GetProductByID(1 sku)──► Products
  └── AddCartItem(+ name, unit_price_cents) ──► Cart
```

## Goals / Non-Goals

**Goals:**

- Cart display cost scales as **O(1) inter-service RPCs** (single `GetCart`)
- Checkout product authority scales as **O(1) products RPC** for all lines in the merchant group
- Checkout cart cleanup scales as **O(1) cart RPCs** for all removed product IDs
- Money and merchant checks on checkout use **Postgres-backed** product reads (batch)
- Cart service stays a pure Redis session store (no catalog coupling)

**Non-Goals:** see proposal.

## Decisions

### 1. Snapshot fields (minimal)

Extend `CartItem` (proto + Redis JSON + service struct) with:

| Field              | Role                            |
| ------------------ | ------------------------------- |
| `product_name`     | Display                         |
| `unit_price_cents` | Estimated line total / UI price |

Do **not** snapshot description, stock, or product `updated_at` in v1.

**Rationale:** Matches `CartItemView` needs (`ProductName`, `ProductPrice`, line total). Keeps Redis doc small. Merchant still from existing `merchant_id` (also re-checked at checkout).

**Merge on add qty:** when line already exists, increment quantity **and** refresh `product_name` + `unit_price_cents` from the request stamp (latest write wins). Same for Set.

### 2. Who stamps the snapshot — web BFF, not cart

On `POST /cart/items` and quantity set:

1. `Products.GetProductByID(product_id)` (1 RPC — unavoidable for that SKU write)
2. Reject if NotFound / wrong merchant vs form `merchant_id`
3. `Cart.AddCartItem` / `SetCartItemQuantity` with name + price + merchant

Cart validates UUIDs, qty, **and** snapshot: non-empty `product_name` and `unit_price_cents > 0` (InvalidArgument). Incomplete stamps fail closed at the cart boundary.

**Rationale:** Avoids products client + mesh dep on cart pods; composition stays at the edge where both clients already exist.

**Alternatives considered:**

| Option                                        | Why not (v1)                                                         |
| --------------------------------------------- | -------------------------------------------------------------------- |
| Cart calls Products on GetCart                | Reintroduces N dependency + coupling; defeats “cart free of catalog” |
| Stamp-only at Add, never refresh on Set       | Stale name/price on qty path forever until re-add                    |
| Web-only cache map by product_id outside cart | Extra store; cart multi-device/cookie restart loses it               |

### 3. Cart paint uses snapshots only

`mapCartView`:

- No Products client loop
- `ProductName` / `ProductPrice` / `LineTotalCents` from snapshot
- `Available = true` iff snapshot present (`product_name != ""` and `unit_price_cents > 0`); else treat like incomplete (same UI stance as previous NotFound: line shows without nice name/price)
- Cart page no longer fails entirely when products is down (big reliability win for abandoned sessions)

### 4. `GetProductsByIDs` — Postgres multi-get

```proto
rpc GetProductsByIDs(GetProductsByIDsRequest) returns (GetProductsByIDsResponse);
// Request: repeated string ids
// Response: repeated Product products  // found only; missing IDs omitted
```

- sqlc: `WHERE id = ANY($1::uuid[])` join inventory like `GetProductByID` (return stock fields free)
- Cap input size at **100** IDs; over limit → InvalidArgument
- Empty `ids` → empty product list (valid no-op)
- Dedup ids server-side preferred (implementation detail)

**Checkout:**

1. Collect selected merchant group product IDs from cart
2. One `GetProductsByIDs`
3. Build map; for each selected line: missing → error (fail closed); product.merchant_id ≠ selected merchant → conflict; build `CreateOrderItem` with **product.PriceCents** (not cart snapshot)

**Price lag policy v1:** snapshot may differ; order always uses batch SoR price. No user toast required for #33.

**Alternatives considered:** N parallel GetProductByID (still N RPC frames); Meilisearch multi-get (wrong for money path); only use cart prices at checkout (unsafe).

### 5. `RemoveCartItems` — single Redis RMW

```proto
rpc RemoveCartItems(RemoveCartItemsRequest) returns (Cart);
// cart_id + repeated product_ids
```

Semantics:

- Validate cart_id; each product_id UUID
- Load cart; if missing key → NotFound (or empty cart policy consistent with single remove)
- Drop every item whose product_id ∈ set (set membership O(1))
- **Missing product_ids in cart are ignored** (idempotent multi-remove) — better than failing after CreateOrder on partial double-submit
- Save once; return cart
- Empty `product_ids` → InvalidArgument
- If result empty after removes, web still clears cart cookie when appropriate (existing behavior)
- **One remove API only:** no `RemoveCartItem`. UI and qty-zero call `RemoveCartItems` with a one-element `product_ids` list. Missing IDs always ignored.

**Why multi-remove vs ClearCart for partial merchant checkout:** cart may still hold other merchants’ lines; ClearCart would wipe them. Multi by product_id is correct.

**Why not rebuild-write from web without new RPC:** possible (`GetCart` + filter in web + hypothetical `ReplaceCart`) but invents Replace and races worse; multi-remove is local, one atomic save.

### 6. Legacy Redis carts

JSON without new keys → Go zero values → incomplete snapshot → UI degrades per line. **No migration job.** Add path re-stamps when user re-adds or adjusts qty (Set stamps again).

### 7. Proto / API surface (concrete sketch)

**cart.proto CartItem / write RPCs:**

```text
message CartItem {
  string product_id = 1;
  int32 quantity = 2;
  string merchant_id = 3;
  string product_name = 4;       // NEW
  int64 unit_price_cents = 5;    // NEW
}

message AddCartItemRequest {
  // existing +
  string product_name = 5;
  int64 unit_price_cents = 6;
}
// SetCartItemQuantityRequest same extras
```

Remove: `RemoveCartItemsRequest { cart_id, repeated product_ids }`.

**products:** request/response as above field numbers in real proto as next free slots on service.

### 8. Checkout failure ordering (existing risk, not expanded)

Order today: validate → CreateOrder → remove cart items → payment session.

If remove fails after CreateOrder, buyer can have order + items still in cart (double-checkout risk) — **pre-existing**. Out of #33 scope to invent transactional saga; multi-remove only reduces N partial failure surface to one RMW.

## File / ownership map

| Area                                                     | Touch                                                         |
| -------------------------------------------------------- | ------------------------------------------------------------- |
| `shared/proto/cart/v1/cart.proto`                        | CartItem, Add/Set, RemoveCartItems                            |
| `shared/proto/products/v1/products.proto`                | GetProductsByIDs                                              |
| `services/cart/...`                                      | CartItem fields, Add/Set signatures, RemoveCartItems, mapCart |
| `services/products/db/queries` + sqlc                    | GetProductsByIDs                                              |
| `services/products/internal/service` + grpc              | multi-get                                                     |
| `services/web/internal/clients` + `handlers/shared/deps` | client methods + interfaces                                   |
| `services/web/internal/handlers/cart`                    | add stamp, mapCartView, checkout batch + multi-remove         |
| `services/web/tests/fakes`                               | new methods                                                   |
| `generate-proto` / module tidy                           | codegen                                                       |

## Risks / Trade-offs

| Risk                             | Mitigation                                                                                  |
| -------------------------------- | ------------------------------------------------------------------------------------------- |
| Stale cart UI prices vs checkout | Documented; money path re-validates                                                         |
| Form merchant_id spoof on add    | Already present; stamp after GetProduct and require product.merchant_id == form merchant_id |
| Giant batch                      | Cap ID list; cart sizes tiny in practice                                                    |
| Incomplete old carts             | Degraded UI; Set/Add re-stamps                                                              |
| Add slower by +1 product RPC     | Acceptable; replaces N future Gets                                                          |

## Migration / roll-out

1. Proto + services (backward-incompatible for add without snapshot is fine — no external clients expected; web is sole caller)
2. Deploy cart, products, web together (Tilt/rolling: web last after cart+products understand new fields/RPCs)
3. No Redis schema migration needed (JSON flexible)

## Open Questions

None blocking apply — decisions pinned above (web stamps, batch SoR at checkout, idempotent multi-remove, fail closed on missing checkout product, cart write requires snapshot, max batch 100).
