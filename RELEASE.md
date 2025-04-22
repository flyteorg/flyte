# Release Process

[![hackmd-github-sync-badge](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w/badge)](https://hackmd.io/sVOAyv6LTwiQllQUctxP1w)

The below are steps to release a new version of Flyte.
Please note that each step must be completed before proceeding to the next one.
Running steps in parallel may cause errors.


## 1. Release a new flyteidl version at PyPI
[Create a FlyteIDL release](https://github.com/flyteorg/flyte/actions/workflows/flyteidl-release.yml)

This will take you near 1 minute.

## 2. Create a Release PR and Merge it
1. Run [Generate Flyte Manifests workflow](https://github.com/flyteorg/flyte/actions/workflows/generate_flyte_manifest.yml). It’ll create a PR ([example](https://github.com/flyteorg/flyte/pull/888))
2. Update [docs version](https://github.com/flyteorg/flyte/blob/master/docs/conf.py#L33) to match the milestone version.
3. Create a CHANGELOG file ([example](https://github.com/flyteorg/flyte/pull/888/files#diff-0c33dda4ecbd7e1116ddce683b5e143d85b22e43223ca258ecc571fb3b240a57))
4. Wait for endtoend tests to finish then Merge PR.

## 3. Create a release
1. Run [Create Flyte Release workflow](https://github.com/flyteorg/flyte/actions/workflows/create_release.yml):
   It will create a tag and then publish all deployment manifest in github release and will create a discussion thread in github release
2. Kick off a run of the functional tests in https://github.com/unionai/genesis-device/actions/workflows/update_cluster_and_run_tests.yml
3. Close the milestone
4. Ping #core ([slack](https://slack.flyte.org/) channel) to send announcements about the milestone with the contents of the CHANGELOG to all social channels.

## (Optional) Resolve Issues
1. Create [a new milestone](https://github.com/flyteorg/flyte/milestones) if one doesn't exist.
2. Open [issues](https://github.com/flyteorg/flyte/issues) and filter by milestone and make sure they are either closed or moved over to the next milestone.
