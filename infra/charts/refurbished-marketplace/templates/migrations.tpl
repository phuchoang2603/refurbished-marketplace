{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.migration $svc.migration.enabled }}
{{- $owner := default (printf "%s_app" $name) $svc.db.owner }}
{{- $initResources := default $.Values.defaults.initResources $svc.initResources }}
{{- $migrationResources := default $.Values.defaults.migrationResources $svc.migration.resources }}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ printf "%s-migrate" $name }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/hook: Sync
    argocd.argoproj.io/sync-wave: "4"
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 300
  template:
    metadata:
      labels:
        app: {{ printf "%s-migrate" $name }}
    spec:
      restartPolicy: OnFailure
      initContainers:
        - name: wait-for-db
          image: postgres:16-alpine
          command: ["sh", "-c"]
          args:
            - >-
              until pg_isready -h {{ $svc.db.host }} -p {{ $svc.db.port }};
              do echo "waiting for database {{ $svc.db.host }}"; sleep 2; done
{{- with $initResources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
      containers:
        - name: goose
          image: {{ include "refurbished-marketplace.image" (list $ $svc.migration.image $svc.migration.imageTag) }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          command: ["/bin/goose"]
          args:
            - -dir=/migrations
            - -env=none
            - up
{{- with $migrationResources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
          env:
            - name: GOOSE_DRIVER
              value: "postgres"
            - name: PGUSER
              value: {{ $owner | quote }}
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.db.secretName }}
                  key: {{ $svc.db.passwordKey }}
            - name: PGHOST
              value: {{ $svc.db.host | quote }}
            - name: PGPORT
              value: {{ $svc.db.port | quote }}
            - name: PGDATABASE
              value: {{ $svc.db.name | quote }}
            - name: PGSSLMODE
              value: "disable"
            - name: GOOSE_DBSTRING
              value: "sslmode=disable"
{{- end }}
{{- end }}
