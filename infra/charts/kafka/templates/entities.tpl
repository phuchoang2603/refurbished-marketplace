{{- $appNamespace := .Values.appNamespace | default .Release.Namespace }}
{{- range $entityName, $entity := .Values.entities }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ printf "%s-debezium-connector-role" $entityName }}
  namespace: {{ $appNamespace }}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: [{{ $entity.connector.sourceSecretName | quote }}]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ printf "%s-debezium-connector-rolebinding" $entityName }}
  namespace: {{ $appNamespace }}
subjects:
- kind: ServiceAccount
  name: {{ printf "%s-connect" $.Values.connect.clusterName }}
  namespace: {{ $.Release.Namespace }}
roleRef:
  kind: Role
  name: {{ printf "%s-debezium-connector-role" $entityName }}
  apiGroup: rbac.authorization.k8s.io
{{- range $entity.topics }}
---
apiVersion: kafka.strimzi.io/v1
kind: KafkaTopic
metadata:
  name: {{ .name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ $.Values.kafka.clusterName }}
spec:
  partitions: {{ .partitions }}
  replicas: {{ $.Values.kafka.replicas }}
  config:
    retention.ms: {{ .retention }}
{{- end }}
---
apiVersion: kafka.strimzi.io/v1
kind: KafkaConnector
metadata:
  name: {{ $entity.connector.name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ $.Values.connect.clusterName }}
spec:
  class: io.debezium.connector.postgresql.PostgresConnector
  tasksMax: 1
  # Self-heal transient startup failures (e.g. DB not ready yet). Restarts the
  # connector/task with exponential back-off, indefinitely by default.
  autoRestart:
    enabled: true
  config:
    # --- Database Connection ---
    database.hostname: {{ printf "%s.%s.svc" $entity.connector.databaseHost $appNamespace }}
    database.port: {{ $entity.connector.databasePort }}
    database.user: ${secrets:{{ $appNamespace }}/{{ $entity.connector.sourceSecretName }}:username}
    database.password: ${secrets:{{ $appNamespace }}/{{ $entity.connector.sourceSecretName }}:password}
    database.dbname: {{ $entity.connector.databaseName }}
    
    # --- Debezium Engine Settings ---
    topic.prefix: {{ $entityName | quote }}
    plugin.name: pgoutput
    slot.name: {{ printf "debezium_%s_slot" $entityName }}
    table.include.list: {{ $entity.connector.outboxTable }}
    tombstones.on.delete: "false"
{{- if $entity.connector.publicationAutocreateMode }}
    publication.autocreate.mode: {{ $entity.connector.publicationAutocreateMode | quote }}
{{- end }}
    
    # This transforms the DB row into a clean Kafka message
    transforms: outbox
    transforms.outbox.type: io.debezium.transforms.outbox.EventRouter
    transforms.outbox.table.field.event.id: "id"
    transforms.outbox.table.field.event.key: "aggregate_id"
    transforms.outbox.table.field.event.payload: "payload"
    transforms.outbox.route.by.field: "event_type"
    transforms.outbox.route.topic.replacement: "${routedByValue}"
    transforms.outbox.tracing.span.context.field: tracingspancontext
    transforms.outbox.tracing.operation.name: debezium-read
    transforms.outbox.tracing.with.context.field.only: "true"
    key.converter: org.apache.kafka.connect.storage.StringConverter
    transforms.outbox.table.expand.json.payload: "false"
    value.converter: org.apache.kafka.connect.converters.ByteArrayConverter
{{- end }}
