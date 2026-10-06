---
# Ambient enrollment of this namespace is owned by the app-of-apps root.
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: default
  namespace: {{ .Release.Namespace }}
spec:
  mtls:
    mode: STRICT
