# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime (Colima + Docker / k3s recommended)
- [Doppler](https://www.doppler.com/) account for cluster secrets — see [secrets.md](secrets.md)
- A Cloudflare Zero Trust tunnel for local `.dev` hostnames (same pattern as staging)

### Colima

Give the VM enough memory for ambient Istio + Kafka + CNPG (about **10 GiB**). Keep Traefik disabled:

```yaml
cpu: 4
memory: 10
kubernetes:
  enabled: true
  k3sArgs:
    - --disable=traefik
```

## Development shell

```bash
devenv shell
```

The shell provides Go, protobuf tooling, Kubernetes tooling (`kubectl`, `helm`), Doppler, OpenSpec, and devenv scripts for local Argo. On enter, devenv tasks regenerate proto/sqlc/templ/tailwind when those inputs change. A gitignored `.env` file is loaded automatically.

## Local Argo CD + Cloudflare Tunnel

Local development mirrors staging GitOps: Argo CD syncs [`infra/argocd/local/`](../../infra/argocd/local/). Chart `values.yaml` enables ambient mesh and Istio ingress for:

| Hostname                 | Backend                     |
| ------------------------ | --------------------------- |
| `shop.dev.phuchoang.sbs` | `web`                       |
| `pay.dev.phuchoang.sbs`  | `payment-gateway-simulator` |

Browser traffic: Cloudflare Tunnel → `ecommerce-ingress-istio.ecommerce.svc:80` (no per-service port-forwards).

1. Secrets: copy `infra/k8s/doppler-token.dev.secret.yaml.example` → `doppler-token.dev.secret.yaml` and paste the Doppler `dev` token.
2. In Doppler `dev`, set `CLOUDFLARE_TUNNEL_TOKEN` for a dedicated local tunnel.
3. In Cloudflare Zero Trust → that tunnel → Public Hostnames:
   - `shop.dev.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
   - `pay.dev.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
4. Push this branch (Argo reads GitHub). Bootstrap pins Applications to the **current git branch** (override with `ARGO_REVISION`).
5. Bootstrap and build:

```bash
bootstrap-local-argocd
build-images
# or one service: build-images web
```

6. Open https://shop.dev.phuchoang.sbs

Web assets: `templ-gen` / `tailwind-gen` (also run automatically via devenv tasks `web:templ` / `web:tailwind` when those files change on shell enter). After UI changes, rebuild the `web` image: `build-images web`.

Smoke-check:

```bash
kubectl get applications -n argo-cd
kubectl get gateway,httproute -n ecommerce
kubectl get pods -n cloudflare-tunnel
kubectl get pods -n ecommerce
```

## Integration testing

Integration tests rely on Testcontainers for Kafka, PostgreSQL, and Redis/Valkey. Prefer verifying full-service flows against the local Argo stack; run targeted Go tests when they add meaningful coverage.
