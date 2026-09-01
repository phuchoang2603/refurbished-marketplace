# Secrets (Doppler + ESO)

Application secrets are **not** committed to Git. External Secrets Operator syncs them from Doppler into Kubernetes.

## Doppler project

1. Create a Doppler project named `refurbished-marketplace`.
2. Use `prd` (or another config) for the Talos cluster. `devenv.nix` may still default the CLI to `dev` for local experiments.

## Application secrets

| Doppler key               | K8s Secret                                         | K8s key      |
| ------------------------- | -------------------------------------------------- | ------------ |
| `USERS_APP_PASSWORD`      | `users-app`                                        | `password`   |
| `PRODUCTS_APP_PASSWORD`   | `products-app`                                     | `password`   |
| `ORDERS_APP_PASSWORD`     | `orders-app`                                       | `password`   |
| `PAYMENT_APP_PASSWORD`    | `payment-app`                                      | `password`   |
| `JWT_SECRET`              | `users-auth`                                       | `JWT_SECRET` |
| `CLOUDFLARE_TUNNEL_TOKEN` | `cloudflare-tunnel-token` (ns `cloudflare-tunnel`) | `token`      |

`CLOUDFLARE_TUNNEL_TOKEN` is the Zero Trust tunnel whose Public Hostnames point at `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

## Bootstrap service token

ESO reads `operators/doppler-token` key `dopplerToken`.

```bash
cp infra/k8s/doppler-token.prd.secret.yaml.example infra/k8s/doppler-token.prd.secret.yaml
# paste the read-only service token
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
```

Do not use Tilt to apply this Secret.

```bash
kubectl get clustersecretstore doppler
kubectl get externalsecrets,secrets -n ecommerce
```
