{{- range $name, $svc := .Values.services }}
{{- if $svc.enabled }}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ $svc.namespace }}
{{- end }}
{{- end }}
