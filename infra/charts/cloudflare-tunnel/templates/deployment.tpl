apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
  labels:
    app.kubernetes.io/name: cloudflare-tunnel
    app.kubernetes.io/instance: {{ .Release.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudflare-tunnel
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudflare-tunnel
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      securityContext:
        sysctls:
          - name: net.ipv4.ping_group_range
            value: "65532 65532"
      containers:
        - name: cloudflared
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          args:
            - tunnel
            - --no-autoupdate
            - --loglevel
            - info
            - --metrics
            - 0.0.0.0:2000
            - run
          env:
            - name: TUNNEL_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.tunnel.existingSecret }}
                  key: {{ .Values.tunnel.existingSecretKey }}
          ports:
            - name: metrics
              containerPort: 2000
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /ready
              port: metrics
            failureThreshold: 1
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: metrics
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 10
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 65532
            runAsGroup: 65532
            readOnlyRootFilesystem: true
{{- with .Values.resources }}
          resources:
{{ toYaml . | nindent 12 }}
{{- end }}
