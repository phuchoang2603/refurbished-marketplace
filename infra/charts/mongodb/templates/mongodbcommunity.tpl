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
    # Debezium MongoDbConnector config validation calls listDatabaseNames() and
    # treats an empty result as unauthorized, even with capture.scope=database.
    roles:
      - role: catalogListDatabases
        db: admin
        privileges:
          - resource:
              cluster: true
            actions:
              - listDatabases
  users:
    - name: {{ .Values.user.name }}
      db: {{ .Values.user.database | quote }}
      passwordSecretRef:
        name: {{ .Values.user.passwordSecretName }}
      roles:
        - name: readWrite
          db: {{ .Values.user.database | quote }}
        - name: catalogListDatabases
          db: admin
      scramCredentialsSecretName: {{ .Values.user.scramCredentialsSecretName }}
  additionalMongodConfig:
    storage.wiredTiger.engineConfig.cacheSizeGB: {{ .Values.wiredTigerCacheSizeGB }}
  statefulSet:
    spec:
      template:
        spec:
          {{- with .Values.nodeAffinity }}
          affinity:
            nodeAffinity:
              {{- toYaml . | nindent 14 }}
          {{- end }}
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
