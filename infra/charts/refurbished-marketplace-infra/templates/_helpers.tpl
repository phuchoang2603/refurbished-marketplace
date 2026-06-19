{{- define "refurbished-marketplace-infra.dopplerKeyPrefix" -}}
{{- . | upper | replace "-" "_" -}}
{{- end -}}
