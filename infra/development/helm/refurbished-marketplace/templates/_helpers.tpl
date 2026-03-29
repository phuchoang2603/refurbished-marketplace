{{- define "refurbished-marketplace.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "refurbished-marketplace.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "refurbished-marketplace.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
