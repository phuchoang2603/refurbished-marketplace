{{- if .Values.externalSecrets.enabled }}
{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.db }}
{{- range $credential := list (dict "name" $svc.db.secretName "user" (default (printf "%s_app" $name) $svc.db.owner)) (dict "name" $svc.db.runtimeSecretName "user" $svc.db.runtimeRole) }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ $credential.name }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  refreshInterval: {{ default "1h" $.Values.externalSecrets.refreshInterval }}
  secretStoreRef:
    kind: {{ $.Values.externalSecrets.secretStoreRef.kind }}
    name: {{ $.Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ $credential.name }}
    creationPolicy: Owner
    template:
      metadata:
        labels:
          cnpg.io/reload: "true"
      type: kubernetes.io/basic-auth
      engineVersion: v2
      data:
        username: {{ $credential.user | quote }}
        password: "{{`{{ .password }}`}}"
  data:
    - secretKey: password
      remoteRef:
        key: {{ include "refurbished-marketplace.dopplerKeyPrefix" $credential.name }}_PASSWORD
{{- end }}
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
{{- with .Values.services.checkout }}
{{- if and .enabled .paymentGateway }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ .paymentGateway.secretName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  refreshInterval: {{ default "1h" $.Values.externalSecrets.refreshInterval }}
  secretStoreRef:
    kind: {{ $.Values.externalSecrets.secretStoreRef.kind }}
    name: {{ $.Values.externalSecrets.secretStoreRef.name }}
  target:
    name: {{ .paymentGateway.secretName }}
    creationPolicy: Owner
  data:
{{- range $key := list "PAYMENT_GATEWAY_API_KEY" "PAYMENT_GATEWAY_WEBHOOK_SECRET" }}
    - secretKey: {{ $key }}
      remoteRef:
        key: {{ $key }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
