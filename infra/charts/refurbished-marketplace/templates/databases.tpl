{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.db }}
{{- $dbResources := default $.Values.defaults.dbResources $svc.db.resources }}
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ printf "%s-db" $name }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  {{- with $.Values.databaseNodeSelector }}
  affinity:
    nodeSelector:
      {{- toYaml . | nindent 6 }}
  {{- end }}
  instances: {{ default 1 $svc.db.instances }}
  postgresql:
    parameters:
      wal_level: "logical" # Required for Debezium CDC
  managed:
    roles:
      - name: {{ $svc.db.owner | quote }}
        login: true
        replication: true  # Required for Debezium CDC
        passwordSecret:
          name: {{ $svc.db.secretName }}
      - name: {{ $svc.db.runtimeRole | quote }}
        login: true
        passwordSecret:
          name: {{ $svc.db.runtimeSecretName }}
  storage:
    size: {{ default "512Mi" $svc.db.storageSize | quote }}
{{- with $dbResources }}
  resources:
{{ toYaml . | nindent 4 }}
{{- end }}
  bootstrap:
    initdb:
      database: {{ $svc.db.name | quote }}
      owner: {{ default (printf "%s_app" $name) $svc.db.owner | quote }}
      secret:
        name: {{ $svc.db.secretName }}
{{- end }}
{{- end }}
