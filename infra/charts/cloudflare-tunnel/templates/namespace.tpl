apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-options: Prune=false
  labels:
    app.kubernetes.io/name: cloudflare-tunnel
