## 1. Complete the ownership split

- [x] 1.1 Confirm the existing operator and observability chart deletions and remove remaining platform catalog entries, drift exceptions, and unused namespace metadata branches from `infra/argocd`; verify dev/prod app-of-apps renders contain exactly the five application children and no platform installation paths.
- [x] 1.2 Set the committed dev root revision to `main` and remove its historical branch instructions; verify both root renders preserve destination names, revision inheritance, dev SHA/prod main image settings, and production overlays.
- [x] 1.3 Audit remaining application charts against the sibling platform's pinned APIs and watch scopes, adapting consumers only where required; verify rendered inventories retain CNPG, MongoDBCommunity, Strimzi, ExternalSecret, VMPodScrape, Cilium, Gateway/HTTPRoute resources, workload RBAC, and schema initialization jobs, with no shared operator or CRD installation.

## 2. Retain application observability assets

- [x] 2.1 Add marketplace RED/log dashboard JSON and a dedicated dashboard ConfigMap template to the marketplace chart; verify JSON parsing, unique names, `monitoring` namespace, `grafana_dashboard: "1"` labels, server-side apply, and datasource references matching platform configuration.
- [x] 2.2 Confirm retained scrape/export settings and policies against platform VMAgent selectors, Grafana sidecar settings, log collection, and OTLP service configuration; verify dev/prod renders preserve `/metrics`, scrape access, and the trace endpoint and contain no monitoring Namespace, Grafana deployment, or datasource provisioning.

## 3. Update deployment guidance and project context

- [x] 3.1 Rewrite GitOps documentation for a fresh platform-first deployment, using current platform root names/paths and explicit prerequisite checks; verify the guide distinguishes waves from readiness guarantees and includes required image availability without ownership-transfer steps.
- [x] 3.2 Rewrite observability documentation around marketplace telemetry, dashboards, access, and troubleshooting; verify all platform administration links point to the current sibling layout and no local observability chart or custom folder guarantee remains.
- [x] 3.3 Update secrets guidance and root comments to distinguish platform ESO/store ownership from application token preparation and ExternalSecrets; verify documented remote keys match all active consumers and token examples remain placeholders.
- [x] 3.4 Update `openspec/config.yaml`, CI documentation, and current README/diagram ownership references as needed; verify no current guidance claims this repo installs shared operators or observability, excluding archived change records from the cleanup.

## 4. Validate application deployment configuration

- [x] 4.1 Remove the Helm CI job and its path-filter output; retain existing treefmt/Oxfmt source formatting.
- [x] 4.2 Check treefmt-nix integrations for Helm/Kubernetes and remove the custom validation script, devenv task, and validation-only packages when no built-in equivalent exists.
- [x] 4.3 Update documentation and specifications to remove automated manifest validation requirements; retain prior offline verification evidence without claiming live readiness.

## 5. Reconcile specifications

- [x] 5.1 Sync the four delta specs through the OpenSpec sync workflow and update the argocd-gitops/platform-observability Purpose text to match consumption ownership; verify current specs no longer require local platform installers or obsolete branch defaults and archived history remains untouched.
- [x] 5.2 Run strict OpenSpec validation and inspect the final diff for consistency across charts, CI, docs, and specs; verify there are no migration/adoption mechanisms, sibling-repo edits, or unrelated service changes.

## 6. Resolve verification findings

- [x] 6.1 Align DB username specifications with generated usernames and add actionable scrape-health checks.
- [x] 6.2 Replace the structured-logging Purpose placeholder and rerun strict validation.

Verification evidence: prior offline checks confirmed five children per environment, destination names, inherited revisions, production overlays, and dev SHA/prod main workload image tags. The validation script and task used for those checks have been removed at the user’s request. No live cluster readiness was tested.
