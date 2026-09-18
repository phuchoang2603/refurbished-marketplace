# Secrets (Doppler + ESO)

Application secret values are **not** committed to Git. The platform-owned External Secrets Operator syncs this repository's `ExternalSecret` resources through the marketplace-owned namespaced `SecretStore/doppler` in `ecommerce`. The bootstrap token Secret is also created in `ecommerce`. The platform owns cloudflared and its directly provisioned tunnel token.

## Doppler project

1. Create a Doppler project named `refurbished-marketplace`.
2. Use Doppler config `dev` on talos-dev and `prd` on prod. The service token Secret on that cluster selects the config; Argo does not set it. Apply tokens with the **workload** kubeconfig, not the management cluster kubeconfig.

## Application secrets

| Doppler key              | K8s Secret                                | K8s key                            |
| ------------------------ | ----------------------------------------- | ---------------------------------- |
| `USERS_APP_PASSWORD`     | `users-app`                               | `password`                         |
| `PRODUCTS_APP_PASSWORD`  | `mongodb-catalog-app` (ns `ecommerce`)    | `password`, `username` (`catalog`) |
| `MEILI_MASTER_KEY`       | `meilisearch-master-key` (ns `ecommerce`) | `MEILI_MASTER_KEY`                 |
| `INVENTORY_APP_PASSWORD` | `inventory-app`                           | `password`                         |
| `ORDERS_APP_PASSWORD`    | `orders-app`                              | `password`                         |
| `PAYMENT_APP_PASSWORD`   | `payment-app`                             | `password`                         |
| `JWT_SECRET`             | `users-auth`                              | `JWT_SECRET`                       |

`mongodb-catalog-app` is mounted on products and read by the Kafka Connect products-outbox Mongo connector (Role in `ecommerce`). The SCRAM user is `catalog`, authenticated against database `catalog` (same DB as listings/outbox), with `readWrite` there. There is no products Postgres secret after catalog cutover.

`meilisearch-master-key` is mounted on Meilisearch and the search service. The Doppler value must be at least 16 bytes. Plaintext Meilisearch keys are not committed.

`CLOUDFLARE_TUNNEL_TOKEN` is the Zero Trust tunnel whose Public Hostnames point at `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

## Bootstrap service token

The marketplace `SecretStore/doppler` reads `ecommerce/doppler-token` key `dopplerToken`. Copy the tracked `.example` manifest for the target environment, replace only `REPLACE_ME` in the untracked copy, and apply it with that workload cluster's kubeconfig.

```bash
# talos-dev workloads
kubectl --kubeconfig="$HOME/.kube/talos-dev.yaml" apply -f infra/k8s/doppler-token.dev.secret.yaml
# prod workloads
kubectl --kubeconfig="$HOME/.kube/talos-prod.yaml" apply -f infra/k8s/doppler-token.prd.secret.yaml
```

Do not commit tokens. Do not set Doppler `dev`/`prd` in Helm or Argo values.

```bash
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets,secrets -n ecommerce
kubectl get secret mongodb-catalog-app -n ecommerce
kubectl get secret meilisearch-master-key -n ecommerce
```

Do not apply marketplace roots until `kubectl get secretstore doppler -n ecommerce` reports `Ready=True`.
