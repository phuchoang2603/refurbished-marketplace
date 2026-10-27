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

{{- define "refurbished-marketplace.dopplerKeyPrefix" -}}
{{- . | upper | replace "-" "_" -}}
{{- end -}}

{{- define "refurbished-marketplace.image" -}}
{{- $root := index . 0 -}}
{{- $image := index . 1 -}}
{{- $tagOverride := "" -}}
{{- if gt (len .) 2 -}}
{{- $tagOverride = index . 2 | default "" -}}
{{- end -}}
{{- if $root.Values.global.imageRegistry -}}
{{- $tag := $tagOverride | default $root.Values.global.imageTag | default "" -}}
{{- if $tag -}}
{{- printf "%s/%s:%s" $root.Values.global.imageRegistry $image $tag -}}
{{- else -}}
{{- printf "%s/%s" $root.Values.global.imageRegistry $image -}}
{{- end -}}
{{- else -}}
{{- $image -}}
{{- end -}}
{{- end -}}

{{- define "refurbished-marketplace.postgresEnv" -}}
- name: PGUSER
  value: {{ .user | quote }}
- name: PGPASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .secret }}
      key: {{ .service.db.passwordKey }}
- name: PGHOST
  value: {{ .service.db.host | quote }}
- name: PGPORT
  value: {{ .service.db.port | quote }}
- name: PGDATABASE
  value: {{ .service.db.name | quote }}
- name: PGSSLMODE
  value: "disable"
{{- end -}}
