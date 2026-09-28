## 1. Single SecretStore owner

- [x] 1.1 Add `infra/charts/secret-store` (Chart.yaml, values.yaml, SecretStore template) and verify `helm template` renders exactly one `SecretStore/doppler`
- [x] 1.2 Add the `secret-store` child (wave `1`, namespace `ecommerce`) to `infra/argocd/app-of-apps/values.yaml` and verify both roots render five children
- [x] 1.3 Delete `secret-store.tpl` from the mongodb, meilisearch, and refurbished-marketplace charts and verify `rg -l 'kind: SecretStore' infra/charts` lists only the new chart

## 2. Burst scheduling knobs

- [x] 2.1 Add a burst scheduling helper and `burst.mode` default to the refurbished-marketplace chart, wire it into service Deployments, and verify default render has no burst toleration
- [x] 2.2 Add the same helper and `connect.burst` to the kafka chart, wire it into `KafkaConnect.spec.template.pod`, and verify default render has no burst toleration

## 3. Sizing and prod placement

- [x] 3.1 Raise Kafka broker and Connect CPU requests to 250m and verify the rendered Kafka and KafkaConnect resources
- [x] 3.2 Set `burst.mode: eligible` in marketplace `values-prod.yaml` and `connect.burst: required` in kafka `values-prod.yaml`; verify prod renders the toleration on all service Deployments (not migration Jobs) and required AWS affinity on Connect

## 4. Docs and verification

- [x] 4.1 Update `docs/deployment/gitops.md` and `docs/development/secrets.md` for the `secret-store` Application and prod burst placement; verify `rg -n 'secret-store' docs/` shows both
- [x] 4.2 Run `openspec validate stabilize-prod-scheduling --strict` and confirm it passes
- [x] 4.3 After merge, confirm `prod-kafka` Connect runs on an `burst.talos.dev/compute=aws` node and no `SharedResourceWarning` remains on either environment
