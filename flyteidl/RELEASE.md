# Release Process

Flyteidl releases map to git tags with the prefix `flyteidl/` followed by a semver string, e.g. [flytectl/v0.9.0](https://github.com/flyteorg/flyte/releases/tag/flytectl%2Fv0.9.0).

To release a new version of flytectl run the <[github workflow](https://github.com/flyteorg/flyte/blob/master/.github/workflows/flyteidl-release.yml), which is responsible for releasing this new version. Be careful to use valid semver versions.

