#!/usr/bin/env bash
# Install Argo CD on the current cluster (Colima/k3s) and apply the local app-of-apps.
# Apps sync charts from the current git branch (override with ARGO_REVISION).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -n "${ARGO_REVISION:-}" ]]; then
  REVISION="$ARGO_REVISION"
else
  REVISION="$(git branch --show-current)"
fi

ARGO_NS="${ARGO_NS:-argo-cd}"
GATEWAY_API_VERSION="${GATEWAY_API_VERSION:-v1.3.0}"
DOPPLER_SECRET="${DOPPLER_SECRET:-infra/k8s/doppler-token.dev.secret.yaml}"

if [[ -z "$REVISION" || "$REVISION" == "HEAD" ]]; then
  echo "Detached HEAD or empty branch; set ARGO_REVISION to a pushable branch name." >&2
  exit 1
fi

if [[ ! -f "$DOPPLER_SECRET" ]]; then
  echo "Missing $DOPPLER_SECRET — copy from .example and paste the Doppler dev token." >&2
  exit 1
fi

echo "==> Argo will sync git revision: $REVISION"
echo "==> Push first if needed: git push -u origin HEAD"

kubectl get ns "$ARGO_NS" >/dev/null 2>&1 || kubectl create namespace "$ARGO_NS"

if ! kubectl get deploy -n "$ARGO_NS" argocd-server >/dev/null 2>&1; then
  echo "==> Installing Argo CD"
  kubectl apply -n "$ARGO_NS" -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
  kubectl -n "$ARGO_NS" rollout status deploy/argocd-server --timeout=5m
  kubectl -n "$ARGO_NS" rollout status deploy/argocd-repo-server --timeout=5m
fi

echo "==> Gateway API CRDs (waypoint)"
kubectl get crd gateways.gateway.networking.k8s.io >/dev/null 2>&1 || \
  kubectl apply --server-side -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/standard-install.yaml"

echo "==> Doppler bootstrap secret"
kubectl apply -f "$DOPPLER_SECRET"

echo "==> Applying local app-of-apps (targetRevision=$REVISION)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
for f in "$ROOT"/infra/argocd/local/root.yaml "$ROOT"/infra/argocd/local/apps/*.yaml; do
  sed "s/^    targetRevision: .*/    targetRevision: ${REVISION}/" "$f" > "$tmpdir/$(basename "$f")"
done
kubectl apply -n "$ARGO_NS" -f "$tmpdir"

echo "==> Applications (OutOfSync until $REVISION is on GitHub)"
kubectl -n "$ARGO_NS" get applications 2>/dev/null || true

echo
echo "Next:"
echo "  1. git push -u origin HEAD"
echo "  2. build-images"
echo "  3. Put CLOUDFLARE_TUNNEL_TOKEN in Doppler dev + Public Hostnames for"
echo "     shop.dev.phuchoang.sbs / pay.dev.phuchoang.sbs → ecommerce-ingress-istio"
echo "  4. open https://shop.dev.phuchoang.sbs"
echo "  Argo UI: kubectl -n $ARGO_NS port-forward svc/argocd-server 8088:443"
