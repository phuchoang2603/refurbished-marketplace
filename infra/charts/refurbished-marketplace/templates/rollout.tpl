{{- if .Values.rollout.enabled }}
{{- if not .Values.ingress.enabled }}{{ fail "Web canary requires ingress.enabled" }}{{- end }}
apiVersion: v1
kind: Service
metadata:
  name: web-canary
  namespace: {{ .Release.Namespace }}
spec:
  selector:
    app: web
  ports:
    - name: http
      port: {{ .Values.services.web.port }}
      targetPort: http
---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: web-catalog
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "4"
spec:
  args:
    - name: revision
  metrics:
{{- $selector := printf "\"service.name\"=\"web\",\"deployment.environment\"=%s,\"k8s.namespace.name\"=%s,\"rollout.revision\"=\"{{args.revision}}\",\"http.route\"=~\"/|/products.*\"" (.Values.global.environment | quote) (.Release.Namespace | quote) }}
{{- $count := printf "{__name__=\"http.server.request.duration_count\",%s}" $selector }}
{{- $buckets := printf "{__name__=\"http.server.request.duration_bucket\",%s}" $selector }}
{{- $errors := printf "{__name__=\"http.server.request.duration_count\",%s,\"http.response.status_code\"=~\"5..\"}" $selector }}
{{- $window := .Values.rollout.analysis.window }}
{{- $metrics := list
  (dict "name" "requests" "condition" (printf "result[0] >= %v" .Values.rollout.analysis.minimumRequests) "query" (printf "sum(increase(%s[%s]))" $count $window))
  (dict "name" "freshness" "condition" (printf "result[0] <= %v" .Values.rollout.analysis.maximumSampleAgeSeconds) "query" (printf "time() - max(timestamp(%s))" $count))
  (dict "name" "error-rate" "condition" (printf "result[0] <= %v" .Values.rollout.analysis.maximumErrorRate) "query" (printf "(sum(increase(%s[%s])) or vector(0)) / sum(increase(%s[%s]))" $errors $window $count $window))
  (dict "name" "p95" "condition" (printf "result[0] <= %v" .Values.rollout.analysis.maximumP95Seconds) "query" (printf "histogram_quantile(0.95, sum by (le) (rate(%s[%s])))" $buckets $window)) }}
{{- range $metrics }}
    - name: {{ .name }}
      initialDelay: {{ $.Values.rollout.analysis.warmup }}
      interval: {{ $.Values.rollout.analysis.interval }}
      count: {{ $.Values.rollout.analysis.count }}
      failureLimit: 0
      consecutiveErrorLimit: 2
      successCondition: {{ printf "len(result) == 1 && !isNaN(result[0]) && !isInf(result[0]) && result[0] >= 0 && %s" .condition | quote }}
      provider:
        prometheus:
          address: {{ $.Values.rollout.metricsAddress }}
          timeout: 10
          query: {{ .query | quote }}
{{- end }}
{{- end }}
