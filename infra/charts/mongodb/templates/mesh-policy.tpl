apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-{{ .Values.name }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  description: Products, Kafka Connect, and kubelet to MongoDB Community 27017. No Cilium mTLS.
  endpointSelector:
    matchLabels:
      app: {{ printf "%s-svc" .Values.name }}
{{- if .Values.meshPolicy.enforce }}
  enableDefaultDeny:
    ingress: true
{{- else }}
  enableDefaultDeny:
    ingress: false
{{- end }}
  ingress:
    - fromEntities:
        - host
      toPorts:
        - ports:
            - port: "27017"
              protocol: TCP
    - fromEndpoints:
        - matchLabels:
            app: products
      toPorts:
        - ports:
            - port: "27017"
              protocol: TCP
    - fromEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: {{ default "kafka" .Values.meshPolicy.connectNamespace }}
            strimzi.io/kind: KafkaConnect
            strimzi.io/cluster: {{ default "ecommerce-connect-cluster" .Values.meshPolicy.connectCluster }}
      toPorts:
        - ports:
            - port: "27017"
              protocol: TCP
