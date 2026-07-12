{{- if .Values.istioScrapes.enabled }}
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: istio-istiod
  namespace: {{ .Release.Namespace }}
spec:
  namespaceSelector:
    matchNames:
      - istio-system
  selector:
    matchLabels:
      app: istiod
  podMetricsEndpoints:
    - port: http-monitoring
      path: /metrics
---
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: istio-ztunnel
  namespace: {{ .Release.Namespace }}
spec:
  namespaceSelector:
    matchNames:
      - istio-system
  selector:
    matchLabels:
      app: ztunnel
  podMetricsEndpoints:
    - port: ztunnel-stats
      path: /metrics
---
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMPodScrape
metadata:
  name: istio-cni
  namespace: {{ .Release.Namespace }}
spec:
  namespaceSelector:
    matchNames:
      - istio-system
  selector:
    matchLabels:
      k8s-app: istio-cni-node
  podMetricsEndpoints:
    - port: metrics
      path: /metrics
---
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
{{- end }}
