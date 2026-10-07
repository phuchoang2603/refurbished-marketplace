---
# Strimzi's generated NetworkPolicies allow only listener ports, but ambient delivers mesh traffic
# to every enrolled pod on the HBONE port. ztunnel still enforces STRICT mTLS behind it.
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-hbone
  namespace: {{ .Release.Namespace }}
spec:
  podSelector:
    matchExpressions:
      - key: strimzi.io/cluster
        operator: Exists
  policyTypes:
    - Ingress
  ingress:
    - ports:
        - protocol: TCP
          port: 15008
