## 1. Single SecretStore owner

- [x] 1.1 Add `infra/charts/secret-store` (Chart.yaml, values.yaml, SecretStore template) and verify `helm template` renders exactly one `SecretStore/doppler`
- [x] 1.2 Add the `secret-store` child (wave `1`, namespace `ecommerce`) to `infra/argocd/app-of-apps/values.yaml` and verify both roots render five children
- [x] 1.3 Delete `secret-store.tpl` from the mongodb, meilisearch, and refurbished-marketplace charts and verify `rg -l 'kind: SecretStore' infra/charts` lists only the new chart

## 2. Sizing

- [x] 2.1 Raise Kafka broker and Connect CPU requests to 250m and verify the rendered Kafka and KafkaConnect resources

## 3. Docs and verification

- [x] 3.1 Update `docs/deployment/gitops.md` and `docs/development/secrets.md` for the `secret-store` Application; verify `rg -n 'secret-store' docs/` shows both
- [x] 3.2 Run `openspec validate stabilize-prod-scheduling --strict` and confirm it passes
- [x] 3.3 After merge, confirm no `SharedResourceWarning` remains on either environment
