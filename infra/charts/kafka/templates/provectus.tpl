{{- if .Values.provectus.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Values.provectus.name }}
  namespace: {{ .Release.Namespace }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Values.provectus.name }}
  template:
    metadata:
      labels:
        app: {{ .Values.provectus.name }}
    spec:
      containers:
        - name: {{ .Values.provectus.name }}
          image: {{ .Values.provectus.image }}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: {{ .Values.provectus.port }}
              protocol: TCP
          env:
            - name: KAFKA_CLUSTERS_0_NAME
              value: {{ .Values.kafka.clusterName }}
            - name: KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS
              value: {{ printf "%s-kafka-bootstrap:9092" .Values.kafka.clusterName | quote }}
            - name: AUTH_TYPE
              value: "DISABLED"
            - name: MANAGEMENT_HEALTH_LDAP_ENABLED
              value: "FALSE"
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.provectus.name }}-svc
  namespace: {{ .Release.Namespace }}
spec:
  selector:
    app: {{ .Values.provectus.name }}
  type: ClusterIP
  ports:
    - name: http
      port: {{ .Values.provectus.port }}
      targetPort: {{ .Values.provectus.port }}
{{- end }}
