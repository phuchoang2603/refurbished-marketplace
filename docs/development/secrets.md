# Secrets (Doppler + ESO)

Application secret values are **not** committed to Git. The platform-owned External Secrets Operator syncs this repository's `ExternalSecret` resources through the marketplace-owned namespaced `SecretStore/doppler` in `ecommerce`. The bootstrap token Secret is also created in `ecommerce`.

Marketplace resources do not use the platform `ClusterSecretStore/doppler`. That store reads the `talos-proxmox` Doppler project and delivers platform credentials such as the Cloudflare operator's API token. Public hostnames need no marketplace secret: the operator creates their DNS records from the chart's `TunnelBinding`.

## Doppler project

1. Create a Doppler project named `refurbished-marketplace`.
2. Use Doppler config `prd` on the prod cluster, which serves as the development server. The service token Secret on that cluster selects the config; Argo does not set it.

## Application secrets

| Doppler key                      | K8s Secret                                | K8s key                          |
| -------------------------------- | ----------------------------------------- | -------------------------------- |
| `USERS_APP_PASSWORD`             | `users-app`                               | `password`                       |
| `USERS_RUNTIME_PASSWORD`         | `users-runtime`                           | `password`                       |
| `PRODUCTS_APP_PASSWORD`          | `products-app`                            | `password`                       |
| `PRODUCTS_RUNTIME_PASSWORD`      | `products-runtime`                        | `password`                       |
| `MEILI_MASTER_KEY`               | `meilisearch-master-key` (ns `ecommerce`) | `MEILI_MASTER_KEY`               |
| `INVENTORY_APP_PASSWORD`         | `inventory-app`                           | `password`                       |
| `INVENTORY_RUNTIME_PASSWORD`     | `inventory-runtime`                       | `password`                       |
| `ORDERS_APP_PASSWORD`            | `orders-app`                              | `password`                       |
| `ORDERS_RUNTIME_PASSWORD`        | `orders-runtime`                          | `password`                       |
| `CHECKOUT_APP_PASSWORD`          | `checkout-app`                            | `password`                       |
| `CHECKOUT_RUNTIME_PASSWORD`      | `checkout-runtime`                        | `password`                       |
| `PAYMENT_GATEWAY_API_KEY`        | `checkout-payment-gateway` (prod only)    | `PAYMENT_GATEWAY_API_KEY`        |
| `PAYMENT_GATEWAY_WEBHOOK_SECRET` | `checkout-payment-gateway` (prod only)    | `PAYMENT_GATEWAY_WEBHOOK_SECRET` |
| `JWT_SECRET`                     | `users-auth`                              | `JWT_SECRET`                     |

Only Checkout references `checkout-payment-gateway`. Chart defaults render neither that Secret reference nor its ExternalSecret. The gateway API key is sent as `Authorization: Bearer`; Edge derives the integration ID from its hash, so Checkout has no separate integration-ID environment variable. The webhook signing secret must match the gateway notifier's integration configuration.

Gateway prerequisite (managed in the gateway repository): `GATEWAY_INTEGRATIONS_JSON` supplies the same integration UUID to Edge and notifier, the SHA-256 API key hash, `synthetic: false`, and `webhook: {url, secret, events}`. Use `http://web.ecommerce.svc.cluster.local:8080/webhooks/payment-gateway` and subscribe to `attempt.declined`, `payment.succeeded`, and `payment.expired`. The raw API key and webhook secret are the two marketplace Doppler keys above. Provision these credentials before deploying Checkout; this repository does not create the gateway integration or its credentials.

Kafka Connect reads only catalog, inventory, orders and checkout credentials through its generated Roles.

`products-app` supplies the `products_app` database-owner credentials to the migration init container and PostgreSQL outbox connector. The existing `PRODUCTS_APP_PASSWORD` Doppler key is reused. The application uses restricted `products_runtime` credentials from `products-runtime`; provision `PRODUCTS_RUNTIME_PASSWORD` before rollout.

`meilisearch-master-key` is mounted on Meilisearch and the search service. The Doppler value must be at least 16 bytes. Plaintext Meilisearch keys are not committed.

`CLOUDFLARE_TUNNEL_TOKEN` lives in the `talos-proxmox` Doppler project, not here. Its tunnel's Public Hostnames point at `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`.

## Bootstrap service token

The marketplace `SecretStore/doppler` reads `ecommerce/doppler-token` key `dopplerToken`. Copy the tracked `.example` manifest for prod and replace only `REPLACE_ME` in the untracked copy.

Apply it with the environment kubeconfig from the `talos-proxmox` Doppler project (the same kubeconfig used for the Argo CD project and root). Create `ecommerce` first; the root's `CreateNamespace` would otherwise create it only after the token is needed.

```bash
umask 077
doppler secrets get KUBECONFIG --plain \
  --project talos-proxmox --config prod \
  > "$HOME/.kube/talos-prod.yaml"
export KUBECONFIG="$HOME/.kube/talos-prod.yaml"

kubectl create namespace ecommerce --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/doppler-token.prd.secret.yaml
```

Do not commit tokens. Do not set the Doppler config in Helm or Argo values.

```bash
kubectl get secretstore doppler -n ecommerce
kubectl get externalsecrets,secrets -n ecommerce
kubectl get secret products-app products-runtime -n ecommerce
kubectl get secret meilisearch-master-key -n ecommerce
```

The `secret-store` child Application (`infra/charts/secret-store`) is the only renderer of `SecretStore/doppler`; it syncs in wave 1, before Meilisearch, marketplace, and Kafka. If `prod-secret-store` is Degraded or the store does not report `Ready=True`, check the `ecommerce/doppler-token` Secret and its Doppler config.
