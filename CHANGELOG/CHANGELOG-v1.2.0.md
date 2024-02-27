# Flyte 1.2 Release

## Platform
- Support for Ray (https://github.com/flyteorg/flyte/issues/2641) - Also see the [blog post](https://blog.flyte.org/ray-and-flyte).
- Execution names can be longer now, up to 63 characters (https://github.com/flyteorg/flyteadmin/pull/466)
- Offloading FlyteWorkflow CRD static workflow spec (https://github.com/flyteorg/flyte/issues/2705)
- Enabled FlytePropeller subqueue - this means that every time a pod is updated in the workflow it reevals for faster downstream scheduling
- Add container configuration to default pod template (https://github.com/flyteorg/flyte/issues/2703)
- Fixed issues with blobstore writes - GCS had duplicate writes and subworkflow inputs were rewritten on every evaluation, this meant slower evaluations
- Support external deletion of non-terminal map task subtasks (as a result of https://github.com/flyteorg/flyte/issues/2701)
- Better truncation of flytepropeller logs (https://github.com/flyteorg/flyte/issues/2818)
- Bug: Recovering subworkflows is now supported (https://github.com/flyteorg/flyte/issues/2840)
- Fix snowflake plugin (https://github.com/flyteorg/flyteplugins/pull/286)


## Flytekit
- Hugging Face plugin (https://github.com/flyteorg/flytekit/pull/1116)
- dbt plugin (https://github.com/flyteorg/flyte/issues/2202)
- cache overriding behavior is now open to all types (https://github.com/flyteorg/flyte/issues/2912)
- Bug: Fallback to pickling in the case of unknown types used Unions (https://github.com/flyteorg/flyte/issues/2823)
- [pyflyte run](https://docs.flyte.org/en/latest/api/flytekit/design/clis.html#pyflyte-run) now supports [imperative workflows](https://docs.flyte.org/en/latest/user_guide/basics/imperative_workflows.html)
- Newlines are now stripped from client secrets (https://github.com/flyteorg/flytekit/pull/1163)
- Ensure repeatability in the generation of cache keys in the case of dictionaries (https://github.com/flyteorg/flytekit/pull/1126)
- Support for multiple images in the yaml config file (https://github.com/flyteorg/flytekit/pull/1106)

And more. See the full changelog in https://github.com/flyteorg/flytekit/releases/tag/v1.2.0


## Flyteconsole
- fix: Make sure groups used in graph aren't undefined [#545](https://github.com/flyteorg/flyteconsole/pull/545)
- fix: Graph Center on initial render [#541](https://github.com/flyteorg/flyteconsole/pull/541)
- fix: Graph edge overlaps nodes [#542](https://github.com/flyteorg/flyteconsole/pull/542)
- Fix searchbar X button [#564](https://github.com/flyteorg/flyteconsole/pull/564)
- fix: Update timeline view to show dynamic wf internals on first render [#562](https://github.com/flyteorg/flyteconsole/pull/562)
- fix: Webmanifest missing crossorigin attribute [#566](https://github.com/flyteorg/flyteconsole/pull/566)
- fix: console showing subworkflows as unknown [#570](https://github.com/flyteorg/flyteconsole/pull/570)
- fix: Dict value loses 1 trailing character on UI Launch. [#561](https://github.com/flyteorg/flyteconsole/pull/561)
- fix: launchform validation [#557](https://github.com/flyteorg/flyteconsole/pull/557)
- fix: integrate timeline and graph tabs wrappers under one component [#572](https://github.com/flyteorg/flyteconsole/pull/572)
- added none type in union type [#577](https://github.com/flyteorg/flyteconsole/pull/577)
- minor: inputHelpers InputProps [#579](https://github.com/flyteorg/flyteconsole/pull/579)
- fix: fix test of launchform [#581](https://github.com/flyteorg/flyteconsole/pull/581)
- Pruning some unused packages [#583](https://github.com/flyteorg/flyteconsole/pull/583)
- fix: floor seconds to int in the edge case moment returns it as float [#582](https://github.com/flyteorg/flyteconsole/pull/582)
- fix: add BASE_URL to dev startup, open deeply nested urls [#589](https://github.com/flyteorg/flyteconsole/pull/589)
- fix: add default disabled state for only mine filter [#585](https://github.com/flyteorg/flyteconsole/pull/585)
- fixed graph/timeline support for launchplanRef [#601](https://github.com/flyteorg/flyteconsole/pull/601)
- fix: enable deeplinks in development [#602](https://github.com/flyteorg/flyteconsole/pull/602)
