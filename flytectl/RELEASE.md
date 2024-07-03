# Release Process

Flytectl releases map to git tags with the prefix `flytectl/` followed by a semver string, e.g. [flytectl/v0.9.0](https://github.com/flyteorg/flyte/releases/tag/flytectl%2Fv0.9.0).

To release a new version of flytectl push a new git tag in the format described above. This will kick off a <[github workflow](https://github.com/flyteorg/flyte/blob/master/.github/workflows/flytectl-release.yml) responsible for releasing this new version. Note how the git tag has to be formatted a certain way for the workflow to run.
