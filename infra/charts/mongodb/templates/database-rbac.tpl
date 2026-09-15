# MCK helm creates these SAs only in the operator namespace when watchNamespace is "*".
# Community mongod pods run in ecommerce and default to mongodb-kubernetes-appdb.
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mongodb-kubernetes-appdb
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mongodb-kubernetes-database-pods
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: mongodb-kubernetes-appdb
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
rules:
  - apiGroups:
      - ""
    resources:
      - secrets
    verbs:
      - get
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - patch
      - delete
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: mongodb-kubernetes-appdb
  namespace: {{ .Release.Namespace }}
  annotations:
    argocd.argoproj.io/sync-wave: "1"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: mongodb-kubernetes-appdb
subjects:
  - kind: ServiceAccount
    name: mongodb-kubernetes-appdb
    namespace: {{ .Release.Namespace }}
