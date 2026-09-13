## Context

See proposal.md for motivation. Operators already live in `infra/charts/operators/` (ESO, CNPG, Strimzi) at sync wave `0`. Database CRs for Postgres live in the marketplace chart; Kafka CRs live in a dedicated chart destining `kafka`. Mongo must be a **replica set** in `ecommerce` so later catalog cutover (#57) and Meilisearch projection (#7) can use change streams. Cilium CNPs today only select marketplace `app:` labels; mongod will not be covered unless this change adds a policy. ClusterSecretStore `doppler` already exists.

## Goals / Non-Goals

**Goals:**

- Land MCK Community so a `MongoDBCommunity` CR is the source of mongod pods/PVCs, not a handmade StatefulSet.
- Keep shop traffic on Postgres; products MAY get unused `MONGO_ADDR`.
- Secrets and Cilium allow-list follow existing Doppler + CNP patterns.

**Non-Goals:**

- Application catalog writes, Go driver usage, or `wait-for-mongo` initContainers (those are #57).
- Ops Manager / Atlas / `spec.backup`.
- Teaching marketplace mesh-policy.tpl about Mongo (Mongo chart owns the CNP so operator labels stay local).

## Decisions

### 1. MCK Community, not Atlas, not homemade STS

Use [MongoDB Controllers for Kubernetes](https://github.com/mongodb/mongodb-kubernetes) Community (`MongoDBCommunity` CR) via Helm chart `mongodb/mongodb-kubernetes`. Do not install the deprecated `mongodb-kubernetes-operator` / `community-operator` chart.

**Rationale:** Same operator+CR shape as CNPG/Strimzi; replica-set bootstrap, SCRAM users, and version bumps are the operator’s job. Atlas Operator manages cloud Atlas, which this repo rejected. A raw STS is easier to get “Ready” while still being standalone.

**Alternatives considered:** Bitnami STS wrapper (arbiter/replicaCount traps, image policy); official `mongo` image + `rs.initiate` (we own keyfile forever); Atlas Operator (wrong data plane).

### 2. Two Argo apps: operator vs database CR

| App            | Path                              | Namespace   | Wave |
| -------------- | --------------------------------- | ----------- | ---- |
| MCK operator   | `infra/charts/operators/mongodb/` | `operators` | `0`  |
| Mongo database | `infra/charts/mongodb/`           | `ecommerce` | `2`  |

Wrapper operator chart depends on upstream `mongodb-kubernetes` (pin version in Chart.lock). Database chart templates: `MongoDBCommunity`, ExternalSecret(s), CiliumNetworkPolicy. Do **not** template a Namespace for `ecommerce`.

**Rationale:** Mirrors CNPG operator vs marketplace `Cluster` CRs, except Mongo CR is not stuffed into the marketplace chart so Mongo lifecycle is not tied to shop Deployments. Wave 2 is after ESO and before marketplace (3) so secrets and CRDs exist first.

**Alternatives considered:** CR inside marketplace chart (couples Mongo to shop sync); Mongo in its own namespace (worse CNP / DNS vs the epic’s `ecommerce` decision).

### 3. Operator watch scope: `ecommerce`

Set `operator.watchNamespace` to `ecommerce` so MCK does not reconcile random namespaces. If the chart cannot grant the extra Role/RoleBinding in `ecommerce` cleanly, fall back to `watchNamespace: "*"` (Strimzi already uses cluster-wide watch).

**Rationale:** Least privilege first; Strimzi-style `*` is an acceptable fallback, not the opening bid.

### 4. `members: 1` on talos-dev

Chart `values.yaml`: one replica-set member. `values-prod.yaml` MAY raise members/resources later without changing Community vs Enterprise.

**Rationale:** Change streams work on a 1-member replica set. Three mongod next to Kafka Connect on talos-dev is a RAM problem. Quorum HA is a prod overlay, not P0.

### 5. Doppler SCRAM user; CNP without SPIRE

ExternalSecret in the database chart maps Doppler keys (document as `MONGODB_APP_PASSWORD` and any admin password the CR needs) into the Secret keys `MongoDBCommunity` `users[].passwordSecretRef` expects. CNP: `enableDefaultDeny.ingress: true` when we follow marketplace enforce; allow `app: products` and `fromEntities: host` to port `27017`; **no** `authentication.mode: required`. Select **operator-generated** mongod labels (confirm from a rendered/sample CR; do not invent `app: mongodb`).

**Rationale:** Mongod is not SPIRE-enrolled (same family as Kafka TLS). Kubelet probes need `host` or the STS never becomes Ready.

### 6. Backup is not P0

Community MCK does not include Ops Manager continuous backup. Later options: CSI/Velero snapshots of operator PVCs, `mongodump`, or a separate Ops Manager epic.

**Rationale:** User chose MCK for operational headroom; P0 still must not stand up Ops Manager.

### 7. Unused `MONGO_ADDR` only

Marketplace `services.products.env.MONGO_ADDR` = in-cluster Service DNS (headless/service name the operator creates). No password in values. Products `LoadConfig` stays unchanged.

## Risks / Trade-offs

| Risk                                      | Mitigation                                                                           |
| ----------------------------------------- | ------------------------------------------------------------------------------------ |
| CRDs not ready when wave 2 applies the CR | Operator wave 0; Argo retry; document `kubectl get crd \| grep mongodb`              |
| CNP selector misses operator labels       | Apply CR in a dry-run/dev first; copy labels into the CNP; probe both allow and deny |
| CNP without `host`                        | Mongo never Ready; always allow kubelet                                              |
| Two Argo apps destine `ecommerce`         | Neither templates Namespace; each owns only its objects                              |
| 1-member RS is not HA                     | Accepted for talos-dev; prod overlay later                                           |
| Assuming MCK Community has `spec.backup`  | Document Community vs Ops Manager; do not enable backup fields                       |

## Migration Plan

1. Add Doppler keys on `dev`/`prd` before the database Application can become Ready.
2. Merge Git; point `dev-root` at the branch; wait for operator then `MongoDBCommunity` Running.
3. Verify `rs.status()`, empty change stream, CNP allow/deny, shop still on Postgres.
4. Rollback: disable/remove the two app-of-apps entries (or the database app first). PVCs may remain until pruned; shop does not use Mongo so rollback is Git-only for application traffic.

## Open Questions

- Exact upstream chart version and the precise `passwordSecretRef` key names — pin at implement time from the chart README.
- Whether namespaced watch needs extra RBAC templates vs `watchNamespace: "*"`.
