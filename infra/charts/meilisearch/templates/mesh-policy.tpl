apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-{{ .Values.name }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  description: Search and kubelet to Meilisearch 7700. No Cilium mTLS.
  endpointSelector:
    matchLabels:
      app: {{ .Values.name }}
{{- if .Values.meshPolicy.enforce }}
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
            - port: "7700"
              protocol: TCP
    - fromEndpoints:
        - matchLabels:
            app: search
      toPorts:
        - ports:
            - port: "7700"
              protocol: TCP
