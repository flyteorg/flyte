# Flyte v1.0.2 Changelog

## General
1. [Housekeeping](https://github.com/flyteorg/flyte/pull/2572) Update single binary component in releases and bump version of contour helm chart


## Platform
1. [Bugfix](https://github.com/flyteorg/flyte/pull/2539) fix flyte-deps helm chart version
1. [Bugfix](https://github.com/flyteorg/flyte/pull/2542) Re-order clusterresourcesync annotations in helm chart
1. [Feature](https://github.com/flyteorg/flyte/issues/2516) Server-side compiler should strip Type Metadata
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2444) With GRPC v1.46.0 non-ascii chars are not permitted in grpc metadata
1. [Housekeeping](https://github.com/flyteorg/flyte/issues/1698) Configure grpc_health_prob in admin
1. [Feature](https://github.com/flyteorg/flyte/issues/2329) In Flytectl use Launchplan with latest version for scheduled workflows
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2262) Pods started before InjectFinalizer is disabled are never deleted
1. [Housekeeping](https://github.com/flyteorg/flyte/issues/2504) Checksum grpc_health_probe
1. [Feature](https://github.com/flyteorg/flyte/issues/2284) Allow to choose Spot Instances at workflow start time
1. [Feature](https://github.com/flyteorg/flyte/pull/2439) Use the same pod annotation formatting in syncresources cronjob
1. [Housekeeping](https://github.com/flyteorg/flyte/pull/2446) Migrate flyte sandbox docker image to use flyte-dep & flyte-core
1. [Feature](https://github.com/flyteorg/flyteidl/pull/300) Buf integration for proto release


## Flyteconsole
1. [Refactor](https://github.com/flyteorg/flyteconsole/issues/431) Move to monorepo structure to allow separate FlyteConsole into plugins system
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/414) Add support for StructuredDataSet Input/Output type
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/448) Updated Task details page to be able to browse different task versions
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/445) Updated Input/Output and TaskDetails representation - now you can collapse parts of it.
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/311) Show new map tasks in the GraphView with information of which subtasks are in which phase(running, erroring, succeeded)
1. [Feature](https://github.com/flyteorg/flyte/issues/2284) Add interruptible override to launch forms
1. [Bugfix](https://github.com/flyteorg/flyteconsole/issues/463) ensure that setups with different domains for console and admin API properly works
1. [Bugfix](https://github.com/flyteorg/flyteconsole/issues/416) Fixes one of the crashes in Graph component for Viel All Workflow section
1. [Bugfix](https://github.com/flyteorg/flyteconsole/issues/465) ensure that admin version is fully shown FC#465
1. [Bugfix] Small cosmetic updates: https://github.com/flyteorg/flyteconsole/issues/451, https://github.com/flyteorg/flyteconsole/issues/386, and https://github.com/flyteorg/flyteconsole/issues/398
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/312) Map Tasks - allow to check statuses of all retries per child task
1. [Feature](https://github.com/flyteorg/flyteconsole/issues/461) Allow to rerun single task in execution page
1. [Plugins]: Adds flyte-api plugin package. You can check basic how to info and consume package from https://www.npmjs.com/package/@flyteconsole/flyte-api . It will allow you to authorize you flytectl queries and perform api requests from your app. At this point only request without included data in body is allowed. More work is planned in future updates.


## Flytekit
1. [Feature](https://github.com/flyteorg/flyte/issues/2471) pyflyte run should support executing tasks
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2476) Dot separated python packages does not work for pyflyte
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2474) Pyflyte run doesn't respect the --config flag
1. [Bugfix](https://github.com/flyteorg/flytekit/pull/1002) Read packages from environment variables

