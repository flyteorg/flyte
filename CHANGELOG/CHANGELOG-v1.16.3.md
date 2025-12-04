# Flyte 1.16.3

## Added

- Add start- and end-time log link template vars to Ray plugin ([#6751](https://github.com/flyteorg/flyte/pull/6751))
- Add metric to track how long nodes are stuck blocked by max parallelism being reached ([#6737](https://github.com/flyteorg/flyte/pull/6737))

## Changed

- Update Flyte to Go 1.24 ([#6603](https://github.com/flyteorg/flyte/pull/6603))

## Fixed

- Upstream array node fixes + update mockery ([#6763](https://github.com/flyteorg/flyte/pull/6763))
- Handle etcd request too large error ([#6752](https://github.com/flyteorg/flyte/pull/6752))
- Immediately fail if plugin.BuildResource fails instead of retrying until system retry budget is exhausted ([#6740](https://github.com/flyteorg/flyte/pull/6740))
- Avoid merging pod template twice if template container name matches task container name ([#6733](https://github.com/flyteorg/flyte/pull/6733))

## Dependencies / Security

- Update x/crypto to v0.45.0 to fix security warnings ([#6774](https://github.com/flyteorg/flyte/pull/6774))
- Bump golang.org/x/crypto from 0.36.0 to 0.45.0 in /datacatalog ([#6764](https://github.com/flyteorg/flyte/pull/6764))
- Bump golang.org/x/crypto from 0.36.0 to 0.45.0 in /flyteadmin ([#6770](https://github.com/flyteorg/flyte/pull/6770))
- Pin neoeinstein-prost to v0.4.0 ([#6773](https://github.com/flyteorg/flyte/pull/6773))
- Bump stow to latest ([#6741](https://github.com/flyteorg/flyte/pull/6741))

## Housekeeping

- Update Flyte Components ([#8e488c4b6](https://github.com/flyteorg/flyte/commit/8e488c4b6))
- Update docs version ([#424d7775f](https://github.com/flyteorg/flyte/commit/424d7775f))

## Contributors

Thanks to all the contributors who made this release possible: @fg91 ([#6733](https://github.com/flyteorg/flyte/pull/6733), [#6740](https://github.com/flyteorg/flyte/pull/6740), [#6751](https://github.com/flyteorg/flyte/pull/6751)), @Sovietaced ([#6603](https://github.com/flyteorg/flyte/pull/6603), [#6773](https://github.com/flyteorg/flyte/pull/6773), [#6774](https://github.com/flyteorg/flyte/pull/6774)).

Full Changelog: https://github.com/flyteorg/flyte/compare/v1.16.2...v1.16.3