# Istio observe baseline

Staging installs Istio ambient mode through four wrapper charts under `infra/charts/operators/istio/`, pinned to official Istio `1.30.2`.

## Components

| Wrapper   | Upstream chart                      | Argo CD Application  | Sync wave | Namespace      |
| --------- | ----------------------------------- | -------------------- | --------- | -------------- |
| `base`    | `istio/base`                        | `staging-istio-base` | 0         | `istio-system` |
| `istiod`  | `istio/istiod` (`profile=ambient`)  | `staging-istiod`     | 1         | `istio-system` |
| `cni`     | `istio/cni` (`profile=ambient`)     | `staging-istio-cni`  | 1         | `istio-system` |
| `ztunnel` | `istio/ztunnel` (`profile=ambient`) | `staging-ztunnel`    | 2         | `istio-system` |

Marketplace enrollment and Kafka follow at waves **3** and **4**.

## Marketplace enrollment

Staging enables ambient + waypoint via Helm values on `staging-refurbished-marketplace`:

```yaml
mesh:
  ambient:
    enabled: true
  waypoint:
    enabled: true
```

That labels the `ecommerce` namespace with `istio.io/dataplane-mode=ambient` and creates an `istio-waypoint` Gateway for L7 telemetry.

Kafka/Connect/UI live in the separate `kafka` namespace (not ambient-enrolled) so Strimzi TLS and Debezium are not intercepted by ztunnel/waypoint. Marketplace services reach brokers at `ecommerce-kafka-cluster-kafka-bootstrap.kafka.svc:9092`.

**Tilt defaults keep `mesh.ambient.enabled: false`** so local pods are not redirected into ztunnel.

## Protocol-aware Service ports

Marketplace Services render port names from `services.<name>.protocol`:

- HTTP: `web`, `payment-gateway-simulator`
- gRPC: `users`, `products`, `orders`, `cart`, `payment`

## Rollback

1. Disable enrollment first — set `mesh.ambient.enabled: false` and `mesh.waypoint.enabled: false` on the staging marketplace Application (or remove those values), then sync.
2. Restart marketplace pods so they leave ambient redirection.
3. Only then prune/disable the Istio Applications (`staging-ztunnel`, `staging-istio-cni`, `staging-istiod`, `staging-istio-base`) if you need to remove the mesh platform.

Do not delete Istio while `ecommerce` is still labeled for ambient mode.

## Canal / NetworkPolicy note

Staging uses RKE2 Canal. Ambient HBONE uses TCP **15008**. Any allow-list NetworkPolicy that covers mesh traffic must permit that port.

## Verification

```bash
kubectl get pods -n istio-system
kubectl get daemonset -n istio-system
kubectl get ns ecommerce --show-labels
kubectl get gateway -n ecommerce
kubectl get pods -n kafka
```

Then exercise web → product → cart → checkout → payment flows and inspect Grafana / VictoriaTraces (see [observability.md](../observability.md)).

## Production

Production Istio Applications and marketplace mesh enrollment are intentionally omitted until staging is verified.
