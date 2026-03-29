{{- range $name, $svc := .Values.services }}
{{- if $svc.enabled }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  namespace: {{ $svc.namespace }}
  labels:
    app: {{ $name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ $name }}
  template:
    metadata:
      labels:
        app: {{ $name }}
    spec:
      containers:
        - name: {{ $name }}
          image: {{ $svc.image }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            - containerPort: {{ $svc.port }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  namespace: {{ $svc.namespace }}
  labels:
    app: {{ $name }}
spec:
  selector:
    app: {{ $name }}
  ports:
    - name: http
      port: {{ $svc.port }}
      targetPort: {{ $svc.port }}
{{- end }}
{{- end }}
