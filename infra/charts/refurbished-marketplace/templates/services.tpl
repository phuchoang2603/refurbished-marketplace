{{- range $name, $svc := .Values.services }}
{{- if $svc.enabled }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  namespace: {{ $.Release.Namespace }}
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
{{- if $svc.db }}
      initContainers:
        - name: wait-for-db
          image: postgres:16-alpine
          command: ["sh", "-c"]
          args:
            - >-
              until pg_isready -h {{ $svc.db.host }} -p {{ $svc.db.port }};
              do echo "waiting for database {{ $svc.db.host }}"; sleep 2; done
{{- end }}
      containers:
        - name: {{ $name }}
          image: {{ $svc.image }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            - containerPort: {{ $svc.port }}
          env:
{{- if $svc.db }}
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.db.secretName }}
                  key: {{ $svc.db.usernameKey }}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.db.secretName }}
                  key: {{ $svc.db.passwordKey }}
            - name: DB_URL
              value: {{ printf "postgres://$(DB_USER):$(DB_PASSWORD)@%s:%v/%s?sslmode=disable" $svc.db.host $svc.db.port $svc.db.name | quote }}
{{- end }}
{{- if $svc.auth }}
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: {{ $svc.auth.secretName }}
                  key: {{ $svc.auth.secretKey }}
{{- end }}
{{- if $svc.env }}
{{- range $key, $value := $svc.env }}
            - name: {{ $key }}
              value: {{ $value | quote }}
{{- end }}
{{- end }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  namespace: {{ $.Release.Namespace }}
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
