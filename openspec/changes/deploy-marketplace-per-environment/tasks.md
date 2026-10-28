## 1. Delivery configuration

- [x] 1.1 Add a dev root and coordinate immutable Git/image overrides across root and children; keep lean shared topology and environment hostnames.
- [x] 1.2 Validate charts in CI and deploy trusted PRs to dev and main to prod after the complete image matrix succeeds; remove overlay-pin commits and unsafe PR-close image deletion.
- [x] 1.3 Bootstrap environment-scoped marketplace credentials and GitHub Environment publication rules without exposing values.
- [x] 1.4 Update affected deployment/secrets documentation and stale OpenSpec context for the new environment contract.

## 2. Validation and rollout

- [x] 2.1 Validate both environment renders, workflow guards and OpenSpec artifacts.
- [ ] 2.2 Open a marketplace PR and verify CI image publication and dev deployment.
- [ ] 2.3 Verify dev application/secret/storage readiness, migrations, event delivery, search, public storefront and unavailable checkout without the gateway.
