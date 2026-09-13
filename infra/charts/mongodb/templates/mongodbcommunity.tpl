apiVersion: mongodbcommunity.mongodb.com/v1
kind: MongoDBCommunity
metadata:
  name: {{ .Values.name }}
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  members: {{ .Values.members }}
  type: ReplicaSet
  version: {{ .Values.version | quote }}
  security:
    authentication:
      modes:
        - SCRAM
  users:
    - name: {{ .Values.user.name }}
      db: admin
      passwordSecretRef:
        name: {{ .Values.user.passwordSecretName }}
      roles:
        - name: readWrite
          db: {{ .Values.user.database | quote }}
        - name: dbAdmin
          db: {{ .Values.user.database | quote }}
      scramCredentialsSecretName: {{ .Values.user.scramCredentialsSecretName }}
  additionalMongodConfig:
    storage.wiredTiger.engineConfig.cacheSizeGB: {{ .Values.wiredTigerCacheSizeGB }}
  statefulSet:
    spec:
      template:
        spec:
          containers:
            - name: mongod
              resources:
                {{- toYaml .Values.resources | nindent 16 }}
      volumeClaimTemplates:
        - metadata:
            name: data-volume
          spec:
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: {{ .Values.persistence.size | quote }}
