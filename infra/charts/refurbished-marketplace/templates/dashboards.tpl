{{- range $path, $_ := .Files.Glob "dashboards/*.json" }}
{{- $name := base $path | trimSuffix ".json" }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ printf "refurbished-marketplace-dashboard-%s" $name | trunc 63 | trimSuffix "-" }}
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
    app.kubernetes.io/name: refurbished-marketplace
    app.kubernetes.io/component: grafana-dashboard
    app.kubernetes.io/instance: {{ $.Release.Name }}
    app.kubernetes.io/managed-by: {{ $.Release.Service }}
  annotations:
    argocd.argoproj.io/sync-options: ServerSideApply=true
data:
  {{ base $path }}: |-
{{ $.Files.Get $path | nindent 4 }}
{{- end }}
