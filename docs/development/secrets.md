# Secrets (Doppler + ESO)

Application secret values are **not** committed to Git. The platform-owned External Secrets Operator syncs this repository's `ExternalSecret` resources through the marketplace-owned namespaced `SecretStore/doppler` in `ecommerce`. The bootstrap token Secret is also created in `ecommerce`.

Marketplace resources do not use the platform `ClusterSecretStore/doppler`. That store reads the `talos-proxmox` Doppler project and delivers platform credentials such as the Cloudflare operator's API token. Public hostnames need no marketplace secret: the operator creates their DNS records from the chart's `TunnelBinding`.

## Doppler project

1. Create a Doppler project named `refurbished-marketplace`.
2. Use Doppler config `dev` on talos-dev and `prd` on prod. The service token Secret on that cluster selects the config; Argo does not set it.

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

`CLOUDFLARE_TUNNEL_TOKEN` lives in the `talos-proxmox` Doppler project, not here. Its tunnel's Public Hostnames point at `http://cilium-gateway-ecommerce-ingress.ecommerce.svc.cluster.local:80`.

## Bootstrap service token

The marketplace `SecretStore/doppler` reads `ecommerce/doppler-token` key `dopplerToken`. Copy the tracked `.example` manifest for the target environment and replace only `REPLACE_ME` in the untracked copy.

Apply it with the environment kubeconfig from the `talos-proxmox` Doppler project (the same kubeconfig used for the Argo CD project and root). Create `ecommerce` first; the root's `CreateNamespace` would otherwise create it only after the token is needed.

```bash
export CLUSTER_ENV=dev  # use prod for production
umask 077
doppler secrets get KUBECONFIG --plain \
  --project talos-proxmox --config "$CLUSTER_ENV" \
  > "$HOME/.kube/talos-${CLUSTER_ENV}.yaml"
export KUBECONFIG="$HOME/.kube/talos-${CLUSTER_ENV}.yaml"

kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.dev.secret.yaml   # prod: doppler-token.prd.secret.yaml
```

Do not commit tokens. Do not set Doppler `dev`/`prd` in Helm or Argo values.

```bash
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets,secrets -n ecommerce
kubectl get secret mongodb-catalog-app -n ecommerce
kubectl get secret meilisearch-master-key -n ecommerce
```

The `secret-store` child Application (`infra/charts/secret-store`) is the only renderer of `SecretStore/doppler`; it syncs in wave 1, before MongoDB, Meilisearch, marketplace, and Kafka. If `<env>-secret-store` is Degraded or the store does not report `Ready=True`, check the `ecommerce/doppler-token` Secret and its Doppler config.
