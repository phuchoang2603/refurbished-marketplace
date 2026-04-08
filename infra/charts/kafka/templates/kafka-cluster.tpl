---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: {{ printf "%s-dual-role" .Values.kafka.clusterName }}
  namespace: {{ $.Release.Namespace }}
  labels:
    strimzi.io/cluster: {{ .Values.kafka.clusterName }}
spec:
  replicas: {{ .Values.kafka.replicas }}
  roles:
    - controller
    - broker
  storage:
      type: jbod
      volumes:
        - id: 0
          type: persistent-claim
          size: {{ .Values.kafka.storage.size }}
          kraftMetadata: shared

---
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: {{ .Values.kafka.clusterName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    strimzi.io/node-pools: enabled
    strimzi.io/kraft: enabled
spec:
  kafka:
    version: {{ .Values.kafka.version }}
    metadataVersion: {{ .Values.kafka.metadataVersion }}
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
      offsets.topic.replication.factor: {{ .Values.kafka.replicationFactor }}
      transaction.state.log.replication.factor: {{ .Values.kafka.replicationFactor }}
      transaction.state.log.min.isr: {{ .Values.kafka.minInsyncReplicas }}
      default.replication.factor: {{ .Values.kafka.replicationFactor }}
      min.insync.replicas: {{ .Values.kafka.minInsyncReplicas }}
  entityOperator:
    topicOperator: {}
    userOperator: {}
