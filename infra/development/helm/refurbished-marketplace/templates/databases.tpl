{{- range $name, $db := .Values.databases }}
{{- if $db.enabled }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ printf "%s-app" $name }}
  namespace: {{ $db.namespace }}
type: kubernetes.io/basic-auth
stringData:
  username: {{ $db.owner | quote }}
  password: {{ $db.password | quote }}
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ printf "%s-db" $name }}
  namespace: {{ $db.namespace }}
spec:
  instances: {{ $db.instances }}
  storage:
    size: {{ $db.storage.size | quote }}
  bootstrap:
    initdb:
      database: {{ $db.dbName | quote }}
      owner: {{ $db.owner | quote }}
      secret:
        name: {{ printf "%s-app" $name }}
{{- end }}
{{- end }}
