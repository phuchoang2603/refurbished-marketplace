## Why

The recreated Talos clusters keep Cilium for CNI but use Istio for ambient mesh and Gateway API ingress; this repository still renders the disabled Cilium Gateway, Cilium authentication policies, and pre-saga no-retry rules. Adopt Istio's native resilience controls without duplicating the durable checkout saga or replaying unsafe browser mutations.

## What Changes

- **BREAKING** Replace marketplace Cilium Gateway/mesh policies with Istio ingress, declarative `ecommerce` and `kafka` ambient enrollment, namespace-scoped strict mTLS, and a selected-service waypoint. Keep Cloudflare Tunnel as the public HTTPS edge and keep Cilium platform-owned as the CNI.
- Configure native Istio traffic controls for HTTP/gRPC Services: route timeouts, bounded retries for verified replay-safe operations, connection/request limits, retry budgets, and endpoint outlier handling only where multiple healthy endpoints exist. Preserve saga deadlines, transactional inbox/outbox replay, and explicit buyer intent identity.
- Remove blanket pre-saga retry bans, but prohibit indiscriminate retries of cart increments, browser checkout POSTs, hosted payment submissions, and callbacks. Correct Web's callback-triggered cart removal before treating callback replays as safe.
- Remove unused pre-saga payment-session creation, non-saga callback and expiry branches, obsolete Cilium policy templates/values, and stale networking specs/docs/comments. No data migration or compatibility path is required for the recreated clusters.
- Coordinate an independently owned `talos-proxmox` platform change for Gateway API Experimental retry CRDs, explicit Istio default retry behavior, and mesh proxy telemetry; this repo does not install platform components.

## Capabilities

### New Capabilities

- `istio-ambient-networking`: Istio ingress, Cloudflare origin, ambient enrollment, mTLS, and networking ownership.
- `istio-resilience`: Waypoint-scoped circuit breaking, timeouts, retry safety, budgets, and verification signals.

### Modified Capabilities

- `cilium-ingress`: Retire all obsolete Cilium Gateway requirements in favor of Istio ingress.
- `cilium-mesh-policy`: Retire Cilium mTLS, allow-lists, and the pre-saga blanket retry prohibition.
- `cilium-observability`: Retire post-Istio telemetry and no-enrollment requirements; retain application OTLP behavior under its existing capabilities.
- `argocd-gitops`: Own ambient namespace labels declaratively and render Istio rather than Cilium consumer resources.
- `payment`: Remove pre-saga session creation and expiration obligations while preserving saga-owned payment finality.
- `cart`: Forbid callback-driven cart removal before the saga reaches a paid order, including failed/expired outcomes.

## Impact

Changes marketplace Helm charts, app-of-apps namespace ownership, selected Web/Payment/Cart cleanup, OpenSpec specs, and deployment/architecture docs. Requires separate platform-owned changes under `../talos-proxmox` before enabling retry fields in marketplace manifests. Only production requires live deployment and acceptance checks; retain dev chart defaults and dev/prod render checks without deploying to dev. Does not replace Cilium as CNI, introduce a bespoke retry library, migrate cluster/data state, automatically retry Kafka messages through Istio, or enroll platform namespaces in ambient.
