{{- if .Values.externalSecrets.enabled }}
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ .Values.tunnel.existingSecret }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  refreshInterval: {{ .Values.externalSecrets.refreshInterval | quote }}
  secretStoreRef:
    kind: {{ .Values.externalSecrets.secretStoreRef.kind }}
    name: {{ .Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ .Values.tunnel.existingSecret }}
    creationPolicy: Owner
  data:
    - secretKey: {{ .Values.tunnel.existingSecretKey }}
      remoteRef:
        key: {{ .Values.externalSecrets.remoteKey }}
{{- end }}
