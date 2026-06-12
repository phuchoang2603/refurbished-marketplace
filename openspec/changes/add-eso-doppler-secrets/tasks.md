## 1. Doppler and devenv

- [x] 1.1 Add `doppler` to `devenv.nix` packages and enable `dotenv.enable` for `.env`
- [x] 1.2 Document `doppler login`, `doppler setup`, service token creation, and `.env` `DOPPLER_TOKEN` in CONTRIBUTING
- [x] 1.3 Add optional `scripts.tilt` wrapper that requires `DOPPLER_TOKEN` before exec'ing tilt

## 2. ESO manifests

- [x] 2.1 Create `infra/eso/cluster-secret-store.yaml` (Doppler provider, service token ref)
- [x] 2.2 Create `ExternalSecret` manifests for `users-app`, `products-app`, `orders-app`, `payment-app`, `users-auth`
- [x] 2.3 Document Doppler `dev` config key names matching K8s secret shape (seed from former `secrets.yaml`)

## 3. Tilt integration

- [x] 3.1 Add `eso-operator-install` local_resource (upstream external-secrets Helm chart)
- [x] 3.2 Add `doppler-token` local_resource (`kubectl create secret … dopplerToken="$DOPPLER_TOKEN"`)
- [x] 3.3 Replace `k8s_yaml('./infra/k8s/secrets.yaml')` with `k8s_yaml('./infra/eso/')` and wire resource_deps before app charts

## 4. Remove legacy secrets

- [x] 4.1 Delete `infra/k8s/secrets.yaml`
- [x] 4.2 Verify `tilt up` with Doppler dev config: CNPG, services, Debezium connectors become healthy

## 5. Documentation and tracking

- [x] 5.1 Document provider swap procedure (ClusterSecretStore only) in CONTRIBUTING or `docs/`
- [x] 5.2 Update GitHub #10 acceptance criteria and link OpenSpec change `add-eso-doppler-secrets`
