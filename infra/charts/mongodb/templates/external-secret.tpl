{{- if .Values.externalSecrets.enabled }}
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ .Values.user.passwordSecretName }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  refreshInterval: {{ .Values.externalSecrets.refreshInterval | quote }}
  secretStoreRef:
    kind: {{ .Values.externalSecrets.secretStoreRef.kind }}
    name: {{ .Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ .Values.user.passwordSecretName }}
    creationPolicy: Owner
  data:
    - secretKey: {{ .Values.user.passwordSecretKey }}
      remoteRef:
        key: {{ .Values.externalSecrets.remoteKey }}
{{- end }}
