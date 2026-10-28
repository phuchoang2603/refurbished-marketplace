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
{{- $repository := printf "%s/%s" $root.Values.global.imageRegistry $image -}}
{{- $tag := required (printf "image tag missing for %s" $repository) (index $root.Values.global.imageTags $repository) -}}
{{- printf "%s:%s" $repository $tag -}}
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
