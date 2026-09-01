{{- if .Values.ingress.enabled }}
{{- $gatewayName := default "grafana" .Values.ingress.name }}
{{- $host := required "ingress.hostname is required when ingress.enabled is true" .Values.ingress.hostname }}
{{- $port := default 80 .Values.ingress.port }}
{{- $backend := default "observability-grafana" .Values.ingress.backendService }}
{{- $backendPort := default 80 .Values.ingress.backendPort }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ $gatewayName }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "5"
spec:
  gatewayClassName: cilium
  listeners:
    - name: http
      port: {{ $port }}
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: Same
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grafana
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  parentRefs:
    - name: {{ $gatewayName }}
  hostnames:
    - {{ $host | quote }}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      # TLS terminates at Cloudflare; origin is HTTP.
      # cilium-gateway-{{ $gatewayName }}.{{ .Release.Namespace }}.svc.cluster.local:{{ $port }}
      filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            set:
              - name: X-Forwarded-Proto
                value: https
              - name: X-Forwarded-Host
                value: {{ $host | quote }}
      backendRefs:
        - name: {{ $backend }}
          port: {{ $backendPort }}
{{- end }}
