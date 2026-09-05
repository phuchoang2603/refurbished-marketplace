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
  description: Allow web and kubelet to {{ $name }} gRPC. Unknown identities are denied when enforce is true.
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
---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-metrics-scrape
  namespace: {{ $ns }}
  annotations:
    argocd.argoproj.io/sync-wave: "7"
spec:
  description: VMAgent in monitoring scrapes /metrics without SPIRE mTLS.
  endpointSelector:
    matchLabels:
      marketplace.metrics: "true"
  ingress:
    - fromEntities:
        - host
      toPorts:
        - ports:
            - port: "9100"
              protocol: TCP
    - fromEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: monitoring
      toPorts:
        - ports:
            - port: "9100"
              protocol: TCP
{{- end }}
