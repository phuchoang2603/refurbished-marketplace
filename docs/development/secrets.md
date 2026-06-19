# Secrets (Doppler + ESO)

Application secrets are **not** committed to Git. Tilt syncs them from Doppler via [External Secrets Operator](https://external-secrets.io/) (ESO). Helm charts continue to reference the same Kubernetes Secret names (`users-app`, `users-auth`, …).

## One-time setup

```bash
devenv shell
doppler login                    # once, to manage secrets via CLI
# Create a read-only service token for the dev config in the Doppler dashboard
echo 'DOPPLER_TOKEN=dp.st.dev.xxxx' >> .env
devenv shell                     # re-enter so devenv links infra/k8s/doppler-token.secret.yaml
```

Project and config (`refurbished-marketplace` / `dev`) are set in `devenv.nix` as `DOPPLER_PROJECT` and `DOPPLER_CONFIG` — no `.doppler.yaml` needed.

devenv generates `infra/k8s/doppler-token.secret.yaml` from `DOPPLER_TOKEN` when you enter the shell (symlinked, gitignored). Tilt applies it with `infra/k8s/cluster-secret-store.yaml`. Application `ExternalSecret` resources are defined in `infra/charts/refurbished-marketplace/values.yaml` and rendered by the marketplace Helm chart.

## Doppler project setup

1. Create a Doppler project named `refurbished-marketplace` (or change `DOPPLER_PROJECT` in `devenv.nix`).
2. Create configs: `dev` (local Tilt), `prd` (production), and others as needed.
3. `devenv.nix` sets `DOPPLER_PROJECT` and `DOPPLER_CONFIG=dev` for the Doppler CLI.

To seed secrets via CLI (after `doppler login`):

```bash
devenv shell
doppler secrets set USERS_APP_USERNAME=users_app
# … see key table below
```

## Service token for ESO

Create a **read-only service token** scoped to the target config ([Doppler service tokens](https://docs.doppler.com/docs/service-tokens)):

- Local Tilt: token for the `dev` config
- Remote cluster: separate token for `prd` (bootstrap once on the cluster; not in Git)

Add the local dev token to gitignored `.env`:

```bash
DOPPLER_TOKEN=dp.st.dev.xxxx
```

Re-enter `devenv shell` after changing `.env`. The ESO [Doppler provider](https://external-secrets.io/latest/provider/doppler/) reads Kubernetes Secret `operators/doppler-token` key `dopplerToken`.

## Seed `dev` config keys

Add these keys to the Doppler **`dev`** config. Values match the former `infra/k8s/secrets.yaml` dev credentials.

DB credentials use Doppler keys derived from each service's `db.secretName` (for example `users-app` → `USERS_APP_USERNAME` / `USERS_APP_PASSWORD`). Auth secrets use the `auth.secretKey` as the Doppler key (for example `JWT_SECRET` on `users-auth`).

| Doppler key             | Example value               | K8s Secret     | K8s key      |
| ----------------------- | --------------------------- | -------------- | ------------ |
| `USERS_APP_USERNAME`    | `users_app`                 | `users-app`    | `username`   |
| `USERS_APP_PASSWORD`    | `users_app_dev_password`    | `users-app`    | `password`   |
| `PRODUCTS_APP_USERNAME` | `products_app`              | `products-app` | `username`   |
| `PRODUCTS_APP_PASSWORD` | `products_app_dev_password` | `products-app` | `password`   |
| `ORDERS_APP_USERNAME`   | `orders_app`                | `orders-app`   | `username`   |
| `ORDERS_APP_PASSWORD`   | `orders_app_dev_password`   | `orders-app`   | `password`   |
| `PAYMENT_APP_USERNAME`  | `payment_app`               | `payment-app`  | `username`   |
| `PAYMENT_APP_PASSWORD`  | `payment_app_dev_password`  | `payment-app`  | `password`   |
| `JWT_SECRET`            | `dev-jwt-secret`            | `users-auth`   | `JWT_SECRET` |

After `tilt up`, verify ExternalSecrets synced from the marketplace chart:

```bash
kubectl get externalsecrets,secrets -n ecommerce
kubectl describe externalsecret -n ecommerce users-app
```

## Swapping the secrets provider

Helm and application code reference Kubernetes Secret names only. To move off Doppler, edit `infra/k8s/cluster-secret-store.yaml` to use another [ESO provider](https://external-secrets.io/latest/provider/overview/) and adjust `externalSecrets` / service `db` / `auth` settings in `infra/charts/refurbished-marketplace/values.yaml` if remote key names change. Service deployment templates do not need changes.

Remote clusters use the same cluster-level manifests under `infra/k8s/` and the marketplace chart values; bootstrap a config-scoped service token for the target environment separately. On staging, the ESO operator itself is installed by Terraform (alongside Argo CD), not by an Argo CD Application — see [gitops.md](../deployment/gitops.md).

## Related issues

- [#10 — ESO + Doppler](https://github.com/phuchoang2603/refurbished-marketplace/issues/10)
- [#4 — ArgoCD GitOps](https://github.com/phuchoang2603/refurbished-marketplace/issues/4)
