# Local setup

## Prerequisites

- [Nix](https://nixos.org/) with [devenv](https://devenv.sh/) for pinned tooling
- A local Kubernetes runtime (Colima + Docker / k3s recommended)
- [Doppler](https://www.doppler.com/) account for cluster secrets — see [secrets.md](secrets.md)
- A Cloudflare Zero Trust tunnel for local `.dev` hostnames (same pattern as staging)

### Colima

Chart defaults target **4 CPU / 8 GiB** Colima (tight Istio/Kafka/CNPG budgets; observability is apps-only: ecommerce/kafka metrics–logs–traces, no node-exporter/ksm/Alertmanager). Keep Traefik disabled:

```yaml
cpu: 4
memory: 8
kubernetes:
  enabled: true
  k3sArgs:
    - --disable=traefik
```

## Development shell

```bash
devenv shell
```

The shell provides Go, protobuf tooling, Kubernetes tooling (`kubectl`, `helm`), Doppler, OpenSpec, and Tilt. On enter, devenv tasks regenerate proto/sqlc when those inputs change. A gitignored `.env` file is loaded automatically.

## Hybrid Tilt + Argo CD

| Layer                                                               | Local owner                                 |
| ------------------------------------------------------------------- | ------------------------------------------- |
| Operators, Istio, Kafka, apps-only observability, Cloudflare Tunnel | Argo CD (`local-root` → `app-of-apps`)      |
| `refurbished-marketplace` chart (DBs, secrets, services, ingress)   | Tilt                                        |
| Image builds + `templ` / Tailwind watches                           | Tilt                                        |
| Browser                                                             | Cloudflare Tunnel → Istio Gateway           |
| Debug (optional)                                                    | Tilt port-forwards (`8080` web, gRPC, CNPG) |

Chart `values.yaml` enables ambient mesh and Istio ingress for:

| Hostname                 | Backend                     |
| ------------------------ | --------------------------- |
| `shop-dev.phuchoang.sbs` | `web`                       |
| `pay-dev.phuchoang.sbs`  | `payment-gateway-simulator` |

1. Secrets: copy `infra/k8s/doppler-token.dev.secret.yaml.example` → `doppler-token.dev.secret.yaml` and paste the Doppler `dev` token.
2. In Doppler `dev`, set `CLOUDFLARE_TUNNEL_TOKEN` for a dedicated local tunnel.
3. In Cloudflare Zero Trust → that tunnel → Public Hostnames:
   - `shop-dev.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
   - `pay-dev.phuchoang.sbs` → `http://ecommerce-ingress-istio.ecommerce.svc.cluster.local:80`
4. Push this branch (Argo reads GitHub). Tilt applies `local-root` with the **current git branch**; child Applications inherit that revision via `$ARGOCD_APP_SOURCE_TARGET_REVISION`.
5. Start Tilt:

```bash
tilt up
```

6. Open https://shop-dev.phuchoang.sbs (or use Tilt’s web port-forward on `localhost:8080` for debug).

`templ-watch` / `tailwind-watch` run under Tilt; rebuilds of the `web` image pick up those assets via `docker_build`.

Smoke-check:

```bash
kubectl get applications -n argo-cd
kubectl get gateway,httproute -n ecommerce
kubectl get pods -n cloudflare-tunnel
kubectl get pods -n ecommerce
kubectl get pods -n monitoring
```

## Integration testing

Integration tests rely on Testcontainers for Kafka, PostgreSQL, and Redis/Valkey. Prefer verifying full-service flows against the local Tilt + Argo stack; run targeted Go tests when they add meaningful coverage.
