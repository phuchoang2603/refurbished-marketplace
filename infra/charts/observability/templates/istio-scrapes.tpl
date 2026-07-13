{{- if .Values.istioScrapes.enabled }}
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: istio-waypoint
  namespace: {{ .Release.Namespace }}
spec:
  namespaceSelector:
    matchNames:
{{- range .Values.istioScrapes.waypointNamespaces }}
      - {{ . | quote }}
{{- end }}
  selector:
    matchLabels:
      gateway.networking.k8s.io/gateway-name: {{ .Values.istioScrapes.waypointName | quote }}
  podMetricsEndpoints:
    - port: http-envoy-prom
      path: /stats/prometheus
    - port: metrics
      path: /stats/prometheus
---
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: istio-ingress
  namespace: {{ .Release.Namespace }}
spec:
  namespaceSelector:
    matchNames:
{{- range .Values.istioScrapes.ingressNamespaces }}
      - {{ . | quote }}
{{- end }}
  selector:
    matchLabels:
      gateway.networking.k8s.io/gateway-name: {{ .Values.istioScrapes.ingressName | quote }}
  podMetricsEndpoints:
    - port: http-envoy-prom
      path: /stats/prometheus
    - port: metrics
      path: /stats/prometheus
{{- end }}
