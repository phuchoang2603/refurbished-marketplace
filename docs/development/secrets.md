# Secrets (Doppler + ESO)

Application secrets are **not** committed to Git. Tilt syncs them from Doppler via [External Secrets Operator](https://external-secrets.io/) (ESO). Helm charts continue to reference the same Kubernetes Secret names (`users-app`, `users-auth`, …).

## One-time setup

```bash
devenv shell
doppler login                    # once, to manage secrets via CLI
# Create a read-only service token for the dev config in the Doppler dashboard
echo 'DOPPLER_TOKEN=dp.st.dev.xxxx' >> .env
devenv shell                     # re-enter so devenv links infra/eso/doppler-token.secret.yaml
```

Project and config (`refurbished-marketplace` / `dev`) are set in `devenv.nix` as `DOPPLER_PROJECT` and `DOPPLER_CONFIG` — no `.doppler.yaml` needed.

devenv generates `infra/eso/doppler-token.secret.yaml` from `DOPPLER_TOKEN` when you enter the shell (symlinked, gitignored). Tilt applies it with the other manifests in `infra/eso/` and waits for synced secrets before database and app charts deploy.

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

Re-enter `devenv shell` after changing `.env`. The ESO [Doppler provider](https://external-secrets.io/latest/provider/doppler/) reads Kubernetes Secret `external-secrets/doppler-token` key `dopplerToken`.

## Seed `dev` config keys

Add these keys to the Doppler **`dev`** config. Values match the former `infra/k8s/secrets.yaml` dev credentials:

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

ESO manifests live in `infra/eso/`. After `tilt up`, verify:

```bash
kubectl get externalsecrets,secrets -n ecommerce
kubectl describe externalsecret -n ecommerce users-app
```

## Swapping the secrets provider

Helm and application code reference Kubernetes Secret names only. To move off Doppler, edit `infra/eso/cluster-secret-store.yaml` to use another [ESO provider](https://external-secrets.io/latest/provider/overview/) and update `ExternalSecret` `remoteRef` mappings. No changes to `refurbished-marketplace` or `kafka` chart templates are required.

Remote clusters use the same `infra/eso/` manifests; bootstrap a config-scoped service token for the target environment separately.

## Related issues

- [#10 — ESO + Doppler](https://github.com/phuchoang2603/refurbished-marketplace/issues/10)
- [#4 — ArgoCD GitOps](https://github.com/phuchoang2603/refurbished-marketplace/issues/4)
