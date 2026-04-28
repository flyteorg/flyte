# Flyte 1.16.6

## Added

- docs: add comment about passwordPath usage in helm chart ([#6895](https://github.com/flyteorg/flyte/pull/6895))

## Changed

- Use retryable failure for failed Ray job deployments ([#7153](https://github.com/flyteorg/flyte/pull/7153))
- Fix copilot TerminationGracePeriodSeconds set to nanoseconds instead of seconds ([#7183](https://github.com/flyteorg/flyte/pull/7183))
- Fix: Correctly set 'child data dir/output dir' in branch handler's Abort and Finalize ([#7117](https://github.com/flyteorg/flyte/pull/7117))
- Fix: Make flytectl hydrate pod template spec in task node overrides ([#7096](https://github.com/flyteorg/flyte/pull/7096))
- fix(sandbox-bundled): Use a yaml string for FLYTE_PLATFORM_INSECURE ([#7079](https://github.com/flyteorg/flyte/pull/7079))
- Fix: Allow toggling dynamic log links for Ray, Dask, Spark plugins ([#7193](https://github.com/flyteorg/flyte/pull/7193))
- Fix type error for suspended kubeflow jobs ([#7241](https://github.com/flyteorg/flyte/pull/7241))

## Dependencies / Security

- Bump pygments from 2.18.0 to 2.20.0 in /flytectl/docs ([#7118](https://github.com/flyteorg/flyte/pull/7118))
- Bump requests from 2.32.4 to 2.33.0 in /flytectl/docs ([#7095](https://github.com/flyteorg/flyte/pull/7095))
- Bump github.com/go-jose/go-jose/v3 from 3.0.4 to 3.0.5 ([#7149](https://github.com/flyteorg/flyte/pull/7149))
- Upgrade viper to v1.21.0 and fix case-sensitive key handling ([#7013](https://github.com/flyteorg/flyte/pull/7013))
- Update to Go 1.25 ([#7201](https://github.com/flyteorg/flyte/pull/7201))
- Bump the go_modules group across 10 directories with 3 updates ([#7216](https://github.com/flyteorg/flyte/pull/7216))
- Bump the go_modules group across 5 directories with 2 updates ([#7228](https://github.com/flyteorg/flyte/pull/7228))
- Bump the go_modules group across 4 directories with 1 update ([#7256](https://github.com/flyteorg/flyte/pull/7256))
- Bump urllib3 from 2.6.0 to 2.6.3 in /flytectl/docs ([#6849](https://github.com/flyteorg/flyte/pull/6849))

## New Contributors
* @spwoodcock made their first contribution in https://github.com/flyteorg/flyte/pull/6895
* @strigazi made their first contribution in https://github.com/flyteorg/flyte/pull/7241
* @bergman made their first contribution in https://github.com/flyteorg/flyte/pull/7183

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.5...v1.16.6
