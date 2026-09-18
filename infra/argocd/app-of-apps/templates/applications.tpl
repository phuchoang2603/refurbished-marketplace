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
  project: refurbished-marketplace
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
          imagePullPolicy: {{ $.Values.global.imagePullPolicy | default "IfNotPresent" | quote }}
{{- end }}
  destination:
    name: {{ $.Values.destinationName | quote }}
    namespace: {{ $app.namespace }}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
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
  project: refurbished-marketplace
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
          imagePullPolicy: {{ .Values.global.imagePullPolicy | default "IfNotPresent" | quote }}
{{- end }}
  destination:
    name: {{ .Values.destinationName | quote }}
    namespace: ecommerce
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ServerSideApply=true
    # Argo owns the ecommerce namespace (do not template a Namespace in the chart).
{{- end }}
