# Flyte v1.13.0-rc1 Release Notes

## What's Changed
* Add CustomHeaderMatcher to pass additional headers by @andrewwdye in https://github.com/flyteorg/flyte/pull/5563
* Turn flyteidl and flytectl releases into manual gh workflows by @eapolinario in https://github.com/flyteorg/flyte/pull/5635
* docs: fix typo by @cratiu222 in https://github.com/flyteorg/flyte/pull/5643
* Use enable_deck=True in docs by @thomasjpfan in https://github.com/flyteorg/flyte/pull/5645
* Fix flyteidl release  checkout all tags by @eapolinario in https://github.com/flyteorg/flyte/pull/5646
* Install pyarrow in sandbox functional tests by @eapolinario in https://github.com/flyteorg/flyte/pull/5647
* docs: add documentation for configuring notifications in GCP by @desihsu in https://github.com/flyteorg/flyte/pull/5545
* Correct "sucessfile" to "successfile" by @shengyu7697 in https://github.com/flyteorg/flyte/pull/5652
* Fix ordering for custom template values in cluster resource controller by @katrogan in https://github.com/flyteorg/flyte/pull/5648
* Don't error when attempting to trigger schedules for inactive projects by @katrogan in https://github.com/flyteorg/flyte/pull/5649
* Update Flyte components - v1.13.1-rc0 by @flyte-bot in https://github.com/flyteorg/flyte/pull/5656
* Add offloaded path to literal by @katrogan in https://github.com/flyteorg/flyte/pull/5660
* Improve error messaging for invalid arguments by @pingsutw in https://github.com/flyteorg/flyte/pull/5658
* DOC-462 Update "Try Flyte in the browser" text by @neverett in https://github.com/flyteorg/flyte/pull/5654
* DOC-533 Remove outdated duplicate notification config content by @neverett in https://github.com/flyteorg/flyte/pull/5672
* Validate labels before creating flyte CRD by @pingsutw in https://github.com/flyteorg/flyte/pull/5671
* Add FLYTE_INTERNAL_POD_NAME environment variable that holds the pod name by @bgedik in https://github.com/flyteorg/flyte/pull/5616
* Upstream Using InMemory token cache for admin clientset in propeller by @pvditt in https://github.com/flyteorg/flyte/pull/5621
* [Bug] Update resource failures w/ Finalizers set  (#423) by @pvditt in https://github.com/flyteorg/flyte/pull/5673
* [BUG] array node eventing bump version by @pvditt in https://github.com/flyteorg/flyte/pull/5680
* Add custominfo to agents by @ddl-rliu in https://github.com/flyteorg/flyte/pull/5604
* [BUG] use deep copy of bit arrays when getting array node state by @pvditt in https://github.com/flyteorg/flyte/pull/5681
* More concise definition of launchplan by @eapolinario in https://github.com/flyteorg/flyte/pull/5682
* Auth/prevent lookup per call by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5686

## New Contributors
* @cratiu222 made their first contribution in https://github.com/flyteorg/flyte/pull/5643
* @desihsu made their first contribution in https://github.com/flyteorg/flyte/pull/5545
* @shengyu7697 made their first contribution in https://github.com/flyteorg/flyte/pull/5652
* @bgedik made their first contribution in https://github.com/flyteorg/flyte/pull/5616

**Full Changelog**: https://github.com/flyteorg/flyte/compare/flytectl/v0.9.1...v1.13.1-rc1
