{{- if and .Values.mesh.ambient.enabled .Values.mesh.tracing.enabled }}
---
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: ecommerce-tracing
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  # Ambient L7 Gateways need targetRefs (namespace-scoped Telemetry does not
  # attach to ingress/waypoint the way sidecar Telemetry does). Both Gateways
  # share the same provider + sampling, so one CR is enough.
  targetRefs:
    - kind: Gateway
      group: gateway.networking.k8s.io
      name: {{ default "ecommerce-ingress" .Values.ingress.name }}
{{- if .Values.mesh.waypoint.enabled }}
    - kind: Gateway
      group: gateway.networking.k8s.io
      name: {{ default "ecommerce-waypoint" .Values.mesh.waypoint.name }}
{{- end }}
  tracing:
    - providers:
        - name: otel-vt
      randomSamplingPercentage: {{ .Values.mesh.tracing.samplingPercentage | default 100 }}
{{- end }}
