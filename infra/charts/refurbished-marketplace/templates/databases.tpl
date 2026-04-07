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
  managed:
    roles:
      - name: {{ $svc.db.owner | quote }}
        login: true
        replication: true  # Required for Debezium CDC
        passwordSecret:
          name: {{ $svc.db.secretName }}
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
