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

{{- /* talos-proxmox burst worker contract. Only for pods without persistent volumes. */ -}}
{{- define "refurbished-marketplace.burstScheduling" -}}
{{- if has . (list "eligible" "required") }}
tolerations:
  - key: burst.talos.dev/stateless
    operator: Equal
    value: "true"
    effect: NoSchedule
affinity:
  nodeAffinity:
{{- if eq . "required" }}
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: burst.talos.dev/compute
              operator: In
              values: [aws]
{{- else }}
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        preference:
          matchExpressions:
            - key: burst.talos.dev/compute
              operator: NotIn
              values: [aws]
{{- end }}
{{- else if ne . "none" }}
{{- fail (printf "burst mode must be none, eligible, or required; got %q" .) }}
{{- end }}
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
