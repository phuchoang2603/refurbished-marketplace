## 1. Doppler and devenv

- [x] 1.1 Add `doppler` to `devenv.nix` packages and enable `dotenv.enable` for `.env`
- [x] 1.2 Set `DOPPLER_PROJECT` / `DOPPLER_CONFIG` in `devenv.nix`; document service token setup in development docs
- [x] 1.3 Add optional `scripts.tilt` wrapper that requires `DOPPLER_TOKEN` before exec'ing tilt

## 2. ESO manifests

- [x] 2.1 Create `infra/k8s/cluster-secret-store.yaml` (Doppler provider, service token ref in `operators`)
- [x] 2.2 Generate gitignored `infra/k8s/doppler-token.secret.yaml` via devenv `files` when `DOPPLER_TOKEN` is set
- [x] 2.3 Add `templates/external-secrets.tpl` to marketplace chart; derive DB ExternalSecrets from `services.<slug>.db` and auth secrets from `services.<slug>.auth`
- [x] 2.4 Document Doppler `dev` config key naming convention in `docs/development/secrets.md`

## 3. Tilt integration

- [x] 3.1 Add `eso-operator-install` local_resource (upstream external-secrets Helm chart in `operators`)
- [x] 3.2 Apply `infra/k8s/doppler-token.secret.yaml` and `cluster-secret-store.yaml` via `k8s_yaml`
- [x] 3.3 Deploy marketplace Helm chart (with ExternalSecrets) before Kafka so Debezium secrets exist

## 4. Remove legacy secrets

- [x] 4.1 Delete `infra/k8s/secrets.yaml`
- [x] 4.2 Verify `tilt up` with Doppler dev config: CNPG, services, Debezium connectors become healthy

## 5. Documentation and tracking

- [x] 5.1 Document provider swap procedure in development docs
- [x] 5.2 Update GitHub #10 acceptance criteria and link OpenSpec change `add-eso-doppler-secrets`
