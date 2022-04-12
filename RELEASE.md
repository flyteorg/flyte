# Release Process

[![hackmd-github-sync-badge](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w/badge)](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w)

## Resolve Issues
1. Create [a new milestone](https://github.com/flyteorg/flyte/milestones) if one doesn't exist.
1. Open [issues](https://github.com/flyteorg/flyte/issues) and filter by milestone and make sure they are either closed or moved over to the next milestone.
## Start a release PR
1. Run [Generate Flyte Manifests workflow](https://github.com/flyteorg/flyte/actions/workflows/generate-flyte-manifest.yml). It’ll create a PR ([example](https://github.com/flyteorg/flyte/pull/888))
1. Update [docs version](https://github.com/flyteorg/flyte/blob/master/rsts/conf.py#L28) to match the milestone version.
1. Create a CHANGELOG file ([example](https://github.com/flyteorg/flyte/pull/888/files#diff-0c33dda4ecbd7e1116ddce683b5e143d85b22e43223ca258ecc571fb3b240a57))
1. Wait for endtoend tests to finish then Merge PR.
## Create a release
1. Run [Create Flyte Release workflow](https://github.com/flyteorg/flyte/actions/workflows/create-release.yml):
   It will create a tag and then publish all deployment manifest in github release and will create a discussion thread in github release 
1. Close the milestone
1. Ping #core (slack channel) to: Send announcements about the milestone with the contents of the CHANGELOG to all social channels..
