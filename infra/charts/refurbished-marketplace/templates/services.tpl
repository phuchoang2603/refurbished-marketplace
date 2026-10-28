{{- range $name, $svc := .Values.services }}
{{- if $svc.enabled }}
{{- $owner := "" }}
{{- if $svc.db }}
{{- $owner = default (printf "%s_app" $name) $svc.db.owner }}
{{- end }}
{{- $resources := default $.Values.defaults.resources $svc.resources }}
{{- $initResources := default $.Values.defaults.initResources $svc.initResources }}
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
  replicas: {{ default 1 $svc.replicas }}
  selector:
    matchLabels:
      app: {{ $name }}
  template:
    metadata:
      labels:
        app: {{ $name }}
        app.kubernetes.io/name: {{ $name }}
      annotations:
        # Read by the platform agent's k8sattributes for stdout log records.
        resource.opentelemetry.io/service.name: {{ $name }}
    spec:
{{- if or $svc.db $svc.meili }}
      initContainers:
{{- if $svc.db }}
        - name: migrate
          image: {{ include "refurbished-marketplace.image" (list $ $svc.image) }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          args: [migrate]
          env:
{{ include "refurbished-marketplace.postgresEnv" (dict "service" $svc "user" $owner "secret" $svc.db.secretName) | nindent 12 }}
            - name: PG_RUNTIME_ROLE
              value: {{ $svc.db.runtimeRole | quote }}
          resources:
{{ toYaml ($svc.migrationResources | default $.Values.defaults.migrationResources) | nindent 12 }}
{{- end }}
{{- if $svc.meili }}
        - name: wait-for-meili
          image: busybox:1.37
          command: ["sh", "-c"]
          args:
            - >-
              until wget -q -O /dev/null http://{{ $svc.meili.host }}:{{ $svc.meili.port }}/health;
              do echo "waiting for meilisearch {{ $svc.meili.host }}"; sleep 2; done
{{- with $initResources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
{{- end }}
{{- end }}
      containers:
{{- range $sidecar := $svc.sidecars }}
        - name: {{ $sidecar.name }}
          image: {{ $sidecar.image }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
{{- with $sidecar.port }}
          ports:
            - containerPort: {{ . }}
{{- end }}
{{- with ($sidecar.resources | default $.Values.defaults.sidecarResources) }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
{{- end }}
        - name: {{ $name }}
          image: {{ include "refurbished-marketplace.image" (list $ $svc.image) }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            - name: {{ default "http" $svc.protocol }}
              containerPort: {{ $svc.port }}
{{- with $resources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
          env:
{{- if $svc.db }}
{{ include "refurbished-marketplace.postgresEnv" (dict "service" $svc "user" $svc.db.runtimeRole "secret" $svc.db.runtimeSecretName) | nindent 12 }}
{{- end }}
{{- if $svc.meili }}
            - name: MEILI_URL
              value: {{ printf "http://%s:%v" $svc.meili.host $svc.meili.port | quote }}
            - name: MEILI_MASTER_KEY
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.meili.secretName }}
                  key: {{ default "MEILI_MASTER_KEY" $svc.meili.secretKey }}
{{- end }}
{{- if $svc.auth }}
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.auth.secretName }}
                  key: {{ $svc.auth.secretKey }}
{{- end }}
{{- if and (eq $name "checkout") $svc.paymentGateway }}
{{- range $key := list "PAYMENT_GATEWAY_API_KEY" "PAYMENT_GATEWAY_WEBHOOK_SECRET" }}
            - name: {{ $key }}
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.paymentGateway.secretName }}
                  key: {{ $key }}
{{- end }}
{{- end }}
{{- with $.Values.defaults.otel }}
            - name: OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
              value: {{ .metricsEndpoint | quote }}
            - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
              value: {{ .tracesEndpoint | quote }}
            - name: OTEL_RESOURCE_ATTRIBUTES
              value: {{ printf "k8s.cluster.name=%s,deployment.environment=%s" $.Values.global.clusterName $.Values.global.environment | quote }}
            - name: OTEL_SERVICE_NAME
              value: {{ $name | quote }}
            - name: OTEL_TRACES_SAMPLER_ARG
              value: "1"
            - name: OTEL_METRIC_EXPORT_INTERVAL
              value: "30000"
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
{{- if and $.Values.mesh.waypoint.enabled (has $name $.Values.mesh.resilience.services) }}
    istio.io/use-waypoint: {{ $.Values.mesh.waypoint.name }}
{{- end }}
spec:
  selector:
    app: {{ $name }}
  ports:
    - name: {{ default "http" $svc.protocol }}
      port: {{ $svc.port }}
      targetPort: {{ $svc.port }}
{{- end }}
{{- end }}
