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
  image: {{ include "kafka.image" (list . .Values.connect.image) }}
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
  tracing:
    type: opentelemetry
  config:
    config.providers: secrets
    config.providers.secrets.class: io.strimzi.kafka.KubernetesSecretConfigProvider
    offset.storage.replication.factor: -1
    config.storage.replication.factor: -1
    status.storage.replication.factor: -1
    connector.client.config.override.policy: All
    metadata.recovery.strategy: none
  template:
    pod:
      metadata:
        annotations:
          resource.opentelemetry.io/service.name: connect-debezium
    connectContainer:
      env:
        - name: OTEL_SERVICE_NAME
          value: connect-debezium
        - name: OTEL_TRACES_EXPORTER
          value: otlp
        - name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
          value: {{ .Values.connect.otelEndpoint | default "http://vtagent-vmks.observability.svc.cluster.local:10429/insert/opentelemetry/v1/traces" | quote }}
        - name: OTEL_EXPORTER_OTLP_PROTOCOL
          value: http/protobuf
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: {{ printf "k8s.cluster.name=%s,deployment.environment=%s" .Values.global.clusterName .Values.global.environment | quote }}
        - name: OTEL_TRACES_SAMPLER
          value: parentbased_always_on
        - name: OTEL_PROPAGATORS
          value: tracecontext
