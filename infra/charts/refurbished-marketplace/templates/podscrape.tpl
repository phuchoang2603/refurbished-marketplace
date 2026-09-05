{{- if .Values.metricsScrape.enabled }}
---
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: marketplace-apps
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
  labels:
    app.kubernetes.io/name: refurbished-marketplace
spec:
  jobLabel: app
  namespaceSelector:
    matchNames:
      - {{ .Release.Namespace }}
  selector:
    matchLabels:
      marketplace.metrics: "true"
  podMetricsEndpoints:
    - port: metrics
      path: /metrics
      scheme: http
{{- end }}
