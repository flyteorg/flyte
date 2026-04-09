# Flyte 1.16.5

## What's Changed
* Fix: Prevent dropping of empty string args when constructing RayJob entrypoint by @fg91 in https://github.com/flyteorg/flyte/pull/6947
* don't reuse parent wf state for array node subnodes by @pvditt in https://github.com/flyteorg/flyte/pull/6929
* Feat: Make RayCluster head node ingress optional by @fg91 in https://github.com/flyteorg/flyte/pull/6852
* Flyte 2 update by @kumare3 in https://github.com/flyteorg/flyte/pull/6961
* Add RBAC support for cross-namespace secret reading by @rohitrsh in https://github.com/flyteorg/flyte/pull/6919
* set run_all_sub_nodes for array node idl by @pvditt in https://github.com/flyteorg/flyte/pull/6966
* Add Config struct with DisableConfigEndpoint option to profutils profiling server. Register config section under "prof" key with pflags generation and conditionally skip the /config HTTP handler when disabled. by @EngHabu in https://github.com/flyteorg/flyte/pull/7016
* Modernize codespell config: move from inline workflow to .codespellrc by @yarikoptic in https://github.com/flyteorg/flyte/pull/7104

## New Contributors
* @rohitrsh made their first contribution in https://github.com/flyteorg/flyte/pull/6919

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.4...v1.16.5
