# Flyte v1.9.0 Release

In this release we're announcing two experimental features, namely (1) ArrayNode map tasks, and (2) Execution Tags.


### ArrayNode map tasks

ArrayNodes are described more fully in [RFC 3346](https://github.com/flyteorg/flyte/blob/master/rfc/system/3346-array-node.md), but the summary is that ArrayNode map tasks are a drop-in replacement for [regular map tasks](https://docs.flyte.org/en/latest/user-guide/advanced_composition/map_tasks.html), the only difference being the submodule used to import the `map_task` function.
More explicitly, let's say you have this code:

```python
from typing import List
from flytekit import map_task, task, workflow

@task
def t(a: int) -> int:
    ...

@workflow
def wf(xs: List[int]) -> List[int]:
    return map_task(t)(a=xs)
```

In order to switch to using array node map tasks you should import map_task from the `flytekit.experimental` module like so:

```python
from typing import List
from flytekit import task, workflow
from flytekit.experimental import map_task

@task
def t(a: int) -> int:
    ...

@workflow
def wf(xs: List[int]) -> List[int]:
    return map_task(t)(a=xs)
```


### Execution tags

Execution tags allow users to can discover their executions and other flyte entities more easily, by creating smarter groupings. The feature is described in this [RFC](https://github.com/flyteorg/flyte/blob/master/rfc/system/0001-flyte-execution-tags.md).

As mentioned before, this feature is shipped in an experimental capacity, the idea being that we're going to incorporate the feedback of the community as we iterate. More work is expected to give prominence to the feature in flyteconsole, in the meanwhile, the feature is supported via [Remote](https://docs.flyte.org/en/latest/api/flytekit/remote.html#remote-access).


## Flytekit
* Improve error handling in ShellTask by @pradithya in https://github.com/flyteorg/flytekit/pull/1732
* use default settings for timeline deck width by @cosmicBboy in https://github.com/flyteorg/flytekit/pull/1748
* Raise an exception in case of local execution of raw containers tasks by @eapolinario in https://github.com/flyteorg/flytekit/pull/1745
* Update contributing.rst by @eapolinario in https://github.com/flyteorg/flytekit/pull/1753
* Skip problematic pyyaml versions by @eapolinario in https://github.com/flyteorg/flytekit/pull/1752
* Fail CI tests faster by @eapolinario in https://github.com/flyteorg/flytekit/pull/1756
* Run unit tests on macos-latest by @eapolinario in https://github.com/flyteorg/flytekit/pull/1749
* add rdzv_configs to kfpytorch elastic by @Nan2018 in https://github.com/flyteorg/flytekit/pull/1751
* Fix: Disable rich logging handler when env var `FLYTE_SDK_RICH_TRACEBACKS=0` is set by @fg91 in https://github.com/flyteorg/flytekit/pull/1760
* Throw warning for nested @Task functions by @oliverhu in https://github.com/flyteorg/flytekit/pull/1727
* Remove dependency on responses by @honnix in https://github.com/flyteorg/flytekit/pull/1762
* Fix mlflow test error by @pingsutw in https://github.com/flyteorg/flytekit/pull/1766
* Use phrase 'git revision SHA1' in comments and docs by @DavidMertz in https://github.com/flyteorg/flytekit/pull/1761
* Dockerize docs requirements generation by @eapolinario in https://github.com/flyteorg/flytekit/pull/1764
* Restrict grpcio<1.53.1 by @eapolinario in https://github.com/flyteorg/flytekit/pull/1767
* Array node map task by @eapolinario in https://github.com/flyteorg/flytekit/pull/1640
* Add agent ctrl-c handler to call the delete function. (Reupload) by @Future-Outlier in https://github.com/flyteorg/flytekit/pull/1782
* Add tags to execution by @pingsutw in https://github.com/flyteorg/flytekit/pull/1723

## Flyteadmin
* Alter ID sequence to bigint by @honnix in https://github.com/flyteorg/flyteadmin/pull/578
* Remove content md5 requirement by @wild-endeavor in https://github.com/flyteorg/flyteadmin/pull/587
* Propagate request id on incoming and outgoing requests by @EngHabu in https://github.com/flyteorg/flyteadmin/pull/582
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flyteadmin/pull/589
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flyteadmin/pull/594
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flyteadmin/pull/596
* Bumping flytestdlib by @gvashishtha in https://github.com/flyteorg/flyteadmin/pull/597
* Flyte Execution tags by @pingsutw in https://github.com/flyteorg/flyteadmin/pull/571

## Flyteplugins
* Add support for using task execution ID fields in log URI templates by @jeevb in https://github.com/flyteorg/flyteplugins/pull/372
* Fix generate check in CI by @jeevb in https://github.com/flyteorg/flyteplugins/pull/377
* Remove welcomebot from boilerplate by @eapolinario in https://github.com/flyteorg/flyteplugins/pull/375
* Carry over hash value for all literal types in remote caching by @nicholasjng in https://github.com/flyteorg/flyteplugins/pull/378
* Send task execution metadata to out-core plugin by @honnix in https://github.com/flyteorg/flyteplugins/pull/369
* Support gRPC config for connector-service plugin by @honnix in https://github.com/flyteorg/flyteplugins/pull/368
* Use agent as name where it fits by @honnix in https://github.com/flyteorg/flyteplugins/pull/381
* Fix deletion of elastic task resource requests by @fg91 in https://github.com/flyteorg/flyteplugins/pull/379

## Flytepropeller
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flytepropeller/pull/591
* fixing max parallelism by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/594
* Updated flytestdlib 1.0.20 by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/593
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flytepropeller/pull/597
* Instrument ArrayNode by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/550
* make singular unions castable to their underlying type by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/599
* correct propagation of launchplan start error by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/598
* Bumping flytestdlib and stow versions by @gvashishtha in https://github.com/flyteorg/flytepropeller/pull/602
* Update boilerplate version by @flyte-bot in https://github.com/flyteorg/flytepropeller/pull/601
* Bump flyteplugins to 1.1.15 by @eapolinario in https://github.com/flyteorg/flytepropeller/pull/603
* updated flyteplugins to 1.1.16 by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/604

## Flyteconsole
* feat: Add localStorage for selected Project/Domain by @jsonporter in https://github.com/flyteorg/flyteconsole/pull/774
* Fix project selector failing test by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/780
* fix: node executions list going blank by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/788
* fix: launch form fixes by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/785
* chore: fix 404 due to bad state by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/797
* Fix: Launch Form fixes by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/798
* fix: map tasks should report caching status accurately by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/800
* feat: add support from structured datasets by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/801
* Top level masonry refactor by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/771
* fix: Tasks status out of sync by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/802
* chore: fix test_coverage by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/803
* Fix: breadcrumb feature flag priority order by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/804
* chore: fix yarn.lock by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/805
* fix: releases by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/806
* chore: fix release retry by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/807
* fix: specify node18, semantic release with exec and git plugins by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/808
* fix: commit linter running on generated messages by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/810
* chore: remove release git step by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/811
* fix: union value handling in launch form by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/812

## New Contributors
* @Nan2018 made their first contribution in https://github.com/flyteorg/flytekit/pull/1751
* @oliverhu made their first contribution in https://github.com/flyteorg/flytekit/pull/1727
* @DavidMertz made their first contribution in https://github.com/flyteorg/flytekit/pull/1761
* @Future-Outlier made their first contribution in https://github.com/flyteorg/flytekit/pull/1782
* @gvashishtha made their first contribution in https://github.com/flyteorg/flyteadmin/pull/597
* @nicholasjng made their first contribution in https://github.com/flyteorg/flyteplugins/pull/378
* @gvashishtha made their first contribution in https://github.com/flyteorg/flytepropeller/pull/602
