# Catalog, Inventory, And Search

How listings, stock, and storefront search relate after inventory is its own runtime. Meilisearch (#7) is the list/search projection; it is not source of truth.

## Stores

| Store                            | Owns                                                                | Readers                                   |
| -------------------------------- | ------------------------------------------------------------------- | ----------------------------------------- |
| Products (SQL now, Mongo in #57) | Listing document: name, description, price, merchant                | Web Get*, checkout price re-validation    |
| Inventory Postgres               | `available_qty`, `reserved_qty`, reservations, outbox               | Web EnsureStock / GetStock / ReserveStock |
| Cart Redis                       | Line snapshots (name, unit price, qty of _items_)                   | Cart GET                                  |
| Meilisearch (#7)                 | Search/list projection of listings (+ optional coarse availability) | Storefront browse/search                  |

Products **never** gRPC inventory.

## Target read model (scale)

Inventory write-DB stays authoritative for checkout. After qty changes, an `InventoryUpdated` (or existing inventory outbox) event updates a **read-optimized** product view (Mongo listing field and/or Meili). The product page and search read that view. Checkout still calls inventory `ReserveStock`.

Until that projector exists, web MAY `GetStock` **once** on the product detail page. Add-to-cart copies the PDP snapshot (name, price). Cart HTML does not call products or inventory per line.

```mermaid
flowchart TB
  subgraph write["Write / strong"]
    Web[web BFF]
    Products[products]
    Inv[inventory]
    Cat[(catalog SoT)]
    Led[(inventory_db)]
    Web -->|"CreateProduct"| Products
    Products --> Cat
    Web -->|"EnsureStock"| Inv
    Web -->|"ReserveStock"| Inv
    Inv --> Led
  end

  subgraph events["Async"]
    K[Kafka]
    Inv -->|"inventory.reserved / qty changed"| K
    Cat -.->|change stream #7| Proj[projector]
    K --> Proj
  end

  subgraph read["Read / eventual"]
    Meili[(Meilisearch)]
    View[PDP / search view]
    Proj --> Meili
    Proj --> View
    Web -->|"GET /products/{id} later"| View
    Web -->|"SearchProducts #7"| Meili
    Web -->|"GET /cart"| Redis[(cart snapshots)]
  end
```

## Create listing

```mermaid
sequenceDiagram
  participant B as Browser
  participant W as Web
  participant P as Products
  participant I as Inventory

  B->>W: POST create listing + initial qty
  W->>P: CreateProduct catalog only
  P-->>W: product id
  W->>I: EnsureStock id, qty
  alt EnsureStock fails
    W->>P: DeleteProduct
    W-->>B: error
  else success
    W-->>B: success
  end
```

## View, cart, checkout

```mermaid
sequenceDiagram
  participant B as Browser
  participant W as Web
  participant P as Products
  participant I as Inventory
  participant C as Cart Redis

  B->>W: GET /products/id
  W->>P: GetProductByID
  opt until read model exists
    W->>I: GetStock
  end
  W-->>B: PDP HTML

  B->>W: POST add to cart snapshot
  W->>C: AddCartItem name+price from form
  Note over W,C: no GetProduct, no inventory

  B->>W: GET /cart
  W->>C: GetCart
  W-->>B: HTML from snapshots

  B->>W: checkout
  W->>P: GetProductsByIDs prices
  W->>I: ReserveStock
```

## Meilisearch (#7, not this extract)

- Index from **Mongo listing** change stream (after #57), not from Postgres `products`.
- Optional: same projector applies `InventoryUpdated` as `in_stock` / `low` / `out` (not the reservation ledger).
- `ListProducts` SQL OFFSET goes away; browse stays dark until this lands.

## What not to do

- Products calling inventory on Get* or Create.
- Treating Meili or the PDP cache as stock SoT at checkout.
- Kafka between CreateProduct and EnsureStock (seller create stays sync in web).
