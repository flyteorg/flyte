# Flyte 1.16.2

## Added

- Added support for configuring horizontal pod autoscaling for flyte admin in flyte-core helm chart ([#6625](https://github.com/flyteorg/flyte/pull/6625))
- Added support for typing.Union in flytecopilot / container tasks ([#6674](https://github.com/flyteorg/flyte/pull/6674))

## Changed

- Only return connector/agent plugins if enabled ([#6644](https://github.com/flyteorg/flyte/pull/6644))
- Support downloading old versions of flytectl by install.sh ([#6668](https://github.com/flyteorg/flyte/pull/6668))
- Improve datacatalog query performance with composite index on tags table ([#6672](https://github.com/flyteorg/flyte/pull/6672))
- Handle child node timeouts in conditional branch nodes ([#6678](https://github.com/flyteorg/flyte/pull/6678))
- Fix auto refresh cache iteration bug causing parent workflows to get stuck in a running state ([#6725](https://github.com/flyteorg/flyte/pull/6725))

## Removed

- Removed unused node selector functions and legacy environment variables ([#6614](https://github.com/flyteorg/flyte/pull/6614))
- Remove unused DefaultWorkflowActiveDeadline config ([#6688](https://github.com/flyteorg/flyte/pull/6688))

## Dependencies / Security 

- Bump urllib3 from 2.2.3 to 2.5.0 in /flytectl/docs ([#6728](https://github.com/flyteorg/flyte/pull/6728))
- Bump requests from 2.32.3 to 2.32.4 in /flytectl/docs ([#6727](https://github.com/flyteorg/flyte/pull/6727))

## Housekeeping

- Removed unused replace directives ([#6638](https://github.com/flyteorg/flyte/pull/6638))
- Removed use of deprecated pod phase ([#6640](https://github.com/flyteorg/flyte/pull/6640))
- Executor code cleanup ([#6629](https://github.com/flyteorg/flyte/pull/6629))
- Updated docker images to use Bitnami legacy repo ([#6631](https://github.com/flyteorg/flyte/pull/6631))
- Updated helm charts to use new twun.io helm repo ([#6726](https://github.com/flyteorg/flyte/pull/6726))

## Contributors

Special thanks to new contributors: @hefeiyun ([#6629](https://github.com/flyteorg/flyte/pull/6629)), @ihvol-freenome ([#6672](https://github.com/flyteorg/flyte/pull/6672)), @Sally-Yang-Jing-Ou ([#6725](https://github.com/flyteorg/flyte/pull/6725)), along with all returning contributors who made this release possible.

Full Changelog: https://github.com/flyteorg/flyte/compare/v1.16.1...v1.16.2
