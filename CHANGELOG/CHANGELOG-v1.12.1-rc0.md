# Flyte 1.12.1-rc0 Release Notes

Flyte 1.12.1-rc0 is a release candidate that focuses on documentation enhancements, bug fixes, and improvements to the core infrastructure. This release also includes contributions from a new member of the Flyte community. Below are the highlights of this release.

## 🚀 New Features & Improvements

1. **Documentation Improvements**  
   - [Removed the source code renderer section from the Decks article](https://github.com/flyteorg/flyte/pull/5397).
   - Added [documentation for OpenAI batch agent backend setup](https://github.com/flyteorg/flyte/pull/5291).
   - [Updated the example Flyte agent Dockerfile](https://github.com/flyteorg/flyte/pull/5412).
   - Fixed [documentation link to testing agent on local cluster](https://github.com/flyteorg/flyte/pull/5398).
   - [Fixed Kubeflow webhook error](https://github.com/flyteorg/flyte/pull/5410) in the documentation.
   - [Updated Flytekit version to 1.12.1b2](https://github.com/flyteorg/flyte/pull/5411) in monodocs requirements.
   - [Updated Flytefile.md](https://github.com/flyteorg/flyte/pull/5428) and replaced [SHA instead of master in RLI links](https://github.com/flyteorg/flyte/pull/5434).

2. **Infrastructure and Configuration**  
   - [Reverted "Ensure token is refreshed on Unauthenticated"](https://github.com/flyteorg/flyte/pull/5404).
   - [Updated core Helm chart for propeller configuration of agent service](https://github.com/flyteorg/flyte/pull/5402).
   - [Fixed Flytectl install script](https://github.com/flyteorg/flyte/pull/5405) in the monorepo.
   - [Moved to upstream mockery](https://github.com/flyteorg/flyte/pull/4937).
   - [Used a different git command to match the Flyteidl tags](https://github.com/flyteorg/flyte/pull/5419).

3. **Bug Fixes**  
   - [Handled auto-refresh cache race condition](https://github.com/flyteorg/flyte/pull/5406).
   - [Fixed typos using codespell CI job](https://github.com/flyteorg/flyte/pull/5418).
   - [Fixed build failure](https://github.com/flyteorg/flyte/pull/5425) in monodocs.
   - [Replaced Azure AD OIDC URL with the correct one](https://github.com/flyteorg/flyte/pull/4075).

4. **Miscellaneous**  
   - [Updated the lock file](https://github.com/flyteorg/flyte/pull/5416).
   - [Added executionClusterLabel](https://github.com/flyteorg/flyte/pull/5394) for better execution cluster management.

## 🆕 New Contributors

- **@EraYaN** for [replacing Azure AD OIDC URL with the correct one](https://github.com/flyteorg/flyte/pull/4075).

**[Full Changelog](https://github.com/flyteorg/flyte/compare/flytectl/v0.8.21...v1.12.1-rc0)**
