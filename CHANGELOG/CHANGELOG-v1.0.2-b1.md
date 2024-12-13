# Flyte v1.0.2-b1 Changelog

## Platform
1. [Feature](https://github.com/flyteorg/flyte/issues/2516) Server-side compiler should strip Type Metadata
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2444) With GRPC v1.46.0 non-ascii chars are not permitted in grpc metadata
1. [Housekeeping](https://github.com/flyteorg/flyte/issues/1698) Configure grpc_health_prob in admin
1. [Feature](https://github.com/flyteorg/flyte/issues/2329) In Flytectl use Launchplan with latest version for scheduled workflows
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2262) Pods started before InjectFinalizer was disabled are never deleted
1. [Housekeeping](https://github.com/flyteorg/flyte/issues/2504) Checksum grpc_health_probe
1. [Feature](https://github.com/flyteorg/flyte/issues/2284) Allow to choose Spot Instances at workflow start time
1. [Feature](https://github.com/flyteorg/flyte/pull/2439) Use the same pod annotation formatting in syncresources cronjob


## Flytekit
1. [Feature](https://github.com/flyteorg/flyte/issues/2471) pyflyte run should support executing tasks
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2476) Dot separated python packages does not work for pyflyte
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2474) Pyflyte run doesn't respect the --config flag
1. [Bugfix](https://github.com/flyteorg/flytekit/pull/1002) Read packages from environment variables
