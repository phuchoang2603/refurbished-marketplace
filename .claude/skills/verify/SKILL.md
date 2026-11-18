---
name: verify
description: Prove a change to refurbished-marketplace works, from local lint through CI, published images, Argo CD and the public hostname.
---

1. Lint locally: `devenv tasks run repo:lint` (ast-grep, golangci-lint with depguard, helm lint, conftest policies in `policy/`). Never run tests locally; CI runs them.
2. Push, then read CI: `gh run list --commit "$(git rev-parse HEAD)"`, `gh run watch <id> --exit-status`, `gh run view <id> --log-failed`. The `test` matrix runs `scripts/check.sh <service>` per affected service.
3. Images: the Release Images run publishes `ghcr.io/phuchoang2603/refurbished-marketplace/<service>:main-<sha>` on main or `:pr-<number>-<merge-sha>` on a PR. Confirm with `gh run view <id>` that the `images` jobs passed and the deploy ref (`deploy/prod` or `deploy/dev`) moved.
4. Argo CD: use the `talos-k8s` skill for kubeconfig, then `kubectl get applications -n argo-cd` and check `<env>-refurbished-marketplace`, `<env>-kafka` and `<env>-meilisearch` are `Synced` and `Healthy` at the new revision.
5. Public check: `curl -fsSI https://shop.phuchoang.sbs/` (`ingress.webHostname` in `infra/charts/refurbished-marketplace/values.yaml` and `values-prod.yaml`) and expect a 2xx or 3xx from `web`.
