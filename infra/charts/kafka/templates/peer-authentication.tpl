---
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: strimzi-operator-access
  namespace: {{ .Release.Namespace }}
spec:
  selector:
    matchLabels:
      strimzi.io/kind: Kafka
  mtls:
    mode: STRICT
  portLevelMtls:
    9090:
      mode: PERMISSIVE
    9091:
      mode: PERMISSIVE
