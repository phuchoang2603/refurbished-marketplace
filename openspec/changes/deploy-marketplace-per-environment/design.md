## Context

See proposal.md. Production runs on independent GCP infrastructure. Marketplace charts consume ambient mesh, platform CRDs and default storage.

## Goals / Non-Goals

Provide a single release pipeline with trusted PR dev delivery and main prod delivery. Keep application ownership in ecommerce and platform ownership in talos-proxmox. Payment gateway rollout and payment acceptance are excluded.

## Decisions

- Reuse the release image matrix and add chart validation and GitOps deployment refs. Every infra/chart/root change builds a complete matching image set to avoid deploying stale tags.
- Bootstrap each Argo root once from a local kubeconfig. CI publishes a deployment commit to deploy/dev or deploy/prod only after image publication; its app-of-apps values pin children to the immutable built source revision and image tag. The root watches its environment deployment ref and retains environment overlays. CI uses optimistic force-with-lease and skips moved source refs, without writing main or accessing Kubernetes.
- GitHub Environments restrict PR dev publication and main prod publication. CI only needs GITHUB_TOKEN for registry and Git writes. Bootstrap a separate read-only marketplace token into each cluster locally; application ESO credentials and kubeconfigs never enter CI.
- Keep one database instance, broker and service replica; use environment default StorageClasses. Production databases select On-Demand nodes.
- Keep dev payment integration disabled. Production retains its existing gateway settings; checkout readiness depends on that separate deployment.
- Retain PR images on close because the dev cluster can still reference them until another deployment. Revisit garbage collection independently.

## Risks / Trade-offs

- Platform readiness or missing operator APIs can delay convergence; child retries absorb ordering and CI checks root submission separately from acceptance.
- PR workflow credentials are available only to same-repository PRs targeting main; fork PRs build without publishing or deploying.
- Failed migrations require schema compatibility review before reverting images; do not claim automatic data rollback.

## Migration Plan

Bootstrap independent dev credentials, validate charts and workflows, open a PR, apply the dev AppProject/root once using the local kubeconfig, and verify live dev state and public storefront after image publication. Main deployment occurs only after user merge. Do not apply cloud infrastructure locally.
