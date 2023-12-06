# Flyte 1.10.6 Release

Due to a mishap in the move to the monorepo, we ended up generating the git tags between 1.10.1 to 1.10.5, so in order to decrease the confusion we decided to skip those patch versions and go straight to the next available version.

We've shipped a ton of stuff in this patch release, here are some of the highlights.

## GPU Accelerators
You'll be able to get more fine-grained in the use GPU Accelerators in your tasks. Here are some examples:

No preference of GPU accelerator to use:
```
@task(limits=Resources(gpu="1"))
def my_task() -> None:
    ...
```

Schedule on a specific GPU accelerator:
```
from flytekit.extras.accelerators import T4


@task(
    limits=Resources(gpu="1"),
    accelerator=T4,
)
def my_task() -> None:
    ...
```

Schedule on a Multi-instance GPU (MIG) accelerator with no preference of partition size:
```
from flytekit.extras.accelerators import A100


@task(
    limits=Resources(gpu="1"),
    accelerator=A100,
)
def my_task() -> None:
    ...
```

Schedule on a Multi-instance GPU (MIG) accelerator with a specific partition size:
```
from flytekit.extras.accelerators import A100


@task(
    limits=Resources(gpu="1"),
    accelerator=A100.partition_1g_5gb,
)
def my_task() -> None:
    ...
```

Schedule on an unpartitioned Multi-instance GPU (MIG) accelerator:
```
from flytekit.extras.accelerators import A100


@task(
    limits=Resources(gpu="1"),
    accelerator=A100.unpartitioned,
)
def my_task() -> None:
    ...
```

## Improved support for Ray logs
https://github.com/flyteorg/flyte/pull/4266 opens the door for RayJob logs to be persisted.

In https://github.com/flyteorg/flyte/pull/4397 we added support for a link to a Ray dashboard to show up in the task card.

## Updated grafana dashboards
We updated the official grafana dashboards in https://github.com/flyteorg/flyte/pull/4382.

## Support for Azure AD
A new version of our stow fork added support for Azure AD in https://github.com/flyteorg/stow/pull/9.

