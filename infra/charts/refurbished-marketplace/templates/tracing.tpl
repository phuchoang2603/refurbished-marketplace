{{- if and .Values.mesh.ambient.enabled .Values.mesh.tracing.enabled }}
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
  name: ecommerce-tracing
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  tracing:
    - providers:
        - name: otel-vt
      randomSamplingPercentage: {{ .Values.mesh.tracing.samplingPercentage | default 100 }}
{{- end }}
