## Why

The marketplace still treats the former Proxmox production cluster as a development server. The platform now provides independent Proxmox dev and AWS prod clusters, but the marketplace has no dev root or automated deployment and its production overlay retains excess replicas.

## What Changes

- Deploy trusted repository PRs to dev after matching immutable images are published; deploy main to prod.
- Add independent environment roots and credentials while sharing marketplace charts and platform operators.
- Keep single-instance databases and minimal service replicas, using each cluster's default storage.
- Keep dev checkout unavailable until a separate payment gateway rollout.
- Validate charts and live dev convergence before considering this rollout complete.

Non-goals: deploying the payment gateway, moving marketplace resources into platform data/ML namespaces, database-engine changes, or applying AWS infrastructure locally.

## Capabilities

### New Capabilities

- `environment-delivery`: Trusted PR delivery to dev and main delivery to prod with matching Git/image revisions and isolated bootstrap credentials.

### Modified Capabilities

- `argocd-gitops`: Independent environment roots and immutable coordinated image overrides.
- `ghcr-release`: Publish deployment changes and retain images still needed by dev.
- `github-actions-ci`: Deploy the published revision instead of committing overlay pins.
- `external-secrets`: Independent dev/prod config selection and local bootstrap.

## Impact

Marketplace Argo roots, image release/deployment workflows, environment overlays, GitHub Environment policies, Doppler dev credentials, and deployment documentation. Shared infrastructure remains owned by talos-proxmox. Existing production checkout integration remains configured; its gateway readiness is a prerequisite for production checkout acceptance.
