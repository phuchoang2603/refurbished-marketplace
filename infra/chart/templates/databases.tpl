{{- range $name, $db := .Values.databases }}
{{- if $db.enabled }}
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ printf "%s-db" $name }}
  namespace: {{ $.Release.Namespace }}
spec:
  instances: {{ $db.instances }}
  storage:
    size: {{ $db.storage.size | quote }}
  bootstrap:
    initdb:
      database: {{ $db.dbName | quote }}
      owner: {{ $db.owner | quote }}
      secret:
        name: {{ $db.appSecretName }}
{{- end }}
{{- end }}
