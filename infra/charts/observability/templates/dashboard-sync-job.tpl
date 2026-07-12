{{- $vmks := index .Values "victoria-metrics-k8s-stack" -}}
{{- $syncJob := $vmks.syncJob | default dict -}}
{{- $dashboards := $vmks.defaultDashboards | default dict -}}
{{- if and ($dashboards.enabled | default true) ($syncJob.enabled | default true) (eq $syncJob.createJob false) (.Values.dashboardSyncJob.enabled | default true) }}
{{- $image := $syncJob.image | default dict -}}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: vmks-sync-job
  namespace: {{ .Release.Namespace }}
  labels:
    app.kubernetes.io/component: syncJob
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    app.kubernetes.io/name: {{ include "observability.name" . }}
  annotations:
    argocd.argoproj.io/hook: PostSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  ttlSecondsAfterFinished: {{ $syncJob.ttlSecondsAfterFinished | default 600 }}
  backoffLimit: {{ $syncJob.backoffLimit | default 3 }}
  template:
    metadata:
      labels:
        app.kubernetes.io/component: syncJob
        app.kubernetes.io/instance: {{ .Release.Name }}
        app.kubernetes.io/managed-by: {{ .Release.Service }}
        app.kubernetes.io/name: {{ include "observability.name" . }}
      {{- with $syncJob.podAnnotations }}
      annotations: {{ toYaml . | nindent 8 }}
      {{- end }}
    spec:
      serviceAccountName: vmks-sync-job
      restartPolicy: OnFailure
      {{- with $syncJob.podSecurityContext }}
      securityContext: {{ toYaml . | nindent 8 }}
      {{- end }}
      containers:
        - name: sync-job
          image: "{{ $image.repository | default "ghcr.io/victoriametrics/sync-job" }}:{{ $image.tag | default "v0.0.9" }}"
          imagePullPolicy: {{ $image.pullPolicy | default "IfNotPresent" }}
          env:
            {{- with $syncJob.env }}
            {{- toYaml . | nindent 12 }}
            {{- end }}
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: RELEASE
              value: {{ .Release.Name | quote }}
            - name: RESOURCE_PREFIX
              value: {{ $syncJob.resourcePrefix | default "vmks" | quote }}
            - name: PRUNE
              value: {{ $syncJob.prune | default true | quote }}
            - name: OWNER_REFERENCES
              value: {{ $syncJob.useOwnerReferences | default true | quote }}
            - name: SERVICE_ACCOUNT
              valueFrom:
                fieldRef:
                  fieldPath: spec.serviceAccountName
          volumeMounts:
            - name: config
              mountPath: /etc/config
              readOnly: true
          {{- with $syncJob.containerSecurityContext }}
          securityContext: {{ toYaml . | nindent 12 }}
          {{- end }}
          {{- with $syncJob.resources }}
          resources: {{ toYaml . | nindent 12 }}
          {{- end }}
      volumes:
        - name: config
          configMap:
            name: vmks-sync-job-config
      {{- with $syncJob.nodeSelector }}
      nodeSelector: {{ toYaml . | nindent 8 }}
      {{- end }}
      {{- with $syncJob.tolerations }}
      tolerations: {{ toYaml . | nindent 8 }}
      {{- end }}
      {{- with $syncJob.affinity }}
      affinity: {{ toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
