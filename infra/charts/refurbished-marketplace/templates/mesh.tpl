---
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: default
  namespace: {{ .Release.Namespace }}
spec:
  mtls:
    mode: STRICT
{{- with .Values.mesh.waypoint }}
{{- if .enabled }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .name }}-options
  namespace: {{ $.Release.Namespace }}
data:
  deployment: |
    spec:
      replicas: {{ .replicas }}
{{- with .resources }}
      template:
        spec:
          containers:
            - name: istio-proxy
              resources:
{{- toYaml . | nindent 16 }}
{{- end }}
{{- if gt (int .replicas) 1 }}
  podDisruptionBudget: |
    spec:
      minAvailable: 1
{{- end }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ .name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    istio.io/waypoint-for: service
  annotations:
    argocd.argoproj.io/sync-wave: "4"
spec:
  gatewayClassName: istio-waypoint
  infrastructure:
    parametersRef:
      group: ""
      kind: ConfigMap
      name: {{ .name }}-options
  listeners:
    - name: mesh
      port: 15008
      protocol: HBONE
{{- end }}
{{- end }}
