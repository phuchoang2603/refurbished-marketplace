# Secrets (Doppler + ESO)

Application secrets are **not** committed to Git. [External Secrets Operator](https://external-secrets.io/) (ESO) syncs them from Doppler into Kubernetes. Helm charts reference Kubernetes Secret names only (`users-app`, `users-auth`, …).

## Doppler project

1. Create a Doppler project named `refurbished-marketplace`.
2. Create two configs:
   - `dev` — local Tilt
   - `prd` — staging and production

`devenv.nix` sets `DOPPLER_PROJECT` and `DOPPLER_CONFIG=dev` for the Doppler CLI.

## Application secrets

Add the keys below in the Doppler UI for each config (`dev` or `prd`).

Open the project → select the config → **Secrets** → add each key and value there. Use the UI password generator for `*_PASSWORD` and `JWT_SECRET` values.

| Doppler key             | K8s Secret     | K8s key      |
| ----------------------- | -------------- | ------------ |
| `USERS_APP_USERNAME`    | `users-app`    | `username`   |
| `USERS_APP_PASSWORD`    | `users-app`    | `password`   |
| `PRODUCTS_APP_USERNAME` | `products-app` | `username`   |
| `PRODUCTS_APP_PASSWORD` | `products-app` | `password`   |
| `ORDERS_APP_USERNAME`   | `orders-app`   | `username`   |
| `ORDERS_APP_PASSWORD`   | `orders-app`   | `password`   |
| `PAYMENT_APP_USERNAME`  | `payment-app`  | `username`   |
| `PAYMENT_APP_PASSWORD`  | `payment-app`  | `password`   |
| `JWT_SECRET`            | `users-auth`   | `JWT_SECRET` |

Guidelines:

- **dev:** simple values are fine for local development.
- **prd:** use unique, strong values. Do not reuse `dev` secrets.
- Usernames should match chart `db.owner` values: `users_app`, `products_app`, `orders_app`, `payment_app`.
- `JWT_SECRET` is shared by `web` and `users` through `users-auth`.

DB keys are derived from `db.secretName` (for example `users-app` → `USERS_APP_USERNAME` / `USERS_APP_PASSWORD`). Auth keys use the `auth.secretKey` name directly (`JWT_SECRET`).

## Bootstrap service tokens

ESO reads a Doppler service token from Kubernetes Secret `operators/doppler-token`. Create one token per config in the Doppler UI:

1. Open the target config (`dev` or `prd`).
2. Go to **Access** → **Service Tokens**.
3. Create a **read-only** token scoped to that config.
4. Copy the token (`dp.st…`) — it is shown only once.

Reference: [Doppler service tokens](https://docs.doppler.com/docs/service-tokens)

Store each token in a gitignored manifest:

| File                            | Config | Applied by             |
| ------------------------------- | ------ | ---------------------- |
| `doppler-token.dev.secret.yaml` | `dev`  | Tilt                   |
| `doppler-token.prd.secret.yaml` | `prd`  | manual `kubectl apply` |

Examples are committed as `infra/k8s/doppler-token.dev.secret.yaml.example` and `infra/k8s/doppler-token.prd.secret.yaml.example`.

### Local (`dev`)

```bash
cp infra/k8s/doppler-token.dev.secret.yaml.example infra/k8s/doppler-token.dev.secret.yaml
```

Paste the `dev` service token, then run `tilt up`.

### Staging / production (`prd`)

```bash
cp infra/k8s/doppler-token.prd.secret.yaml.example infra/k8s/doppler-token.prd.secret.yaml
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
```

Seed `prd` application secrets in Doppler before applying the bootstrap token and syncing Argo CD.

## Verify sync

Local:

```bash
kubectl get externalsecrets,secrets -n ecommerce
```

Remote:

```bash
kubectl get clustersecretstore doppler
kubectl get externalsecrets,secrets -n ecommerce
```

## Swapping the secrets provider

To move off Doppler, edit `infra/charts/operators/external-secrets/values.yaml` and adjust marketplace chart `externalSecrets` values if remote key names change. Service deployment templates do not need changes.

## Related issues

- [#10 — ESO + Doppler](https://github.com/phuchoang2603/refurbished-marketplace/issues/10)
- [#4 — ArgoCD GitOps](https://github.com/phuchoang2603/refurbished-marketplace/issues/4)
