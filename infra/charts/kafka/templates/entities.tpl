{{- range $entityName, $entity := .Values.entities }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ printf "%s-debezium-connector-role" $entityName }}
  namespace: {{ $.Release.Namespace }}
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
  namespace: {{ $.Release.Namespace }}
subjects:
- kind: ServiceAccount
  name: {{ printf "%s-connect" $.Values.connect.clusterName }}
  namespace: {{ $.Release.Namespace }}
roleRef:
  kind: Role
  name: {{ printf "%s-debezium-connector-role" $entityName }}
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaTopic
metadata:
  name: {{ $entity.topic.name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ $.Values.kafka.clusterName }}
spec:
  partitions: {{ $entity.topic.partitions }}
  replicas: {{ $.Values.kafka.replicas }}
  config:
    retention.ms: {{ $entity.topic.retention }}
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: {{ $entity.connector.name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ $.Values.connect.clusterName }}
spec:
  class: io.debezium.connector.postgresql.PostgresConnector
  tasksMax: 1
  config:
    # --- Database Connection ---
    database.hostname: {{ $entity.connector.databaseHost }}
    database.port: {{ $entity.connector.databasePort }}
    database.user: ${secrets:{{ $.Release.Namespace }}/{{ $entity.connector.sourceSecretName }}:username}
    database.password: ${secrets:{{ $.Release.Namespace }}/{{ $entity.connector.sourceSecretName }}:password}
    database.dbname: {{ $entity.connector.databaseName }}
    
    # --- Debezium Engine Settings ---
    topic.prefix: {{ $entityName | quote }}
    plugin.name: pgoutput
    slot.name: {{ printf "debezium_%s_slot" $entityName }}
    publication.autocreate.mode: "filtered"
    table.include.list: {{ $entity.connector.outboxTable }}
    tombstones.on.delete: "false"
    
    # This transforms the DB row into a clean Kafka message
    transforms: outbox
    transforms.outbox.type: io.debezium.transforms.outbox.EventRouter
    transforms.outbox.table.field.event.id: "id"
    transforms.outbox.table.field.event.key: "aggregate_id"
    transforms.outbox.table.field.event.payload: "payload"
    transforms.outbox.route.by.field: "event_type"
    transforms.outbox.route.topic.replacement: "${routedByValue}"
    transforms.outbox.table.expand.json.payload: "true"
{{- end }}
