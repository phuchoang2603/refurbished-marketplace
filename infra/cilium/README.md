# Cilium is not owned by this repo

Do not copy Helm values here. The Talos cluster already installs Cilium from **talos-proxmox**:

- Chart values: `apps/values/cilium.yaml` (Cilium 1.18.13 via `apps/bootstrap.sh`)
- L2 pool + platform Gateways: `apps/manifests/env/{dev,prod}/network.yaml`
- Hubble / Longhorn / Argo HTTPRoutes: `apps/manifests/routes.yaml`

Live `helm get values cilium` on talos-dev matches that file, including WireGuard, Envoy L7, and `cluster.name/id`. A second values file in this repo would drift and could wipe those flags on upgrade.

This marketplace repo only consumes Cilium: `gatewayClassName: cilium` on the shop/pay Gateway, Cloudflare origin DNS, and (later) `CiliumNetworkPolicy` if needed. Hubble UI on the LAN is the cluster Gateway (`http://10.69.100.1` on dev), not something this chart should recreate.
