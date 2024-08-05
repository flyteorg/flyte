# Release Process

Flytectl releases map to git tags with the prefix `flytectl/` followed by a semver string, e.g. [flytectl/v0.9.0](https://github.com/flyteorg/flyte/releases/tag/flytectl%2Fv0.9.0).

To release a new version of flytectl run the <[github workflow](https://github.com/flyteorg/flyte/blob/master/.github/workflows/flytectl-release.yml), which is responsible for releasing this new version. Be careful to use valid semver versions.
