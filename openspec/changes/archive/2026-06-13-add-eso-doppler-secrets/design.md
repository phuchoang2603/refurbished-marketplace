## Context

Tilt previously applied `infra/k8s/secrets.yaml` with plaintext DB passwords and `JWT_SECRET`. Helm charts reference those Secret names via `secretKeyRef`, CNPG `passwordSecret`, and Strimzi `${secrets:namespace/name:key}` for Debezium. GHCR release workflow is done; ArgoCD (#4) needs secrets out of Git.

Doppler provides separate **configs** (`dev`, `prd`, …). ESO's [Doppler provider](https://external-secrets.io/latest/provider/doppler/) syncs into native Kubernetes Secrets. Service tokens ([Doppler docs](https://docs.doppler.com/docs/service-tokens)) are scoped to one config — ideal for ESO auth.

## Goals / Non-Goals

**Goals:**

- Cluster-level ESO auth in `infra/k8s/`; application ExternalSecrets co-located with the marketplace Helm chart
- Doppler `dev` config for local; `prd` for remote — different service tokens, same manifest shape
- Bootstrap token via devenv `.env` (`DOPPLER_TOKEN`) + `files` YAML — no token in Git
- Preserve existing K8s Secret names and keys expected by charts
- Derive DB ExternalSecrets from `services.<slug>.db.secretName` with a shared basic-auth template
- Provider-agnostic via ESO `ClusterSecretStore` (swap Doppler → Vault/AWS later)

**Non-goals:**

- Standalone `infra/eso/` directory of flat ExternalSecret YAML files
- ArgoCD Application for ESO in this change (upstream Helm via Tilt first)
- Automating Doppler seeding in repo scripts — document manual seeding from dev values

## Decisions

### Split manifests: `infra/k8s/` + marketplace Helm chart

| Layer           | Location                                | Contents                                                  |
| --------------- | --------------------------------------- | --------------------------------------------------------- |
| Cluster auth    | `infra/k8s/cluster-secret-store.yaml`   | `ClusterSecretStore` → Doppler                            |
| Bootstrap token | `infra/k8s/doppler-token.secret.yaml`   | gitignored; devenv `files` from `DOPPLER_TOKEN`           |
| App secrets     | `infra/charts/refurbished-marketplace/` | `external-secrets.tpl` + minimal `externalSecrets` values |

Tilt applies `infra/k8s/*` via `k8s_yaml`. Application ExternalSecrets ship with the marketplace Helm release so secret names stay aligned with `services.*.db.secretName` and `services.*.auth`.

**Alternatives considered:** flat YAML under `infra/eso/` for all resources — rejected in favour of reusing service definitions already in chart values.

### Doppler service token bootstrap

| Layer          | Mechanism                                                                           |
| -------------- | ----------------------------------------------------------------------------------- |
| Project/config | `DOPPLER_PROJECT` and `DOPPLER_CONFIG` in `devenv.nix` (no `.doppler.yaml`)         |
| Machine auth   | Dev service token in gitignored `.env` as `DOPPLER_TOKEN=dp.st.dev…`                |
| devenv         | `dotenv.enable` + `files."infra/k8s/doppler-token.secret.yaml".yaml` when token set |
| Tilt           | `k8s_yaml` applies generated manifest alongside `cluster-secret-store.yaml`         |

ESO expects secret key **`dopplerToken`**. Token Secret lives in the **`operators`** namespace (same as the ESO operator Helm release).

```yaml
# ClusterSecretStore (excerpt)
spec:
  provider:
    doppler:
      auth:
        secretRef:
          dopplerToken:
            name: doppler-token
            key: dopplerToken
            namespace: operators
```

Optional devenv `scripts.tilt` wrapper: fail fast if `DOPPLER_TOKEN` unset before exec'ing `tilt`.

### ESO operator install

Upstream Helm chart in the `operators` namespace (alongside CNPG and Strimzi):

```bash
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace operators --create-namespace
```

Pin chart version in Tiltfile.

### ExternalSecret generation from chart values

**DB credentials** — for each enabled `services.<slug>` with `db`:

- ExternalSecret name = `db.secretName` (e.g. `users-app`)
- Doppler keys = `{SECRET_NAME}_USERNAME` / `_PASSWORD` where `SECRET_NAME` is `secretName` uppercased with `-` → `_` (e.g. `USERS_APP_USERNAME`)
- Target template: `kubernetes.io/basic-auth` (shared default in template)

**Auth secrets** — deduped from `services.<slug>.auth`:

- ExternalSecret name = `auth.secretName` (e.g. `users-auth`)
- Doppler remote key = `auth.secretKey` (e.g. `JWT_SECRET`)

| K8s Secret     | Source in values                           | Keys                   |
| -------------- | ------------------------------------------ | ---------------------- |
| `users-app`    | `services.users.db`                        | `username`, `password` |
| `products-app` | `services.products.db`                     | `username`, `password` |
| `orders-app`   | `services.orders.db`                       | `username`, `password` |
| `payment-app`  | `services.payment.db`                      | `username`, `password` |
| `users-auth`   | `services.web.auth`, `services.users.auth` | `JWT_SECRET`           |

All application secrets in the `ecommerce` namespace.

### Tilt dependency order

```
eso-operator-install (helm, operators)
  → k8s_yaml(infra/k8s/doppler-token + cluster-secret-store)
  → cnpg-operator-install
  → refurbished-marketplace helm (ExternalSecrets + apps + DBs)
  → kafka chart (Debezium needs orders-app / payment-app secrets)
```

Deploy marketplace chart before Kafka so connector secrets exist.

### Local and remote use same layout

Remote cluster: admin bootstraps prod service token once (not in Git); same `infra/k8s/` cluster manifests and marketplace chart values. Swap provider by editing `ClusterSecretStore` and, if needed, Doppler key mappings in chart values — not service deployment templates.

## Risks / Trade-offs

- **[New contributor friction]** → Document Doppler + `.env` in `docs/development/secrets.md`
- **[ESO sync race on cold start]** → Deploy marketplace chart before Kafka/Debezium; ESO retries refresh
- **[Token in nix store]** → devenv `files` evaluates at shell build; acceptable for local dev service tokens
- **[Doppler outage]** → ESO uses last synced Secret; local dev needs network on first sync

## Migration Plan

1. Create Doppler project/configs; seed `dev` secrets matching former `secrets.yaml`
2. Add devenv Doppler + dotenv + `files` token manifest; document `.env`
3. Add ESO operator + `infra/k8s/` + marketplace `external-secrets.tpl`
4. Verify `tilt up` without `secrets.yaml`
5. Delete `infra/k8s/secrets.yaml`
6. Update development docs and #10 / #4 notes

Rollback: restore `secrets.yaml` + Tilt line temporarily.

## Open Questions

- Exact Doppler project name (implemented as `refurbished-marketplace`)
- Optional `db.dopplerPrefix` override if a secret name does not match the key convention (not needed for current services)
