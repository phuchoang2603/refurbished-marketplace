FROM quay.io/strimzi/kafka:1.2.0-kafka-4.3.1

USER root
RUN mkdir -p /opt/kafka/plugins/debezium-connector-postgres && \
  curl -L https://repo1.maven.org/maven2/io/debezium/debezium-connector-postgres/3.7.0.Final/debezium-connector-postgres-3.7.0.Final-plugin.tar.gz | \
    tar -xzf - -C /opt/kafka/plugins/debezium-connector-postgres --strip-components=1 && \
  chown -R 1001:0 /opt/kafka/plugins/debezium-connector-postgres

USER 1001
