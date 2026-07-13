{{- range $name, $svc := .Values.services }}
{{- if $svc.enabled }}
{{- $owner := "" }}
{{- if $svc.db }}
{{- $owner = default (printf "%s_app" $name) $svc.db.owner }}
{{- end }}
{{- $resources := default $.Values.defaults.resources $svc.resources }}
{{- $initResources := default $.Values.defaults.initResources $svc.initResources }}
{{- $redisResources := default $.Values.defaults.redisResources $svc.redisResources }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "5"
  labels:
    app: {{ $name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ $name }}
  template:
    metadata:
      labels:
        app: {{ $name }}
    spec:
{{- if $svc.db }}
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
{{- end }}
      containers:
{{- if eq $name "cart" }}
        - name: redis
          image: docker.io/valkey/valkey:7.2.5
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            - containerPort: 6379
{{- with $redisResources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
{{- end }}
        - name: {{ $name }}
          image: {{ include "refurbished-marketplace.image" (list $ $svc.image $svc.imageTag) }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            - containerPort: {{ $svc.port }}
{{- with $resources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
          env:
{{- if $svc.db }}
            - name: DB_USER
              value: {{ $owner | quote }}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.db.secretName }}
                  key: {{ $svc.db.passwordKey }}
            - name: DB_URL
              value: {{ printf "postgres://$(DB_USER):$(DB_PASSWORD)@%s:%v/%s?sslmode=disable" $svc.db.host $svc.db.port $svc.db.name | quote }}
{{- end }}
{{- if $svc.auth }}
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.auth.secretName }}
                  key: {{ $svc.auth.secretKey }}
{{- end }}
{{- with $.Values.defaults.otel.endpoint }}
            - name: OTEL_EXPORTER_OTLP_ENDPOINT
              value: {{ . | quote }}
            - name: OTEL_SERVICE_NAME
              value: {{ $name | quote }}
            - name: OTEL_TRACES_SAMPLER_ARG
              value: "1"
{{- end }}
{{- if $svc.env }}
{{- range $key, $value := $svc.env }}
            - name: {{ $key }}
              value: {{ $value | quote }}
{{- end }}
{{- end }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    app: {{ $name }}
spec:
  selector:
    app: {{ $name }}
  ports:
    - name: {{ default "http" $svc.protocol }}
      port: {{ $svc.port }}
      targetPort: {{ $svc.port }}
{{- end }}
{{- end }}
