{{- range $name, $app := .Values.apps }}
{{- /* sprig `default` treats false as empty — use hasKey for explicit disables */ -}}
{{- if or (not (hasKey $app "enabled")) $app.enabled }}
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ printf "%s-%s" $.Values.namePrefix $name }}
  finalizers:
    - resources-finalizer.argocd.argoproj.io/foreground
  annotations:
    argocd.argoproj.io/sync-wave: {{ $app.syncWave | quote }}
spec:
  project: default
  source:
    repoURL: {{ $.Values.repoURL | quote }}
    targetRevision: {{ $.Values.targetRevision | quote }}
    path: {{ $app.path }}
    helm:
      releaseName: {{ $app.releaseName }}
{{- with $app.valueFiles }}
      valueFiles:
{{- range . }}
        - {{ . | quote }}
{{- end }}
{{- end }}
{{- if and $app.injectGlobalImages $.Values.global.imageRegistry }}
      values: |
        global:
          imageRegistry: {{ $.Values.global.imageRegistry }}
          imageTag: {{ $.Values.global.imageTag | quote }}
{{- end }}
  destination:
    server: https://kubernetes.default.svc
    namespace: {{ $app.namespace }}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
{{- if $app.ignoreDifferences }}
      - RespectIgnoreDifferences=true
{{- end }}
{{- if $app.ignoreDifferences }}
  ignoreDifferences:
    - group: ""
      kind: Secret
      name: observability-grafana
      namespace: monitoring
      jsonPointers:
        - /data/admin-password
    - group: ""
      kind: Secret
      name: observability-victoria-metrics-operator-validation
      namespace: monitoring
      jsonPointers:
        - /data
    - group: admissionregistration.k8s.io
      kind: ValidatingWebhookConfiguration
      name: observability-victoria-metrics-operator-admission
      jqPathExpressions:
        - ".webhooks[]?.clientConfig.caBundle"
    - group: apps
      kind: Deployment
      name: observability-grafana
      namespace: monitoring
      jsonPointers:
        - /spec/template/metadata/annotations/checksum~1secret
{{- end }}
{{- end }}
{{- end }}
{{- if .Values.marketplace.enabled }}
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{ printf "%s-refurbished-marketplace" .Values.namePrefix }}
  finalizers:
    - resources-finalizer.argocd.argoproj.io/foreground
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  project: default
  source:
    repoURL: {{ .Values.repoURL | quote }}
    targetRevision: {{ .Values.targetRevision | quote }}
    path: infra/charts/refurbished-marketplace
    helm:
      releaseName: refurbished-marketplace
{{- with .Values.marketplace.valueFiles }}
      valueFiles:
{{- range . }}
        - {{ . | quote }}
{{- end }}
{{- end }}
{{- if .Values.global.imageRegistry }}
      values: |
        global:
          imageRegistry: {{ .Values.global.imageRegistry }}
          imageTag: {{ .Values.global.imageTag | quote }}
{{- end }}
  destination:
    server: https://kubernetes.default.svc
    namespace: ecommerce
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
    # Own the destination namespace's metadata here instead of templating a
    # Namespace object in the chart (which caused prune+recreate churn and
    # cascade-deleted the databases). Enrolls ecommerce into ambient mesh.
    managedNamespaceMetadata:
      labels:
        istio.io/dataplane-mode: ambient
        istio.io/use-waypoint: ecommerce-waypoint
{{- end }}
