# Release Process

Flytectl releases map to git tags with the prefix `flytectl/` followed by a semver string, e.g. [flytectl/v0.9.0](https://github.com/flyteorg/flyte/releases/tag/flytectl%2Fv0.9.0).

To release a new version of flytectl run the <[github workflow](https://github.com/flyteorg/flyte/blob/master/.github/workflows/flytectl-release.yml), which is responsible for releasing this new version. Remember to use valid semver versions, including adding the prefix `v`, e.g. `v1.2.3`.

# How to yank a release?

Keep in mind that if you remove the git tag corresponding to the latest version before fixing homebrew as per the instructions in the section below then installing flytectl via homebrew will be broken. Consider the option of releasing a new version instead.

## From homebrew

We store the flytectl homebrew formula in https://github.com/flyteorg/homebrew-tap. Notice how only a specific version is exposed, so if the version you need to yank is the latest version, simply push a commit pointing to an earlier version (which should be a previous commit). 

## From the `install.sh` script
Remove the git tag corresponding to the release from the repo (e.g. `flytectl/v0.9.3`) to force the corresponding version to not be returned in the call to list versions in both the [install.sh script](https://github.com/flyteorg/flyte/blob/master/flytectl/install.sh).

