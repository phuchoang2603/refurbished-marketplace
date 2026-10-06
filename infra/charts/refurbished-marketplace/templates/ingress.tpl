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
    # Only the Cloudflare tunnel reaches the origin, so it needs no LoadBalancer VIP.
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
{{- range $backend := list (dict "name" "web" "host" $webHost) (dict "name" "payment-gateway-simulator" "host" $simHost) }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: {{ $backend.name }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  parentRefs:
    - name: {{ $gatewayName }}
  hostnames:
    - {{ $backend.host | quote }}
  rules:
{{- range $rule := list $.Values.ingress.reads $.Values.ingress.writes }}
    - matches:
        - path:
            type: PathPrefix
            value: /
{{- with $rule.method }}
          method: {{ . }}
{{- end }}
      # TLS terminates at Cloudflare; origin is HTTP.
      filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            set:
              - name: X-Forwarded-Proto
                value: https
              - name: X-Forwarded-Host
                value: {{ $backend.host | quote }}
{{- with $rule.timeouts }}
      timeouts:
{{- toYaml . | nindent 8 }}
{{- end }}
{{- with $rule.retry }}
      retry:
{{- toYaml . | nindent 8 }}
{{- end }}
      backendRefs:
        - name: {{ $backend.name }}
          port: {{ index $.Values.services $backend.name "port" }}
{{- end }}
{{- end }}
{{- $origin := printf "http://%s-istio.%s.svc.cluster.local:%v" $gatewayName .Release.Namespace $port }}
{{- with .Values.ingress.tunnel }}
{{- if .enabled }}
---
# The talos-proxmox Cloudflare operator routes these hostnames through its ClusterTunnel to the
# Gateway Service Istio creates, and owns their DNS records. The target is explicit because that
# Service also exposes the Istio status port.
apiVersion: networking.cfargotunnel.com/v1alpha1
kind: TunnelBinding
metadata:
  name: {{ $gatewayName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
subjects:
  - name: {{ $gatewayName }}-istio
    spec:
      fqdn: {{ $webHost | quote }}
      target: {{ $origin }}
  - name: {{ $gatewayName }}-istio
    spec:
      fqdn: {{ $simHost | quote }}
      target: {{ $origin }}
tunnelRef:
  kind: ClusterTunnel
  name: {{ required "ingress.tunnel.clusterTunnel is required when ingress.tunnel.enabled is true" .clusterTunnel }}
{{- end }}
{{- end }}
{{- end }}
