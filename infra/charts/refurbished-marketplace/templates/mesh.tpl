{{- if or .Values.mesh.ambient.enabled .Values.mesh.waypoint.enabled }}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Release.Namespace }}
  labels:
{{- if .Values.mesh.ambient.enabled }}
    istio.io/dataplane-mode: ambient
{{- end }}
{{- if .Values.mesh.waypoint.enabled }}
    istio.io/use-waypoint: {{ default "ecommerce-waypoint" .Values.mesh.waypoint.name | quote }}
{{- end }}
{{- end }}
{{- if .Values.mesh.waypoint.enabled }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: {{ default "ecommerce-waypoint" .Values.mesh.waypoint.name }}
  namespace: {{ .Release.Namespace }}
  labels:
    istio.io/waypoint-for: service
  annotations:
    argocd.argoproj.io/sync-wave: "5"
spec:
  gatewayClassName: istio-waypoint
  listeners:
    - name: mesh
      port: 15008
      protocol: HBONE
{{- end }}
