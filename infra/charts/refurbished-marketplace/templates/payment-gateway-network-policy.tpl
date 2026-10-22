{{- if and .Values.services.web.enabled .Values.services.checkout.enabled .Values.services.checkout.paymentGateway }}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-payment-gateway
  namespace: {{ .Release.Namespace }}
spec:
  podSelector:
    matchLabels:
      app: web
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector: {}
      ports:
        - port: {{ .Values.services.web.port }}
          protocol: TCP
        - port: 15008
          protocol: TCP
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: payment-gateway
          podSelector:
            matchLabels:
              app.kubernetes.io/name: notifier
              payment-gateway/release: notifier
      ports:
        - port: {{ .Values.services.web.port }}
          protocol: TCP
        - port: 15008
          protocol: TCP
    - from:
        - ipBlock:
            cidr: 169.254.7.127/32
      ports:
        - port: {{ .Values.services.web.port }}
          protocol: TCP
{{- end }}
