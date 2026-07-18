{{- /*
The destination Namespace is intentionally NOT templated here. Owning the release
namespace inside the release causes prune+recreate churn (Argo CreateNamespace /
Tilt namespace management fight the manifest), which cascade-deletes workloads.
The deployer owns it instead: Argo via syncPolicy.managedNamespaceMetadata, Tilt
via an out-of-band `kubectl apply`. Both apply the ambient/waypoint labels below.
*/}}
{{- if .Values.mesh.waypoint.enabled }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ default "ecommerce-waypoint" .Values.mesh.waypoint.name }}
  namespace: {{ .Release.Namespace }}
  labels:
    istio.io/waypoint-for: service
  annotations:
    argocd.argoproj.io/sync-wave: "5"
spec:
  gatewayClassName: istio-waypoint
  listeners:
    - name: mesh
      port: 15008
      protocol: HBONE
{{- end }}
{{- /*
Waypoint/ingress Envoy OTEL exporters dial VictoriaTraces via an Istio outbound
cluster. Without this, ambient may select ISTIO_MUTUAL and every OTLP export
fails (rq_error, rq_success=0) against VT's plaintext gRPC (:4317).
*/}}
{{- if and .Values.mesh.tracing.enabled .Values.defaults.otel.endpoint }}
---
apiVersion: networking.istio.io/v1
kind: DestinationRule
metadata:
  name: vtsingle-otlp-plain
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  host: vtsingle-vmks.monitoring.svc.cluster.local
  trafficPolicy:
    tls:
      mode: DISABLE
{{- end }}
