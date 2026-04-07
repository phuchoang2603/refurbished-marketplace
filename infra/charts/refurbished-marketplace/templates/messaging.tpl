---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: {{ printf "%s-dual-role" .Values.messaging.kafkaClusterName }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ .Values.messaging.kafkaClusterName }}
spec:
  replicas: {{ .Values.messaging.kafka.replicas }}
  roles:
    - controller
    - broker
  storage:
      type: jbod
      volumes:
        - id: 0
          type: persistent-claim
          size: {{ .Values.messaging.kafka.storage.size }}
          kraftMetadata: shared

---
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: {{ .Values.messaging.kafkaClusterName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    strimzi.io/node-pools: enabled
    strimzi.io/kraft: enabled
spec:
  kafka:
    version: {{ .Values.messaging.kafka.version }}
    metadataVersion: {{ .Values.messaging.kafka.metadataVersion }}
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    config:
      auto.create.topics.enable: "false"
      offsets.topic.replication.factor: {{ .Values.messaging.kafka.replicationFactor }}
      transaction.state.log.replication.factor: {{ .Values.messaging.kafka.replicationFactor }}
      transaction.state.log.min.isr: {{ .Values.messaging.kafka.minInsyncReplicas }}
      default.replication.factor: {{ .Values.messaging.kafka.replicationFactor }}
      min.insync.replicas: {{ .Values.messaging.kafka.minInsyncReplicas }}
  entityOperator:
    topicOperator: {}
    userOperator: {}

---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: {{ .Values.messaging.connectClusterName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    strimzi.io/use-connector-resources: "true"
spec:
  version: {{ .Values.messaging.kafka.version }}
  replicas: 1
  bootstrapServers: {{ printf "%s-kafka-bootstrap:9092" .Values.messaging.kafkaClusterName }}
  groupID: {{ .Values.messaging.connectClusterName }}
  offsetStorageTopic: {{ printf "%s-offsets" .Values.messaging.connectClusterName }}
  configStorageTopic: {{ printf "%s-configs" .Values.messaging.connectClusterName }}
  statusStorageTopic: {{ printf "%s-status" .Values.messaging.connectClusterName }}
  config:
    config.providers: secrets
    config.providers.secrets.class: io.strimzi.kafka.KubernetesSecretConfigProvider
    offset.storage.replication.factor: -1
    config.storage.replication.factor: -1
    status.storage.replication.factor: -1
  build:
    output:
      type: docker
      image: ttl.sh/ecommerce-debezium-connect:24h
    plugins:
      - name: debezium-postgres-connector
        artifacts:
          - type: maven
            group: io.debezium
            artifact: debezium-connector-postgres
            version: {{ .Values.messaging.debeziumConnectorVersion }}

{{- range $entityName, $entity := .Values.messaging.entities }}
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
  name: {{ printf "%s-connect" $.Values.messaging.connectClusterName }}
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
    strimzi.io/cluster: {{ $.Values.messaging.kafkaClusterName }}
spec:
  partitions: {{ $entity.topic.partitions }}
  replicas: {{ $entity.topic.replicas }}
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: {{ $entity.connector.name }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ $.Values.messaging.connectClusterName }}
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
