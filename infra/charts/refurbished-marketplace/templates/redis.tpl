{{- range $name := list "cart" "checkout" }}
{{- if (index $.Values.services $name).enabled }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $name }}-redis-config
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "3"
data:
  redis-additional.conf: |
    appendonly yes
    appendfsync everysec
    maxmemory {{ $.Values.redis.maxMemory }}
    maxmemory-policy noeviction
---
apiVersion: redis.redis.opstreelabs.in/v1beta2
kind: Redis
metadata:
  name: {{ $name }}-redis
  namespace: {{ $.Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  podSecurityContext:
    fsGroup: 1000
    fsGroupChangePolicy: OnRootMismatch
  kubernetesConfig:
    image: {{ $.Values.redis.image }}
    imagePullPolicy: IfNotPresent
    resources:
{{ toYaml $.Values.redis.resources | nindent 6 }}
  redisConfig:
    additionalRedisConfig: {{ $name }}-redis-config
  storage:
    keepAfterDelete: true
    volumeClaimTemplate:
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: {{ $.Values.redis.storageSize }}
{{- end }}
{{- end }}
