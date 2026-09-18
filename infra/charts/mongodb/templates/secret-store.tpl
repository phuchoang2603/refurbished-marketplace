apiVersion: external-secrets.io/v1
kind: SecretStore
metadata:
  name: doppler
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  provider:
    doppler:
      auth:
        secretRef:
          dopplerToken:
            name: doppler-token
            key: dopplerToken
