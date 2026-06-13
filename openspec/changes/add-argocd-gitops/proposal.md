## Why

Issue #4’s remaining work is GitOps delivery: non-local clusters (staging on Proxmox, production on AWS) must deploy from Git without Tilt. GHCR images and ESO secrets are in place; the repo still lacks ArgoCD Application manifests, env-specific Helm values, and chart support for registry-backed images. A single `global.imageTag` per environment requires every service image to share the same commit tag on each release.

## What Changes

- Add `infra/argocd/` with app-of-apps layout: per-env root Application (`staging`, `production`) and child Applications for operators (ESO, CNPG, Strimzi), `refurbished-marketplace`, and `kafka`
- Add env value overlays under `infra/argocd/values/staging/` and `infra/argocd/values/production/` (no `dev` — Tilt keeps chart default `values.yaml`)
- Extend Helm charts with `global.imageRegistry` and `global.imageTag` (and optional per-service `imageTag` override) so remote clusters pull from GHCR
- Move `payment-gateway-simulator` into the marketplace chart; enable in staging and production via values
- **BREAKING (workflow):** Change `release-images.yml` to build and push **all** twelve images on every qualifying push to `main` (remove per-image path-filter fan-out) so production can pin one SHA across the fleet
- Document GitOps layout and env promotion (`main` tag for staging, commit SHA for production) in development docs

### Non-goals

- Installing ArgoCD, cluster bootstrap, Doppler token, or `ClusterSecretStore` via this change (Terraform / manual bootstrap stays out of repo scope)
- Observability stack Applications (#1–#3)
- `imagePullSecrets` automation (may document manual step if GHCR packages are private)
- ApplicationSet generators (plain app-of-apps per env)

## Capabilities

### New Capabilities

- `argocd-gitops`: ArgoCD app-of-apps manifests, env-specific Helm values, chart image registry wiring, payment-gateway-simulator in marketplace chart

### Modified Capabilities

- `ghcr-release`: Release workflow builds all images on each main push (drop path-filter selective rebuild requirement)

## Impact

- **Add:** `infra/argocd/`, env value files, deploy documentation
- **Update:** `infra/charts/refurbished-marketplace/` (global image refs, simulator), `infra/charts/kafka/` (connect/kafka image refs), `.github/workflows/release-images.yml`, `Tiltfile` (simulator path), `docs/development/`
- **Remove:** standalone `infra/k8s/payment-gateway-simulator.yaml` (absorbed into chart)
- **Issue #4:** satisfies ArgoCD Applications + env values + non-Tilt deploy path (bootstrap still external)
