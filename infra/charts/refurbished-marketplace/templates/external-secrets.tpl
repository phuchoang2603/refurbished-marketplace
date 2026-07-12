{{- if .Values.externalSecrets.enabled }}
{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.db }}
{{- $prefix := include "refurbished-marketplace.dopplerKeyPrefix" $svc.db.secretName }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ $svc.db.secretName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  refreshInterval: {{ default "1h" $.Values.externalSecrets.refreshInterval }}
  secretStoreRef:
    kind: {{ $.Values.externalSecrets.secretStoreRef.kind }}
    name: {{ $.Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ $svc.db.secretName }}
    creationPolicy: Owner
    template:
      type: kubernetes.io/basic-auth
      engineVersion: v2
      data:
        username: "{{`{{ .username }}`}}"
        password: "{{`{{ .password }}`}}"
  data:
    - secretKey: username
      remoteRef:
        key: {{ $prefix }}_USERNAME
    - secretKey: password
      remoteRef:
        key: {{ $prefix }}_PASSWORD
{{- end }}
{{- end }}
{{- $authSecrets := dict }}
{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.auth }}
{{- $_ := set $authSecrets $svc.auth.secretName $svc.auth.secretKey }}
{{- end }}
{{- end }}
{{- range $secretName, $secretKey := $authSecrets }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ $secretName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  refreshInterval: {{ default "1h" $.Values.externalSecrets.refreshInterval }}
  secretStoreRef:
    kind: {{ $.Values.externalSecrets.secretStoreRef.kind }}
    name: {{ $.Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ $secretName }}
    creationPolicy: Owner
  data:
    - secretKey: {{ $secretKey }}
      remoteRef:
        key: {{ $secretKey }}
{{- end }}
{{- end }}
