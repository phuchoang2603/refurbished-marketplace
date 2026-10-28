## MODIFIED Requirements

### Requirement: Release builds pin production chart image tags

After successful image publication, CI SHALL publish a deployment ref pinning all child Applications to the built Git revision and immutable image tag. It SHALL deploy trusted PRs to dev and main to prod, without writing an image-pin commit back to main or deploying fork PRs.

#### Scenario: Images build on a branch push

- **WHEN** all images build successfully for a commit on a branch
- **THEN** CI deploys the matching immutable Git/image revision through the root without writing an overlay-pin commit

#### Scenario: Branch moved during the build

- **WHEN** the branch head no longer matches the built SHA
- **THEN** CI deploys the matching immutable Git/image revision through the root without writing an overlay-pin commit
