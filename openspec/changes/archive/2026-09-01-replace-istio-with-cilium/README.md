# replace-istio-with-cilium

Replace Istio with Cilium on Talos. Argo CD on gpu destines registered `dev` / `prod` clusters. Chart defaults are shop-dev / pay-dev; prod hosts live in `values-prod.yaml`. Observe with app OTEL → VictoriaTraces. Hubble is not required. App-level RED metrics: [#43](https://github.com/phuchoang2603/refurbished-marketplace/issues/43).
