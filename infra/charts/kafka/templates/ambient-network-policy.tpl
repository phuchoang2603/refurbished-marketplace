---
# Strimzi's generated NetworkPolicies allow only listener ports. Ambient delivers mesh traffic on
# the HBONE port, where ztunnel still enforces STRICT mTLS, and rewrites kubelet probes to come
# from a fixed link-local address.
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ambient
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
    - from:
        - ipBlock:
            cidr: 169.254.7.127/32
