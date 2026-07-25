{{- /*
Repo-owned Grafana dashboards → ConfigMaps for the Grafana sidecar
(label grafana_dashboard=1). Enabled on local and staging so marketplace
Istio RED is available without waiting on the default-dashboard sync job.
*/}}
{{- if .Values.customDashboards.enabled }}
{{- range $path, $_ := .Files.Glob "dashboards/*.json" }}
{{- $name := base $path | trimSuffix ".json" }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ printf "dashboard-%s" $name | trunc 63 | trimSuffix "-" }}
  namespace: {{ $.Release.Namespace }}
  labels:
    grafana_dashboard: "1"
    app.kubernetes.io/component: grafana-dashboard
    app.kubernetes.io/instance: {{ $.Release.Name }}
    app.kubernetes.io/managed-by: {{ $.Release.Service }}
    app.kubernetes.io/name: {{ include "observability.name" $ }}
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=true
    grafana_folder: {{ $.Values.customDashboards.folder | default "Marketplace" | quote }}
data:
  {{ base $path }}: |-
{{ $.Files.Get $path | nindent 4 }}
{{- end }}
{{- end }}
