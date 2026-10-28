{{- define "kafka.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kafka.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "kafka.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "kafka.image" -}}
{{- $root := index . 0 -}}
{{- $image := index . 1 -}}
{{- $repository := printf "%s/%s" $root.Values.global.imageRegistry $image -}}
{{- $tag := required (printf "image tag missing for %s" $repository) (index $root.Values.global.imageTags $repository) -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}
