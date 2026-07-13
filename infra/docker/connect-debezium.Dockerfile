FROM quay.io/strimzi/kafka:1.0.0-kafka-4.2.0

# Orders outbox uses BYTEA protobuf; KafkaConnector uses ByteArrayConverter (built into Kafka Connect).
# OpenTelemetry jars enable Debezium EventRouter tracing (tracingspancontext → Kafka traceparent).

USER root
# Flatten plugin jars into one Connect plugin directory (avoid nested tarball dir slow scans).
RUN mkdir -p /opt/kafka/plugins/debezium-connector-postgres /opt/kafka/libs/otel && \
  curl -L https://repo1.maven.org/maven2/io/debezium/debezium-connector-postgres/3.5.0.Final/debezium-connector-postgres-3.5.0.Final-plugin.tar.gz | \
  tar -xzf - -C /opt/kafka/plugins/debezium-connector-postgres --strip-components=1 && \
  cd /opt/kafka/libs/otel && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-api/1.49.0/opentelemetry-api-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-context/1.49.0/opentelemetry-context-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk/1.49.0/opentelemetry-sdk-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-common/1.49.0/opentelemetry-sdk-common-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-trace/1.49.0/opentelemetry-sdk-trace-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-metrics/1.49.0/opentelemetry-sdk-metrics-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-logs/1.49.0/opentelemetry-sdk-logs-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-extension-autoconfigure-spi/1.49.0/opentelemetry-sdk-extension-autoconfigure-spi-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-sdk-extension-autoconfigure/1.49.0/opentelemetry-sdk-extension-autoconfigure-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-exporter-otlp/1.49.0/opentelemetry-exporter-otlp-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-exporter-otlp-common/1.49.0/opentelemetry-exporter-otlp-common-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-exporter-sender-okhttp/1.49.0/opentelemetry-exporter-sender-okhttp-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/opentelemetry-exporter-common/1.49.0/opentelemetry-exporter-common-1.49.0.jar && \
  curl -fsSLO https://repo1.maven.org/maven2/io/opentelemetry/semconv/opentelemetry-semconv/1.32.0/opentelemetry-semconv-1.32.0.jar && \
  chown -R 1001:0 /opt/kafka/libs/otel

ENV CLASSPATH="/opt/kafka/libs/otel/*"

USER 1001
