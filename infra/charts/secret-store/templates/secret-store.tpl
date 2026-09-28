apiVersion: external-secrets.io/v1
kind: SecretStore
metadata:
  name: doppler
  namespace: {{ .Release.Namespace }}
spec:
  provider:
    doppler:
      auth:
        secretRef:
          dopplerToken:
            name: {{ .Values.tokenSecret.name }}
            key: {{ .Values.tokenSecret.key }}
