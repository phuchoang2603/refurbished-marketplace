{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.db }}
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ printf "%s-db" $name }}
  namespace: {{ $.Release.Namespace }}
spec:
  instances: {{ default 1 $svc.db.instances }}
  storage:
    size: {{ default "1Gi" $svc.db.storageSize | quote }}
  bootstrap:
    initdb:
      database: {{ $svc.db.name | quote }}
      owner: {{ default (printf "%s_app" $name) $svc.db.owner | quote }}
      secret:
        name: {{ $svc.db.secretName }}
{{- end }}
{{- end }}
