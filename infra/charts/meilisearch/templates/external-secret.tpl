{{- if .Values.externalSecrets.enabled }}
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ .Values.externalSecrets.secretName }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
spec:
  refreshInterval: {{ .Values.externalSecrets.refreshInterval | quote }}
  secretStoreRef:
    kind: {{ .Values.externalSecrets.secretStoreRef.kind }}
    name: {{ .Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ .Values.externalSecrets.secretName }}
    creationPolicy: Owner
    template:
      engineVersion: v2
      data:
        MEILI_MASTER_KEY: "{{`{{ .password }}`}}"
  data:
    - secretKey: password
      remoteRef:
        key: {{ .Values.externalSecrets.remoteKey }}
{{- end }}
