# [RFC] Eviction of cached task outputs

**Authors:**

- @MorpheusXAUT
- [blackshark.ai](https://blackshark.ai)

This document: https://hackmd.io/qOztkaj4Rb6ypodvGEowAg?view

## 1 Executive Summary

This RFC describes an approach for cache eviction for Flyte executions, both on a workflow and task level. This will allow for invalidation of single (or multiple) task outputs for a single execution, providing an easy way to remove potentially invalid output data or cleaning up no longer required previous runs without having to change the `cache_version` (thus requiring a re-calculation of other related data).

Additionally, we propose to expose the cache eviction functionality via `flyteadmin` as an API to clean up cached values of previously completed executions of workflows or tasks.

## 2 Motivation

Some large executions might produce invalid/undesirable results for some tasks - even though the execution doesn't fail, its output might still be faulty, requiring a re-calculation to be performed using identical inputs.

At the moment, the only way to force Flyte to ignore the (faulty) cached value of a previous execution during a re-run of the workflow would be to change the `cache_version` of the affected task, causing all other calculations to be performed again as well even though their results might've been correct in the first place.

Depending on the workload executed, this might result in unnecessary and potentially high costs and time requirements to re-process. Evicting some of the cached values would allow re-runs of previously executed workflows to leverage caching where desirable while still re-calculating some of the data that has been deemed "invalid".

In addition, exposing cache eviction functionality via a `flyteadmin` API call would allow Flyte users/administrators to perform housekeeping and clean up their cache, cleanly removing the cached data of previously executions that are no longer relevant executions and would otherwise incur continuous storage costs. By exposing the same functionality used during the cache eviction override, we can ensure the cache maintained by `datacatalog` is consistent with the underlying blob storage.

## 3 Proposed Implementation

Our proposed implementation consists of two parts, which - while related to one another - solve two slightly different problems:

### Cache eviction override for a single execution

Similar to the `interruptible` override flag provided for a single execution of a workflow or task, we propose adding a flag (e.g. `cache_override`) to evict the cached output of an execution and force its calculation to be performed (and cached) again.  
This cleanup would be performed automatically by Flyte during the execution of a new workflow or task, returning the updated results:

![cache eviction flowchart](https://i.imgur.com/0NeoYGy.png)

* During the execution preparation, `flytepropeller` would query `datacatalog` as before, trying to find the cached output for a task and its inputs.
* If no cache entry was found, the current behaviour remains the same - the task is executed and its result is stored (if caching has been enabled for the task).
* Should a cache hit occur, `flytepropeller` would check the `cache_override` flag of the execution config:
    * If the override has been set, `flytepropeller` would skip retrieving the entry from `datacatalog` and continue execution, overwriting the cached data with the updated results after the task finished successfully.
    * If no override has been set, the cached value is returned as before.
* ~~As a slight variation, the cache eviction of existing entries could be performed before the tasks is executed again (although that might require additional synchronization to avoid race conditions during concurrent executions).~~

We propose to add the new override flag to the existing execution config, allowing for single tasks to be picked out and re-computed on demand. Whilst this would also expose the override on a workflow level, support for evicting a whole workflow from cache during an execution is not strictly necessary as this would be very similar to incrementing the (already implemented) `cache_version` - using the workflow value as a default for all its tasks (if set) should be relatively easy to add though.

The following Flyte components would need to support this change:

#### `datacatalog`

`datacatalog` needs to support the eviction of its cache entries, removing them from both its database as well as the underlying blob storage.  
All major cloud providers currently supported should provide sufficient functionality in their SDK/CLI.

Current reservations need to be respected - if the requesting executor is not the current owner, we potentially need to wait until the previous reservation has been released or synchronize access to ensure we're not running into race conditions with simultaneous executions.

#### `flyteidl`

The protobuf definitions in `flyteidl` need to include a new (optional) parameter for enabling the cache eviction override for a single execution - this should most likely be added to the `WorkflowExecutionConfig`.

#### `flytepropeller`

`flytepropeller` needs to be extended to support the additional execution config option to evict cache entries as well as calling appropriate `datacatalog` APIs and handling synchronization.

#### `flyteadmin`

`flyteadmin` needs to accept an updated config for execution requests to support the eviction override as well as pass it along to `flytepropeller`. The new flag should be available via API to either be invoked programmatically (e.g. via a user script) or via `flyteconsole`.

#### `flyteconsole`

Similar to the `interruptible` override currently available, we propose to add an optional execution launch setting to `flyteconsole` to easily trigger a single task to be executed again using identical inputs.

#### `flytekit`

We should extend the [`flytekit.remote`](https://docs.flyte.org/projects/flytekit/en/latest/remote.html#remote-access) functionality to support setting the `skip_cache` flag for a single execution.

#### `flytectl`

`flytectl` should support the new `skip_cache` override while creating executions via CLI.

### Evicting cache of a past execution

In addition to the cache eviction override while launching a workflow, we propose to expose the same functionality via `flyteadmin` as an extension to the existing `AdminService`:

#### Extension of existing `flyteadmin` endpoints

We could extend existing `flyteadmin` endpoints (and introduce some new ones) to support eviction of cached workflow/task results:

* The [`UpdateExecution`](https://github.com/flyteorg/flyteidl/blob/ac94eb4c6ffbe17df5b146337135f59914ffe3bf/protos/flyteidl/service/admin.proto#L281) endpoints and its [`ExecutionUpdateRequest`](https://github.com/flyteorg/flyteidl/blob/e749b82a70027d3364f4e5b98ce2cfbffb4eef54/protos/flyteidl/admin/execution.proto#L351) could be extended to include an `evict_cache` flag.
    * If the new flag is set, `UpdateExecution` would evict the cache of the referenced workflow and all of its sub-workflows/-tasks.
    * If a [dynamic](https://github.com/flyteorg/flyteidl/blob/aded40037d1a4d2ea32c6dd18644ccf16072124f/protos/flyteidl/admin/node_execution.proto#L116) or [workflow](https://github.com/flyteorg/flyteidl/blob/aded40037d1a4d2ea32c6dd18644ccf16072124f/protos/flyteidl/admin/node_execution.proto#L108) node are encountered, `flyteadmin` would [list executions](https://github.com/flyteorg/flyteidl/blob/aded40037d1a4d2ea32c6dd18644ccf16072124f/protos/flyteidl/service/admin.proto#L343) for the node and recursively evict the cache of all child executions as well.
    * `flyteadmin` would tolerate partial failures while evicting the cache of an execution - any errors encountered would be propagated back to the user. Running the eviction again would attempt to remove the remaining cached entries.
* We further propose adding a new endpoint equivalent to `UpdateExecution` for handling tasks execution entities (e.g. `UpdateTaskExecution`), supporting the `evict_cache` flag mentioned above.
    * The task execution would be identified using a `TaskExecutionIdentifier`.
    * Whilst `UpdateTaskExecuion` would only contain the `evict_cache` flag at the moment, this approach would be preferred to not break convention of performing CRUD operations on the entity specified in the message name by e.g. re-using `UpdateExecution` to also handle tasks as well.

#### Standalone `flyteadmin` endpoints

As a (less preferred) alternative to the extension of existing functionality mentioned above, we could implement new, completely independent endpoints:

* Similar to `GetExecutionData`, provide an endpoint such as `EvictExecutionCache`, removing all cached outputs of all tasks within the specified workflow execution.
    * The workflow execution would be identified using a `WorkflowExecutionIdentifier`.
    * Execution of the workflow must have already concluded.
    * `flyteadmin` would automatically iterate all tasks included in the workflow and recursively clean up cached data as necessary.
    * If any reservations exist for the cached resources to be deleted, `flyteadmin` would wait until it can ensure no other executor is accessing the data.
* Similar to `GetTaskExecutionData`, provide an endpoint such as `EvictTaskExecutionCache`, removing the cached output of a specific task.
    * The task execution would be identified using a `TaskExecutionIdentifier`.
    * Execution of the task (and its parent workflow, if available) must have already concluded (successfully).
    * If any reservations exist for the cached resource to be deleted, `flyteadmin` would wait until it can ensure no other executor is accessing the data.

As most of the required functionality should already be added while implementing the cache eviction override for an execution, this additional change would likely only require some smaller extension of the `flyteadmin` service.

Whilst our main intent for this `AdminService` extension is for automated/scripted cleanup using user-provided scripts, this functionality could also be added to other Flyte tools:

#### `flytekit`

[`flytekit.remote`](https://docs.flyte.org/projects/flytekit/en/latest/remote.html#remote-access) could be extended to support eviction of a task/workflow's cached results remotely.

#### `flytectl`

`flytectl` should support the task/workflow execution update to evict the cache, similar to the existing `--activate` flag: `flytectl update execution <task/workflow execution identifier> --evictCache`.

#### `flyteconsole`

We propose extending `flyteconsole` to allow users to evict the cache of a finished execution via a button click (including confirmation) using the web UI.

## 4 Metrics & Dashboards

As the implementation of this API extension could potentially lead to additional load on `datacatalog` as well as synchronization delays (due to reservations), it might be useful to track:

* the number of cache evictions (most likely emitted by `flytepropeller`)
* the latency of cache eviction operations (emitted by `datacatalog`)
* the number and delay of synchronization required to evict cache entries/reservations (most likely emitted by `flytepropeller`)

## 5 Drawbacks

The proposed cache eviction changes introduces a slight overhead in execution preparation in case an eviction has been requested since a reservation must be acquired and the stored data has to be removed from `datacatalog` and its underlying blob storage.  
Depending on the size of the cached outputs and the blob storage speed, this might induce increased scheduling times for tasks, although no performance impact should occur if no cache eviction is performed.  
In order to minimize the execution startup delay, cache eviction could be postponed until the task executed successfully, only requiring a quick check beforehand and ensuring the actual computation can start as soon as possible.

We do not anticipate any noticeable impact by the API extension for `flyteadmin` during regular executions, however evicting all cached outputs of a large workflow could introduce some strain on `flyteadmin`/`datacatalog` during processing.

## 6 Alternatives

At the moment, no direct alternative exists to allow for partial re-processing of already cached tasks - the `cache_version` available is not fine-grained enough to achieve the desired outcome. As of now, Flyte users would have to manually remove the "broken" cache entries from `datacatalog`'s database and delete them from the blob storage, potentially resulting in an inconsistent state if not executed properly.

## 7 Potential Impact and Dependencies

As this change spans across most of Flyte's components and repositories, a coordinated effort to introduce and test this proposal will be required. Since we rely on an API change, a `flyteidl` release will be required before additional components can be adapted.

The potential for malicious exploitation is deemed non-existent as no access to any data is provided and most additional communication will be performed intra-cluster. An attacker could potentially cause additional computational costs by removing cached outputs.

## 8 Unresolved questions

1. When should cache eviction happen during a repeated execution?
   Should we evict the cache entry immediately before starting task execution, potentially leaving no cached output available if the execution fails, or postpone the removal until the task has finished (successfully) and new data can be stored right away?
    - **RESOLVED**: eviction will happen after the successful execution of a task, overwriting the existing data.
2. Should the `cache_override` flag also be added to `flytekit`'s task decorator?
   This would allow users to define tasks/workflows which will automatically evict their cached results, however does not strictly fit with the actual task "definition".
3. Which Flyte tools (`flyteconsole`/`flytectl`) should support the proposed `AdminService` API extension for `flyteadmin`, if any?
    - **RESOLVED**: `flytectl`, `flytekit.remote`, `flyteconsole`
4. Should we support automatic eviction of cached results on workflow archival (opt-out via `flyteconsole`)?
5. Should we evict [Infratask Checkpoints](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/control_flow/checkpoint.html) from the cache as well since they might return cached results? If so, should we evict them from the backend side or pass the `cache_override` flag along to `flytekit`/its `Checkpointer` to skip any available entries?
    - **RESOLVED**: not for the initial implementation. Infratask checkpoints are only relevant for consecutive retries of a task - their results would not be considered when launching another execution with a `cache_override` flag set. 

## 9 Conclusion

At the time of writing, Flyte does not provide any functionality to clean up its cache, especially in a targeted, selective manner. This proposal introduces a way for users to control caching behaviour and potentially cut down on computational costs of task re-runs. Whilst the backend portion is relatively clear already, some questions still remain open, especially for the frontend components.
