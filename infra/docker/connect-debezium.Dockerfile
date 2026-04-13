FROM quay.io/strimzi/kafka:0.51.0-kafka-4.2.0

# Orders outbox uses BYTEA protobuf; KafkaConnector uses ByteArrayConverter (built into Kafka Connect).

USER root
RUN mkdir -p /opt/kafka/plugins/debezium-postgres && \
  curl -L https://repo1.maven.org/maven2/io/debezium/debezium-connector-postgres/3.5.0.Final/debezium-connector-postgres-3.5.0.Final-plugin.tar.gz | \
  tar -xzf - -C /opt/kafka/plugins/debezium-postgres/

USER 1001
