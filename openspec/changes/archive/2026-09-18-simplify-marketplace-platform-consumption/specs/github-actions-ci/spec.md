## REMOVED Requirements

### Requirement: Helm validation on chart changes

**Reason**: Helm validation is removed from CI at the user's request. No custom validation script or devenv task replaces it. Treefmt retains the existing Oxfmt integration for plain YAML formatting; Helm templates remain excluded.

**Migration**: Remove the Helm CI job and its path-filter output. Remove the local manifest validation script, devenv task, and validation-only packages. No replacement automated rendering or schema validation is required.
