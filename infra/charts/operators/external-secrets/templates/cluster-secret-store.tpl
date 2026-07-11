{{- if .Values.clusterSecretStore.enabled }}
apiVersion: external-secrets.io/v1
kind: ClusterSecretStore
metadata:
  name: {{ .Values.clusterSecretStore.name }}
spec:
  provider:
    doppler:
      auth:
        secretRef:
          dopplerToken:
            name: {{ .Values.clusterSecretStore.dopplerToken.secretName }}
            key: {{ .Values.clusterSecretStore.dopplerToken.secretKey }}
            namespace: {{ .Values.clusterSecretStore.dopplerToken.namespace }}
{{- end }}
