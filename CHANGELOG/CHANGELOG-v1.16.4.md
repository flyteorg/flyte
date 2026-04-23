# Flyte 1.16.4

## What's Changed
* Bump flyteidl 1 to python 3.13 by @wild-endeavor in https://github.com/flyteorg/flyte/pull/6782
* Remove unused replace directives and have boilerplate use local flytestdlib by @Sovietaced in https://github.com/flyteorg/flyte/pull/6785
* Bump urllib3 from 2.5.0 to 2.6.0 in /flytectl/docs by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/6788
* refactor: omit unnecessary reassignment by @rifeplight in https://github.com/flyteorg/flyte/pull/6768
* [Feat] add user annotations to k8s objects by @ttitsworth-lila in https://github.com/flyteorg/flyte/pull/6710
* chore: execute goimports to format the code by @findfluctuate in https://github.com/flyteorg/flyte/pull/6801
* Fix issue with assertions with time zones by @Sovietaced in https://github.com/flyteorg/flyte/pull/6812
* [FlyteCTL] only get v1 sandbox image by @machichima in https://github.com/flyteorg/flyte/pull/6834
* Unpin k8s client library version by @Sovietaced in https://github.com/flyteorg/flyte/pull/6835
* Unpin controller runtime dependency by @Sovietaced in https://github.com/flyteorg/flyte/pull/6843
* Fix: Correctly handle malformed dynamic workflows to avoid 'failed + succeeded + running' Schroedinger state by @fg91 in https://github.com/flyteorg/flyte/pull/6854
* fixed: Remove print that breaks json output by @honnix in https://github.com/flyteorg/flyte/pull/6872
* Add test to decode access token from cookie values by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6883
* Update Kubernetes Dashboard URL to retired GitHub repository by @kevinliao852 in https://github.com/flyteorg/flyte/pull/6882
* Support flag in ray plugin config to optionally disable ingress for Ray cluster by @mickjermsurawong-openai in https://github.com/flyteorg/flyte/pull/6905
* Prevent panic when retry attempt value in map task gets larger than bitarray was allocated for by @fg91 in https://github.com/flyteorg/flyte/pull/6802

## New Contributors
* @rifeplight made their first contribution in https://github.com/flyteorg/flyte/pull/6768
* @ttitsworth-lila made their first contribution in https://github.com/flyteorg/flyte/pull/6710
* @findfluctuate made their first contribution in https://github.com/flyteorg/flyte/pull/6801
* @mickjermsurawong-openai made their first contribution in https://github.com/flyteorg/flyte/pull/6905

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.3...v1.16.4
