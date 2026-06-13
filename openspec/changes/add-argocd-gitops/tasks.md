## 1. Helm chart image registry

- [x] 1.1 Add `global.imageRegistry` / `global.imageTag` helpers to marketplace and kafka chart templates
- [x] 1.2 Verify `helm template` with empty registry still matches Tilt local image names
- [x] 1.3 Add optional per-service `imageTag` override in template helper (fallback to global)

## 2. Payment gateway simulator

- [x] 2.1 Move simulator Deployment/Service into marketplace chart templates and values
- [x] 2.2 Remove `infra/k8s/payment-gateway-simulator.yaml` and update Tiltfile
- [x] 2.3 Set staging/production overlay URLs for `HOSTED_PAYMENT_BASE_URL` to in-cluster simulator

## 3. Release workflow

- [x] 3.1 Remove path-filter job and conditional matrix skips from `release-images.yml`
- [x] 3.2 Build and push all twelve images on every triggered workflow run
- [x] 3.3 Confirm each image receives `:main` and `:${{ github.sha }}` tags

## 4. ArgoCD manifests

- [x] 4.1 Create `infra/argocd/staging/` root Application and `apps/` child Applications (operators, marketplace, kafka)
- [x] 4.2 Create `infra/argocd/production/` with the same application shape
- [x] 4.3 Add sync-wave annotations: operators (0), marketplace (1), kafka (2)
- [x] 4.4 Pin operator Helm chart versions to match Tiltfile

## 5. Environment values

- [x] 5.1 Add `infra/argocd/values/staging/refurbished-marketplace.yaml` (`imageTag: main`, GHCR registry, simulator URL)
- [x] 5.2 Add `infra/argocd/values/staging/kafka.yaml` (registry + tag, Proxmox-oriented defaults as needed)
- [x] 5.3 Add `infra/argocd/values/production/` overlays with SHA placeholder and AWS-oriented defaults as needed
- [x] 5.4 Wire Argo Applications to chart paths and env value files

## 6. Documentation and tracking

- [x] 6.1 Add `docs/development/deploy-gitops.md` (app-of-apps layout, bootstrap out-of-scope, SHA promotion)
- [x] 6.2 Link from development README / CONTRIBUTING
- [x] 6.3 Update GitHub #4 checklist for ArgoCD Applications and env values

## 7. Verification

- [x] 7.1 `helm template` staging and production value combinations for marketplace and kafka
- [x] 7.2 `tilt up` still works with default values after simulator move
- [x] 7.3 Dry-run or staging cluster sync (manual, when cluster available)
