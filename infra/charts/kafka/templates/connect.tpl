---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: {{ .Values.connect.clusterName }}
  namespace: {{ $.Release.Namespace }}
  annotations:
    strimzi.io/use-connector-resources: "true"
spec:
  version: {{ .Values.kafka.version }}
  image: {{ .Values.connect.image }}
  replicas: 1
  bootstrapServers: {{ printf "%s-kafka-bootstrap:9092" .Values.kafka.clusterName }}
  groupID: {{ .Values.connect.clusterName }}
  offsetStorageTopic: {{ printf "%s-offsets" .Values.connect.clusterName }}
  configStorageTopic: {{ printf "%s-configs" .Values.connect.clusterName }}
  statusStorageTopic: {{ printf "%s-status" .Values.connect.clusterName }}
  config:
    config.providers: secrets
    config.providers.secrets.class: io.strimzi.kafka.KubernetesSecretConfigProvider
    offset.storage.replication.factor: -1
    config.storage.replication.factor: -1
    status.storage.replication.factor: -1
