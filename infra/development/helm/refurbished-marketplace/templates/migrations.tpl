{{- range $name, $svc := .Values.services }}
{{- if and $svc.enabled $svc.migration.enabled }}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ printf "%s-migrate" $name }}
  namespace: {{ $svc.namespace }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
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
              until pg_isready -h {{ $svc.db.host }} -p {{ $svc.db.port }} -U {{ $svc.db.user }} -d {{ $svc.db.name }};
              do echo "waiting for database {{ $svc.db.host }}"; sleep 2; done
      containers:
        - name: goose
          image: {{ $svc.migration.image }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          env:
            - name: GOOSE_DRIVER
              value: "postgres"
            - name: GOOSE_DBSTRING
              value: {{ printf "host=%s port=%v user=%s password=%s dbname=%s sslmode=disable" $svc.db.host $svc.db.port $svc.db.user $svc.db.password $svc.db.name | quote }}
            - name: GOOSE_COMMAND
              value: "up"
{{- end }}
{{- end }}
