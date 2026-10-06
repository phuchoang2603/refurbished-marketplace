{{- if .Values.mesh.waypoint.enabled }}
{{- $resilience := .Values.mesh.resilience }}
{{- range $name := $resilience.services }}
{{- $svc := required (printf "mesh.resilience.services entry %q has no services.%s" $name $name) (index $.Values.services $name) }}
{{- if $svc.enabled }}
---
apiVersion: networking.istio.io/v1
kind: DestinationRule
metadata:
  name: {{ $name }}
  namespace: {{ $.Release.Namespace }}
spec:
  host: {{ printf "%s.%s.svc.cluster.local" $name $.Release.Namespace }}
  trafficPolicy:
{{- with $resilience.connectionPool }}
    connectionPool:
{{- toYaml . | nindent 6 }}
{{- end }}
{{- with $resilience.retryBudget }}
    retryBudget:
{{- toYaml . | nindent 6 }}
{{- end }}
{{- if gt (int (default 1 $svc.replicas)) 1 }}
{{- with $resilience.outlierDetection }}
    outlierDetection:
{{- toYaml . | nindent 6 }}
{{- end }}
{{- end }}
{{- with $resilience.grpcRetries }}
{{- with (index .methods $name) }}
---
# Attached at the waypoint; the catch-all rule keeps unlisted methods routed without retries.
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: {{ $name }}-mesh
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "6"
spec:
  parentRefs:
    - group: ""
      kind: Service
      name: {{ $name }}
      port: {{ $svc.port }}
  rules:
    - matches:
{{- range . }}
        - path:
            type: Exact
            value: /{{ . }}
{{- end }}
      timeouts:
{{- toYaml $resilience.grpcRetries.timeouts | nindent 8 }}
      retry:
{{- toYaml $resilience.grpcRetries.retry | nindent 8 }}
      backendRefs:
        - name: {{ $name }}
          port: {{ $svc.port }}
    - backendRefs:
        - name: {{ $name }}
          port: {{ $svc.port }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
