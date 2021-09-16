# Release Process

[![hackmd-github-sync-badge](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w/badge)](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w)

## Versioning Scheme
Flyte components follow [SemVer conventions](https://semver.org/). In a nutshell, that means every component has a release version that follows this format: `MAJOR.MINOR.PATCH`. This includes the aggregated milestone release described below.

Flyte hasn't reached `v1.0.0` version yet. This version indicates a stable public API. After releasing `v1.0.0` any backward incompatible changes require a `MAJOR` version bump. The community and the team, however, strives to maintain backward compatibility with the current public APIs as much as possible and gracefully deprecating APIs when absolutely required.

Until Flyte reaches `v1.0.0`, we are maintaining the following definitions of versions `MAJOR.MINOR.PATCHbX`:

1. **MAJOR** is pinned to `0`.
1. **MINOR** is bumped on one of two occasions:
    1. A backward incompatible change or,

       The decision to whether to a particular change constitutes a breaking change or not is evaluated at the component level (e.g. flytekit, flyteadmin... etc.)
    1. A major feature push

       At the end of every quarter, the core maintainers will determine if, for each component, the accumlated commits warrant a major feature push.
1. **PATCH** is bumped on every milestone (defined below)
2. **bX** is bumped on every commit.

## Milestones
Flyte manages releases for the platform through github Milestones. There are two types of milestones:

1. Release train milestone: Each milestone follows the georgian calendar month. Here is the summary of what each milestone includes:
   * A PATCH version (unless there is a breaking change, in which case we will bump the MINOR version)
   * The latest versions of all of the flyte components. Tested for regressions.
   * A change log highlighting the changes that went in.
1. Feature release milestone: Each milestone will roughly follow a quarter (3 months). Here is a summary of what each milestone includes:
    * A MINOR version bump (always). 
    * A change log and a blog post highlighting the features that were shipped during the quarter.

    **NOTE** 
    
    Feature releases are expected to be big events for the platform. They may get delayed if an important feature isn't stable enough to be included in the release and may, therefore, be postponed by one or more release train milestones.
   
## Release train milestone process

### Resolve Issues
1. Create [a new milestone](https://github.com/flyteorg/flyte/milestones) if one doesn't exist.
1. Open [issues](https://github.com/flyteorg/flyte/issues) and filter by milestone and make sure they are either closed or moved over to the next milestone.
### Update EndToEndTests
1. Update [requirements](https://github.com/flyteorg/flytetools/blob/master/flytetester/requirements.txt#L1) to the desired (e.g. latest) [flytekit release](https://github.com/flyteorg/flytekit/releases).
1. Build a docker image and push to ghcr.io/flyteorg
   ```prompt
   make -C flytetester docker_build_push
   ```

### Start a release PR
1. Run [Generate Flyte Manifests workflow](https://github.com/flyteorg/flyte/actions/workflows/generate-flyte-manifest.yml). It’ll create a PR ([example](https://github.com/flyteorg/flyte/pull/888))
1. Update [docs version](https://github.com/flyteorg/flyte/blob/master/rsts/conf.py#L28) to match the milestone version.
1. Create a CHANGELOG file ([example](https://github.com/flyteorg/flyte/pull/888/files#diff-0c33dda4ecbd7e1116ddce683b5e143d85b22e43223ca258ecc571fb3b240a57))
1. Update [the sha](https://github.com/flyteorg/flyte/blob/master/end2end/tests/endtoend.yaml#L14) with the latest image released in #2
1. Wait for endtoend tests to finish then Merge PR.

### Create a release
1. Run [Create Flyte Release workflow](https://github.com/flyteorg/flyte/actions/workflows/create-release.yml):
   It will create a tag and then publish all deployment manifest in github release and will create a discussion thread in github release 
1. Close the milestone
1. Ping #core (slack channel) to: Send announcements about the milestone with the contents of the CHANGELOG to all social channels..
