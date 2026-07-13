{{- if .Values.ingress.enabled }}
{{- $gatewayName := default "ecommerce-ingress" .Values.ingress.name }}
{{- $webHost := required "ingress.webHostname is required when ingress.enabled is true" .Values.ingress.webHostname }}
{{- $simHost := required "ingress.simulatorHostname is required when ingress.enabled is true" .Values.ingress.simulatorHostname }}
{{- $port := default 80 .Values.ingress.port }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ $gatewayName }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "5"
    # In-cluster cloudflared uses ClusterIP; avoid allocating a MetalLB address.
    networking.istio.io/service-type: ClusterIP
spec:
  gatewayClassName: istio
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
  name: web
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  parentRefs:
    - name: {{ $gatewayName }}
  hostnames:
    - {{ $webHost | quote }}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: web
          port: {{ index .Values.services "web" "port" }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: payment-gateway-simulator
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  parentRefs:
    - name: {{ $gatewayName }}
  hostnames:
    - {{ $simHost | quote }}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: payment-gateway-simulator
          port: {{ index .Values.services "payment-gateway-simulator" "port" }}
{{- end }}
