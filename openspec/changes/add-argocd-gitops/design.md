## Context

Local development uses Tilt with chart default `values.yaml` and locally built images (`refurbished-marketplace/web`). GHCR publishes images on merge to `main` with `:main` and `:<sha>` tags. ESO + Doppler sync secrets; cluster bootstrap (ArgoCD install, Doppler token, ClusterSecretStore) is handled outside this repo.

Two remote environments, one Kubernetes cluster each:

- **staging** — Proxmox
- **production** — AWS

Each cluster has ArgoCD pre-installed (Terraform). This change adds only the Git manifests Argo syncs ongoing.

## Goals / Non-Goals

**Goals:**

- App-of-apps per environment: root Application → operators, marketplace, kafka
- Env value overlays (`staging` / `production`) without a `dev` overlay
- Helm charts resolve images as `{{ registry }}/{{ name }}:{{ tag }}` with `global.imageRegistry` + `global.imageTag`
- Staging tracks `:main`; production pins a single commit SHA for all services
- Release workflow builds all twelve images on every main push so every `:sha` tag exists for production promotion
- Payment gateway simulator lives in marketplace chart for staging and production

**Non-goals:**

- ArgoCD installation, cluster registration, secrets bootstrap
- ApplicationSet / multi-cluster management plane
- Strict sync-wave ordering beyond loose operators → marketplace → kafka
- Real hosted payment gateway integration

## Decisions

### App-of-apps layout (not ApplicationSet)

```
infra/argocd/
├── staging/
│   ├── root.yaml                 # Application → staging/apps/
│   └── apps/
│       ├── operators-eso.yaml
│       ├── operators-cnpg.yaml
│       ├── operators-strimzi.yaml
│       ├── refurbished-marketplace.yaml
│       └── kafka.yaml
├── production/
│   └── (same shape)
└── values/
    ├── staging/
    │   ├── refurbished-marketplace.yaml
    │   └── kafka.yaml
    └── production/
        ├── refurbished-marketplace.yaml
        └── kafka.yaml
```

Terraform creates one root Application per cluster pointing at `staging/` or `production/`. Plain Applications — no ApplicationSet — because only two clusters and each has its own Argo instance.

**Alternatives considered:** ApplicationSet with cluster generator — rejected as overkill for two independent Argo installs.

### Sync order (loose)

Child Applications use `sync-wave` annotations:

| Wave | Application                    |
| ---- | ------------------------------ |
| 0    | operators (ESO, CNPG, Strimzi) |
| 1    | refurbished-marketplace        |
| 2    | kafka                          |

No secret-first sub-ordering; Argo reconcile + ESO retries are acceptable (same philosophy as relaxed Tilt ordering).

### Image reference model (Option B)

Chart templates prefix service image names:

```yaml
# values.yaml (dev / Tilt — registry empty or omitted → local names unchanged)
global:
  imageRegistry: ""          # empty = use image as-is for Tilt
  imageTag: ""

# staging overlay
global:
  imageRegistry: ghcr.io/<owner>/refurbished-marketplace
  imageTag: main

# production overlay
global:
  imageRegistry: ghcr.io/<owner>/refurbished-marketplace
  imageTag: abc123def   # commit SHA, updated on promote
```

Template helper builds `imageRef(name, tag)` — if `imageRegistry` is empty, return `name` only (Tilt). Optional `services.<slug>.imageTag` override reserved but not required when all images share a release SHA.

Apply same pattern to `kafka` chart for `connect.image` and `connect-debezium`.

### Release workflow: build all images on main

Remove `dorny/paths-filter` gating from `release-images.yml`. On push to `main` (with existing path trigger) or `workflow_dispatch`, matrix always builds all twelve images. Each gets `:main` and `:${{ github.sha }}`.

**Rationale:** Production deploy sets one `global.imageTag` — every service must have that SHA tag. Path-filtered releases left gaps (e.g. `cart:abc123` missing after web-only commit).

**Trade-off:** Longer CI on every main merge; acceptable for small fleet.

### Tag promotion strategy

| Env        | `global.imageTag` | Update mechanism                                   |
| ---------- | ----------------- | -------------------------------------------------- |
| staging    | `main`            | automatic on sync after CI push                    |
| production | commit SHA        | manual Git edit to values overlay (GitOps promote) |

### Payment gateway simulator in marketplace chart

Move `infra/k8s/payment-gateway-simulator.yaml` into chart as `services.payment-gateway-simulator` (or dedicated template block). Default values enable for all envs initially.

Env overrides:

- `HOSTED_PAYMENT_BASE_URL` → in-cluster URL (`http://payment-gateway-simulator:8097`) for staging/production
- Tilt default keeps `http://localhost:8097` in base `values.yaml`

### Operators as Helm Applications

Each operator is an Argo `Application` with upstream Helm chart source (pinned versions matching Tiltfile). Namespace: `operators`. Same chart versions in staging and production unless env values differ later.

### Dev unchanged

Tilt continues using `infra/charts/*/values.yaml` directly. No `infra/argocd/values/dev/`.

## Risks / Trade-offs

- **[Full image matrix CI time]** → Accept for consistent SHA tags; matrix is only twelve images
- **[Private GHCR pull]** → Document `imagePullSecrets` if needed; out of scope to automate
- **[Production SHA bump is manual]** → Intentional GitOps gate; document in deploy guide
- **[Simulator in prod]** → Temporary until real gateway; disable via values when ready
- **[Bootstrap out of repo]** → First sync may fail until external bootstrap completes; document prerequisites

## Migration Plan

1. Add chart `global.imageRegistry` / `imageTag` helpers; verify Tilt still works with empty registry
2. Move simulator into chart; update Tiltfile
3. Simplify `release-images.yml` to full matrix
4. Add `infra/argocd/` manifests and value overlays
5. Document staging vs production setup and production SHA promotion
6. External: point Terraform root Application at new paths; verify staging sync

Rollback: revert production `imageTag` in Git or Argo history; Tilt/dev unaffected.

## Open Questions

- Exact GHCR repository path string in value files (derive from `github.repository` at implement time)
- Proxmox vs AWS storage class names in env overlays (fill during implement when clusters exist)
- Whether production should disable simulator before real gateway (user decision later)
