{{- if .Values.meshPolicy.enabled }}
{{- $mutual := .Values.meshPolicy.mutualAuth }}
{{- $ns := .Release.Namespace }}
---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-web
  namespace: {{ $ns }}
  annotations:
    argocd.argoproj.io/sync-wave: "7"
spec:
  description: Shop HTTP from Cilium Gateway and kubelet; hosted-payment callbacks from the simulator.
  endpointSelector:
    matchLabels:
      app: web
{{- if .Values.meshPolicy.enforce }}
  enableDefaultDeny:
    ingress: true
{{- else }}
  enableDefaultDeny:
    ingress: false
{{- end }}
  ingress:
    - fromEntities:
        - ingress
        - host
      toPorts:
        - ports:
            - port: "8080"
              protocol: TCP
    - fromEndpoints:
        - matchLabels:
            app: payment-gateway-simulator
{{- if $mutual }}
      authentication:
        mode: required
{{- end }}
      toPorts:
        - ports:
            - port: "8080"
              protocol: TCP
{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled (eq (default "http" $svc.protocol) "grpc") }}
---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-{{ $name }}
  namespace: {{ $ns }}
  annotations:
    argocd.argoproj.io/sync-wave: "7"
spec:
  description: Allow only required callers to {{ $name }} gRPC. Unknown identities are denied when enforce is true.
  endpointSelector:
    matchLabels:
      app: {{ $name }}
{{- if $.Values.meshPolicy.enforce }}
  enableDefaultDeny:
    ingress: true
{{- else }}
  enableDefaultDeny:
    ingress: false
{{- end }}
  ingress:
    - fromEntities:
        - host
      toPorts:
        - ports:
            - port: {{ $svc.port | quote }}
              protocol: TCP
{{- if has $name (list "checkout" "orders" "inventory" "products" "search" "users" "cart" "payment") }}
    - fromEndpoints:
        - matchLabels:
            app: web
{{- if $mutual }}
      authentication:
        mode: required
{{- end }}
      toPorts:
        - ports:
            - port: {{ $svc.port | quote }}
              protocol: TCP
{{- end }}
{{- end }}
{{- end }}
---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-payment-gateway-simulator
  namespace: {{ $ns }}
  annotations:
    argocd.argoproj.io/sync-wave: "7"
spec:
  description: Hosted-payment simulator from Cilium Gateway and kubelet only.
  endpointSelector:
    matchLabels:
      app: payment-gateway-simulator
{{- if .Values.meshPolicy.enforce }}
  enableDefaultDeny:
    ingress: true
{{- else }}
  enableDefaultDeny:
    ingress: false
{{- end }}
  ingress:
    - fromEntities:
        - ingress
        - host
      toPorts:
        - ports:
            - port: "8097"
              protocol: TCP
{{- end }}
