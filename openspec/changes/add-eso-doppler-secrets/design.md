## Context

Tilt applies `infra/k8s/secrets.yaml` with plaintext DB passwords and `JWT_SECRET`. Helm charts reference those Secret names via `secretKeyRef`, CNPG `passwordSecret`, and Strimzi `${secrets:namespace/name:key}` for Debezium. GHCR release workflow is done; ArgoCD (#4) needs secrets out of Git.

Doppler provides separate **configs** (`dev`, `prd`, …). ESO's [Doppler provider](https://external-secrets.io/latest/provider/doppler/) syncs into native Kubernetes Secrets. Service tokens ([Doppler docs](https://docs.doppler.com/docs/service-tokens)) are scoped to one config — ideal for ESO auth.

## Goals / Non-Goals

**Goals:**

- One secret path for local Tilt and future ArgoCD clusters (same `infra/eso/` manifests)
- Doppler `dev` config for local; `prd` (or staging) config for remote — different service tokens, same YAML shape
- Bootstrap token via devenv `.env` (`DOPPLER_TOKEN`) + Tilt automation — no token in Git
- Preserve existing K8s Secret names and keys expected by charts
- Provider-agnostic via ESO `ClusterSecretStore` (swap Doppler → Vault/AWS later)

**Non-goals:**

- Dedicated Helm chart for ExternalSecrets
- Extra devenv scripts beyond `doppler` package, `dotenv.enable`, optional `tilt` wrapper
- ArgoCD Application for ESO in this change (manual/upstream Helm first)
- Migrating secret values to Doppler in this PR's automation — document manual seeding from current dev values

## Decisions

### Flat YAML in `infra/eso/` (not a Helm chart)

Store `ClusterSecretStore` and one `ExternalSecret` per target Secret in `infra/eso/`. Tilt uses `k8s_yaml('./infra/eso/')`; ArgoCD later syncs the same directory. Avoids duplicating chart patterns from `refurbished-marketplace` / `kafka`.

**Alternatives considered:** `infra/charts/external-secrets/` — rejected as unnecessary for ~5 ExternalSecrets.

### Doppler service token bootstrap

| Layer        | Mechanism                                                                                                                    |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| Human setup  | `doppler login`, `doppler setup` → commit `.doppler.yaml` (project + default config)                                         |
| Machine auth | Dev service token in gitignored `.env` as `DOPPLER_TOKEN=dp.st.dev…`                                                         |
| devenv       | `dotenv.enable = true` loads `.env` on `devenv shell`                                                                        |
| Tilt         | `local_resource` runs `kubectl create secret … --from-literal=dopplerToken="$DOPPLER_TOKEN"` in `external-secrets` namespace |

ESO expects secret key **`dopplerToken`** (not `DOPPLER_TOKEN`). Service tokens are config-scoped — no `project`/`config` fields on `ClusterSecretStore` when using service token auth.

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
            namespace: external-secrets # required for ClusterSecretStore
```

Optional devenv `scripts.tilt` wrapper: fail fast if `DOPPLER_TOKEN` unset before exec'ing `tilt`.

### ESO operator install

Upstream Helm chart (same pattern as CNPG/Strimzi in Tiltfile):

```bash
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace
```

Pin chart version in Tiltfile when implementing.

### ExternalSecret → existing Secret contract

| K8s Secret     | Type       | Keys (Doppler / K8s)   |
| -------------- | ---------- | ---------------------- |
| `users-app`    | basic-auth | `username`, `password` |
| `products-app` | basic-auth | `username`, `password` |
| `orders-app`   | basic-auth | `username`, `password` |
| `payment-app`  | basic-auth | `username`, `password` |
| `users-auth`   | Opaque     | `JWT_SECRET`           |

All in `ecommerce` namespace. Seed Doppler `dev` config with values matching current `secrets.yaml` before removing that file.

### Tilt dependency order

```
eso-operator-install (helm)
  → doppler-token (local_resource, needs DOPPLER_TOKEN)
  → k8s_yaml(infra/eso/)
  → kafka + refurbished-marketplace charts (resource_deps on ESO secrets ready)
```

Use Tilt `resource_deps` so app charts apply after ExternalSecrets exist.

### Local and remote use same manifests

Remote cluster: admin bootstraps prod service token once (not in Git); same `infra/eso/` synced by ArgoCD later. Swap provider by editing `ClusterSecretStore` provider block only.

## Risks / Trade-offs

- **[New contributor friction]** → Document Doppler + `.env` setup in CONTRIBUTING; commit `.doppler.yaml`
- **[ESO sync race on cold start]** → Tilt resource_deps; ESO retries refresh
- **[Token in shell env]** → `.env` gitignored; never log token in Tilt output
- **[Doppler outage]** → ESO uses last synced Secret; local dev needs network on first sync

## Migration Plan

1. Create Doppler project/configs; seed `dev` secrets matching current YAML
2. Add devenv Doppler + dotenv; document service token in `.env`
3. Add ESO operator + `infra/eso/` + Tilt bootstrap
4. Verify `tilt up` without `secrets.yaml`
5. Delete `infra/k8s/secrets.yaml`
6. Update CONTRIBUTING and #10 / #4 notes

Rollback: restore `secrets.yaml` + Tilt line temporarily.

## Open Questions

- Exact Doppler project name (suggest matching repo: `refurbished-marketplace`)
- Whether to add optional `scripts.tilt` wrapper or rely on Tilt error when token missing
