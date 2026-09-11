## 1. Operator chart

- [x] 1.1 Add `infra/charts/operators/mongodb/` wrapper around upstream Helm `mongodb/mongodb-kubernetes` (pin version in Chart.lock; not the deprecated community-operator chart)
- [x] 1.2 Set operator namespace values for `operators` and `watchNamespace: ecommerce` (fall back to `*` only if namespaced RBAC cannot be granted)
- [x] 1.3 Confirm CRDs for `MongoDBCommunity` are installed by the chart

## 2. Database chart

- [x] 2.1 Add `infra/charts/mongodb/` with `MongoDBCommunity` (`members: 1`, Community replica set, SCRAM user `passwordSecretRef`)
- [x] 2.2 Add ExternalSecret(s) from ClusterSecretStore `doppler` into the Secret keys the CR expects
- [x] 2.3 Add CiliumNetworkPolicy selecting operator mongod labels: allow `app: products` and `fromEntities: host` to 27017; default-deny ingress; no `authentication.mode: required`
- [x] 2.4 Add `values-prod.yaml` for optional larger resources/members; do not template an `ecommerce` Namespace

## 3. App-of-apps and marketplace placeholder

- [x] 3.1 Register operator app (wave `0`, ns `operators`) and database app (wave `2`, ns `ecommerce`) in `infra/argocd/app-of-apps/values.yaml`
- [x] 3.2 Wire prod-root `valueFiles` for the Mongo database chart if prod overlay exists
- [x] 3.3 Add unused `MONGO_ADDR` on products in marketplace `values.yaml` pointing at the operator Service DNS (no password)

## 4. Docs and Doppler

- [x] 4.1 Document Doppler keys and Kubernetes Secret names in `docs/development/secrets.md`
- [x] 4.2 Document both Applications, namespaces, sync waves, Community-not-Ops-Manager backup, and shop independence in `docs/deployment/gitops.md`

## 5. Verify on talos-dev

- [ ] 5.1 Operator Application Healthy; `MongoDBCommunity` Running; `rs.status()` is a replica set
- [ ] 5.2 Empty change stream opens; products-path probe connects; unenrolled probe fails
- [x] 5.3 Shop browse/checkout still works without a Mongo driver
