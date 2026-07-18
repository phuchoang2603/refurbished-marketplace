{{- if and .Values.mesh.ambient.enabled .Values.mesh.tracing.enabled }}
---
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: ecommerce-tracing-ingress
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  targetRefs:
    - kind: Gateway
      group: gateway.networking.k8s.io
      name: {{ default "ecommerce-ingress" .Values.ingress.name }}
  tracing:
    - providers:
        - name: otel-vt
      randomSamplingPercentage: {{ .Values.mesh.tracing.samplingPercentage | default 100 }}
{{- if .Values.mesh.waypoint.enabled }}
---
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: ecommerce-tracing-waypoint
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  targetRefs:
    - kind: Gateway
      group: gateway.networking.k8s.io
      name: {{ default "ecommerce-waypoint" .Values.mesh.waypoint.name }}
  tracing:
    - providers:
        - name: otel-vt
      randomSamplingPercentage: {{ .Values.mesh.tracing.samplingPercentage | default 100 }}
{{- end }}
{{- end }}
