---
apiVersion: kafka.strimzi.io/v1
kind: KafkaConnect
metadata:
  name: {{ .Values.connect.clusterName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    strimzi.io/use-connector-resources: "true"
spec:
  version: {{ .Values.kafka.version }}
  image: {{ include "kafka.image" (list . .Values.connect.image .Values.connect.imageTag) }}
  replicas: 1
  bootstrapServers: {{ printf "%s-kafka-bootstrap:9092" .Values.kafka.clusterName }}
  groupId: {{ .Values.connect.clusterName }}
  offsetStorageTopic: {{ printf "%s-offsets" .Values.connect.clusterName }}
  configStorageTopic: {{ printf "%s-configs" .Values.connect.clusterName }}
  statusStorageTopic: {{ printf "%s-status" .Values.connect.clusterName }}
{{- with .Values.connect.jvmOptions }}
  jvmOptions:
{{ toYaml . | nindent 4 }}
{{- end }}
{{- with .Values.connect.resources }}
  resources:
{{ toYaml . | nindent 4 }}
{{- end }}
{{- with .Values.connect.livenessProbe }}
  livenessProbe:
{{ toYaml . | nindent 4 }}
{{- end }}
{{- with .Values.connect.readinessProbe }}
  readinessProbe:
{{ toYaml . | nindent 4 }}
{{- end }}
  # Activates Strimzi tracing-agent → initializes GlobalOpenTelemetry so Debezium
  # EventRouter can inject W3C traceparent into Kafka record headers.
  tracing:
    type: opentelemetry
  config:
    config.providers: secrets
    config.providers.secrets.class: io.strimzi.kafka.KubernetesSecretConfigProvider
    offset.storage.replication.factor: -1
    config.storage.replication.factor: -1
    status.storage.replication.factor: -1
    # Let connectors override their producer client config. Needed so Debezium
    # connectors can drop the Strimzi-injected TracingProducerInterceptor, which
    # overwrites the traceparent header EventRouter restored from the outbox row
    # (splitting the e2e TraceId at the Kafka hop).
    connector.client.config.override.policy: All
  template:
    connectContainer:
      env:
        - name: OTEL_SERVICE_NAME
          value: connect-debezium
        - name: OTEL_TRACES_EXPORTER
          value: otlp
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: {{ .Values.connect.otelEndpoint | default "http://vtsingle-vmks.monitoring.svc.cluster.local:4317" | quote }}
        - name: OTEL_EXPORTER_OTLP_PROTOCOL
          value: grpc
        - name: OTEL_TRACES_SAMPLER
          value: parentbased_always_on
        - name: OTEL_PROPAGATORS
          value: tracecontext