## Full changelog:
* Restructure Flyte releases by @eapolinario in https://github.com/flyteorg/flyte/pull/4304
* Use debian bookworm as single binary base image  by @eapolinario in https://github.com/flyteorg/flyte/pull/4311
* Use local version in single-binary by @eapolinario in https://github.com/flyteorg/flyte/pull/4294
* Accessibility for README by @mishmanners in https://github.com/flyteorg/flyte/pull/4322
* Add tests in `flytepropeller/pkg /controller/executors`  from 72.3% to 87.3% coverage by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4276
* fix: remove unused setting in deployment charts by @HeetVekariya in https://github.com/flyteorg/flyte/pull/4252
* Document simplified retry behaviour introduced in #3902 by @fg91 in https://github.com/flyteorg/flyte/pull/4022
* Ray logs persistence by @jeevb in https://github.com/flyteorg/flyte/pull/4266
* Not revisiting task nodes and correctly incrementing parallelism by @hamersaw in https://github.com/flyteorg/flyte/pull/4318
* Fix RunPluginEndToEndTest util by @andresgomezfrr in https://github.com/flyteorg/flyte/pull/4342
* Tune sandbox readiness checks to ensure that sandbox is fully accessi… by @jeevb in https://github.com/flyteorg/flyte/pull/4348
* Chore: Ensure Stalebot doesn't close issues we've not yet triaged. by @brndnblck in https://github.com/flyteorg/flyte/pull/4352
* Do not automatically close stale issues by @eapolinario in https://github.com/flyteorg/flyte/pull/4353
* Fix: Set flyteadmin gRPC port to 80 in ingress if using TLS between load balancer and backend by @fg91 in https://github.com/flyteorg/flyte/pull/3964
* Support Databricks WebAPI 2.1 version and Support `existing_cluster_id` and `new_cluster` options to create a Job by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4361
* Fixing caching on maptasks when using partials by @hamersaw in https://github.com/flyteorg/flyte/pull/4344
* Fix read raw limit by @honnix in https://github.com/flyteorg/flyte/pull/4370
* minor fix to eks-starter.yaml by @guyarad in https://github.com/flyteorg/flyte/pull/4337
* Reporting running if the primary container status is not yet reported by @hamersaw in https://github.com/flyteorg/flyte/pull/4339
* completing retries even if minSuccesses are achieved by @hamersaw in https://github.com/flyteorg/flyte/pull/4338
* Add comment to auth scope by @wild-endeavor in https://github.com/flyteorg/flyte/pull/4341
* Update tests by @eapolinario in https://github.com/flyteorg/flyte/pull/4381
* Update order of cluster resources config to work with both uctl and flytectl by @neverett in https://github.com/flyteorg/flyte/pull/4373
* Update tests in single-binary by @eapolinario in https://github.com/flyteorg/flyte/pull/4383
* Passthrough unique node ID in task execution ID for generating log te… by @jeevb in https://github.com/flyteorg/flyte/pull/4380
* Add Sections in the PR Template by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4367
* Update metadata in ArrayNode TaskExecutionEvents by @hamersaw in https://github.com/flyteorg/flyte/pull/4355
* Fixes list formatting in flytepropeller arch docs by @thomasjpfan in https://github.com/flyteorg/flyte/pull/4345
* Update boilerplate end2end tests by @hamersaw in https://github.com/flyteorg/flyte/pull/4393
* Handle all ray job statuses by @EngHabu in https://github.com/flyteorg/flyte/pull/4389
* Relocate sandbox config by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4385
* Refactor task logs framework by @jeevb in https://github.com/flyteorg/flyte/pull/4396
* Add support for displaying the Ray dashboard when a RayJob is active by @jeevb in https://github.com/flyteorg/flyte/pull/4397
* Disable path filtering for monorepo components by @eapolinario in https://github.com/flyteorg/flyte/pull/4404
* Silence NotFound when get task resource by @honnix in https://github.com/flyteorg/flyte/pull/4388
* adding consoleUrl parameterization based on partition by @lauralindy in https://github.com/flyteorg/flyte/pull/4375
* [Docs] Sensor Agent Doc by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4195
* [flytepropeller]  Add Tests in v1alpha.go including `array_test.go`, `branch_test.go`, `error_test.go`, and `iface_test.go` with 0.13% Coverage  Improvement by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4234
* Add more context for ray log template links by @jeevb in https://github.com/flyteorg/flyte/pull/4416
* Add ClusterRole config for Ray by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/4405
* Fix and update scripts for generating grafana dashboards by @Tom-Newton in https://github.com/flyteorg/flyte/pull/4382
* Add artifacts branch to publish to buf on push by @squiishyy in https://github.com/flyteorg/flyte/pull/4450
* Add service monitor for flyte admin and propeller service by @vraiyaninv in https://github.com/flyteorg/flyte/pull/4427
* Fix Kubeflow TF Operator `GetTaskPhase` Bug by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4469
* Instrument opentelemetry by @hamersaw in https://github.com/flyteorg/flyte/pull/4357
* Delete the .github folder from each subdirectory by @pingsutw in https://github.com/flyteorg/flyte/pull/4480
* Fix the loop variable scheduler issue by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/4468
* Databricks Plugin Setup Doc Enhancement by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4445
* Put ticker back in place in propeller gc by @eapolinario in https://github.com/flyteorg/flyte/pull/4490
* Store failed execution in flyteadmin by @iaroslav-ciupin in https://github.com/flyteorg/flyte/pull/4390
* Moving from flyteadmin - Upgrade coreos/go-oidc to v3 to pickup claims parsing fixes by @eapolinario in https://github.com/flyteorg/flyte/pull/4139
* Bump flyteorg/stow to 0.3.8 by @eapolinario in https://github.com/flyteorg/flyte/pull/4312
* Remove 'needs' from generate_flyte_manifest by @eapolinario in https://github.com/flyteorg/flyte/pull/4495
* Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/4302
* Modify how flytecopilot version is parsed from values file by @eapolinario in https://github.com/flyteorg/flyte/pull/4496
* Ignore component tags in goreleaser by @eapolinario in https://github.com/flyteorg/flyte/pull/4497
* Fix indentation of `shell: task` by @eapolinario in https://github.com/flyteorg/flyte/pull/4498
* Implemented simple echo plugin for testing by @hamersaw in https://github.com/flyteorg/flyte/pull/4489
* Correctly handle resource overrides in KF plugins by @jeevb in https://github.com/flyteorg/flyte/pull/4467
* Remove deprecated InjectDecoder by @EngHabu in https://github.com/flyteorg/flyte/pull/4507
* Fix $HOME resolution and webhook namespace by @EngHabu in https://github.com/flyteorg/flyte/pull/4509
* Add note on updating sandbox cluster configuration by @jeevb in https://github.com/flyteorg/flyte/pull/4510
* Add New PR Template by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4512
* [Docs] Databricks Agent Doc by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4008
* Bump version of goreleaser gh action to v5 by @eapolinario in https://github.com/flyteorg/flyte/pull/4519
* Kf operators use `GetReplicaFunc` (Error Handling) by @Future-Outlier in https://github.com/flyteorg/flyte/pull/4471

## New Contributors
* @HeetVekariya made their first contribution in https://github.com/flyteorg/flyte/pull/4252
* @andresgomezfrr made their first contribution in https://github.com/flyteorg/flyte/pull/4342
* @brndnblck made their first contribution in https://github.com/flyteorg/flyte/pull/4352
* @guyarad made their first contribution in https://github.com/flyteorg/flyte/pull/4337
* @neverett made their first contribution in https://github.com/flyteorg/flyte/pull/4373
* @thomasjpfan made their first contribution in https://github.com/flyteorg/flyte/pull/4345
* @lauralindy made their first contribution in https://github.com/flyteorg/flyte/pull/4375
* @Tom-Newton made their first contribution in https://github.com/flyteorg/flyte/pull/4382
* @vraiyaninv made their first contribution in https://github.com/flyteorg/flyte/pull/4427
