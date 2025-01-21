# [RFC Template] Title

**Authors:**

- @eapolinario
- @katrogan

## 1 Executive Summary

This is a proposal to implement workflow execution concurrency, defined at the launch plan level. Concurrency applies to all version of a launch plan.

## 2 Motivation

See the following issues
1. https://github.com/flyteorg/flyte/issues/267
2. https://github.com/flyteorg/flyte/issues/420
3. https://github.com/flyteorg/flyte/discussions/3754
4. https://github.com/flyteorg/flyte/issues/5125

## 3 Proposed Implementation

Introduce a new attribute in [LaunchPlan.get_or_create](https://github.com/flyteorg/flytekit/blob/bc2e000cc8d710ed3d135cdbf3cbf257c5da8100/flytekit/core/launch_plan.py#L195) to allow specifying execution concurrency

e.g.     
```python
my_lp = LaunchPlan.get_or_create(
    name="my_serial_lp",
    workflow=my_wf,
    ...
    concurrency=Concurrency(
        max=1,  # defines how many executions with this launch plan can run in parallel
        policy=ConcurrencyPolicy.WAIT  # defines the policy to apply when the max concurrency is reached
    )
)
```

### FlyteIDL
We propose adding a new IDL message to capture concurrency behavior at CreateExecutionTime and embedding it in the existing [Schedule](https://github.com/flyteorg/flyte/blob/master/flyteidl/protos/flyteidl/admin/schedule.proto) message

```protobuf
message SchedulerPolicy {
  // Defines how many executions with this launch plan can run in parallel
  uint32 max = 1;
  
  // Defines how to handle the execution when the max concurrency is reached.
  ConcurrencyPolicy policy = 2;
}

enum ConcurrencyPolicy {
  UNSPECIFIED = 0;
  
  // wait for previous executions to terminate before starting a new one
  WAIT = 1;
  
  // fail the CreateExecution request and do not permit the execution to start
  ABORT = 2;
  
  // terminate the oldest execution when the concurrency limit is reached and immediately begin proceeding with the new execution
   REPLACE = 3;
}

message Schedule {
  ...

  SchedulerPolicy scheduler_policy = X; 
}

// embedded in the ExecutionClosure
message ExecutionStateChangeDetails  {
  ...
  
  // Includes the reason for the `PENDING` phase
  string description = X;
  
  
}

// Can also add to ExecutionSpec to specify execution time overrides

```
### Control Plane

At a broad level, we'll follow the precedent of the [scheduler](https://github.com/flyteorg/flyte/tree/master/flyteadmin/scheduler) defined in FlyteAdmin and define a singleton to manage concurrency across all launch plans.

1. At CreateExecution time, if the active version of the launch plan in the ExecutionSpec has a concurrency policy
   1. Create the execution in the database with a new `PENDING` execution phase and reason populated in `ExecutionStateChangeDetails`.
      1. or fail the request when the concurrency policy is set to `ABORT`
      2. let the concurrency controller manage scheduling
   1. Do not create the workflow CRD

### Concurrency Controller Singleton

Introduce the Concurrency Controller to poll for all pending executions:
1. Upon start-up, initialize a launch plan informer and a worker pool and spawn N number of worker threads.
   1. The launch plan informer will be responsible for keeping a map of launch plans, by [NamedEntityIdentifier](https://github.com/flyteorg/flyte/blob/25cfe16940f10f9bbef02e288c823db16eb37609/flyteidl/protos/flyteidl/admin/common.proto) (for each active launch plan version) and their concurrency policy: `map[admin.NamedEntityIdentifier]admin.SchedulerPolicy`
1. Periodically query the DB for pending executions `SELECT * FROM executions WHERE phase not in ('SUCCEEDED', 'FAILED', 'ABORTED', 'TIMED_OUT');`
   1. For each `PENDING` execution returned by the above query, `Add()` the pending execution to a [workqueue](https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go). We can fine tune in the future to include differentiated priority.
   1. For each non-`PENDING` execution returned by the above query, update the map value for the specific launch plan named entity using a thread-safe Map of type `rawActiveLaunchPlanExecutions map[admin.NamedEntityIdentifier]util.Set[admin.Execution]` (e.g. using this [set]("k8s.io/apimachinery/pkg/util/sets") library)
1. After processing the complete set of non-terminal executions, transform the `rawActiveLaunchPlanExecutions` map into a thread-safe, ordered list of executions by creation time: `activeLaunchPlanExecutions map[admin.NamedEntityIdentifier][]*core.WorkflowExecutionIdentifier` using an implementation where different keys can be accessed concurrently.
1. For each worker in the workqueue:
   1. Check the in-memory map populated launch plan informer to see:
      1. If the launch plan no longer has a concurrency policy, proceed to create the execution, see below
      1. If the launch plan has an active concurrency policy and max executions has been reached: proceed to respect the concurrency policy:
         1. `WAIT`: do nothing
         1. `ABORT`: mark the execution as `FAILED` in the db with a sensible error explaining the concurrency policy was violated
         1. `REPLACE`: terminate the oldest execution for the execution's launch plan in `activeLaunchPlanExecutions`. If this succeeds, or it's already terminated, then proceed to create the new execution: see below
      1. If the launch plan has an active concurrency policy and max executions have not been reached:
         1. Proceed to create the execution, see below
      1. Finally, always mark the queue item as [Done()](https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go#L33)
      
Creating an execution
1. create the workflow CRD
   1. if the CRD already exists because we've previously processed this pending execution before we had a chance to update the DB state, swallow the already exists error gracefully
1. conditionally mark the execution as `QUEUED` in the db if it's already in `PENDING` or  `QUEUED` to not overwrite any events should the execution have already reported progress in the interim
   2. conditionally meaning use compare-and-swap semantics within a transaction
1. If creating the workflow CRD fails
   2. Use some form of retries (perhaps with back-off) to create the CRD
   3. If the CRD creation still fails, mark the execution as `FAILED` in the db if it's currently in `PENDING` or  `QUEUED`. This will remove its eligiblity from the pending loop
   4. If the CRD creation has failed, but the execution has moved beyond 'QUEUED' and has reported progress from flytepropeller in the interim (due to a network partition, such that the CRD was created but not successfully reported as such) - do not update the execution status in the DB and allow execution to progress to a terminal state
2. Upon successful creation of the workflow CRD **or** failure in step (iv) above to mark the execution as 'FAILED', append the execution identifier to `activeLaunchPlanExecutions` for the launch plan named entity

#### Launch Plan informer
This is an async process we run in the Concurrency Controller to ensure we have an eventually consistent view of launch plans.

Upon Concurrency Controller start-up, we'll query the DB for all active launch plans and populate a map of active launch plans: `map[admin.NamedEntityIdentifier]admin.SchedulerPolicy`

Periodically, the informer will re-issue the query, optionally filtering by [UpdatedAt](https://github.com/flyteorg/flyte/blob/master/datacatalog/pkg/repositories/models/base.go#L7) to only fetch launch plans that have been updated since the last query to repopulate the map. If all versions of a launch plan have been deactivated since the last time the query ran, we'll want to update the in memory map to empty out the launch plan scheduler policy.


### Flyte Admin changes
### Execution Manager
Because we fetch the launch plan to reconcile execution inputs at CreateExecution time, we'll have the concurrency policy available to us at the time of execution creation.
If there is no concurrency policy defined, we'll proceed as [normal](https://github.com/flyteorg/flyte/blob/f14348165ccdfb26f8509c0f1ef380a360e59c4d/flyteadmin/pkg/manager/impl/execution_manager.go#L1169-L1173) and create the workflow execution CRD and then create a database entry for the execution with phase `UNKNOWN`. This way, we don't incur any penalty for executions 

If there is a concurrency policy defined, if it's set to `ABORT` immediately fail the execution. Otherwise, we'll create the execution in the database with a new `PENDING` execution phase and reason populated in `ExecutionStateChangeDetails` _but will not create a workflow CRD_


#### Database
For performance, we can introduce new fields to denormalize the launch plan named entity the execution was launched by
In [models/execution.go](https://github.com/flyteorg/flyte/blob/25cfe16940f10f9bbef02e288c823db16eb37609/flyteadmin/pkg/repositories/models/execution.go)
```go
model Execution {
	...
	LaunchPlanProject string
	LaunchPlanDomain string
	LaunchPlanName string
}
````

We should consider adding an index to the executions table to include
- phase in (`PENDING`, `QUEUED`, `RUNNING`) only (in order to safeguard for well-populated flyteadmin instances with lots of completed, historical executions)

##### Concurrency by specified launch plan versions
Executions are always tied to the versioned launch plan that triggered them (see [here](https://github.com/flyteorg/flyte/blob/38883c721dac2875bdd2333f4cd56e757e81ea5f/flyteadmin/pkg/repositories/models/execution.go#L26))
However, this proposal only applies concurrency at the launch plan Named Entity level, that is by (project, domain, name) and across all versions. The currently active launch plan version will determine the concurrency policy that gets applied for all executions created with the launch plan NamedEntity.

Non-goal, but future proposal: If we wanted to support concurrency by launch plan versions, we'd introduce `LaunchPlanVersion` to the execution model and add duplicates but with update keys for the in memory maps to be by versioned launch plan rather than NamedEntityIdentifier.

We could update usage like so

```python
my_lp = LaunchPlan.get_or_create(
    name="my_serial_lp",
    workflow=my_wf,
    ...
    concurrency=Concurrency(
        max=1,  # defines how many executions with this launch plan can run in parallel
        policy=ConcurrencyPolicy.WAIT  # defines the policy to apply when the max concurrency is reached
        precision=ConcurrencyPrecision.LAUNCH_PLAN_VERSION
    )
)
```

and by default, when the precision is omitted the SDK could register the launch plan using `ConcurrencyPrecision.LAUNCH_PLAN`

We could update the concurrency protobuf definition like so:
```protobuf
message SchedulerPolicy {
   // Defines how many executions with this launch plan can run in parallel
   uint32 max = 1;

   // Defines how to handle the execution when the max concurrency is reached.
   ConcurrencyPolicy policy = 2;

   ConcurrencyLevel level = 3;
}

enum ConcurrencyLevel {
  // Applies concurrency limits across all launch plan versions.
  LAUNCH_PLAN = 0;
  
  // Applies concurrency at the versioned launch plan level
  LAUNCH_PLAN_VERSION = 1;
}
```

Note, in this proposal, registering a new version of the launch plan and setting it to active will determine the concurrency policy across all launch plan versions.

#### Prior Art
The flyteadmin native scheduler (https://github.com/flyteorg/flyte/tree/master/flyteadmin/scheduler) already implements a reconciliation loop to catch up on any missed schedules.


## 4 Metrics & Dashboards
- Time spent in PENDING: It's useful to understand the duration spent in PENDING before a launch plan transitions to RUNNING
- It may be useful for Flyte platform operators to also configure alerts if an execution stays in PENDING for too long of a threshold

## 5 Drawbacks

The [executions model](https://github.com/flyteorg/flyte/blob/38883c721dac2875bdd2333f4cd56e757e81ea5f/flyteadmin/pkg/repositories/models/execution.go)
already has indices on
- primary key index
- launch plan id
- workflow id
- task id (for single task executions)
- execution created at
- error kind (code)
- user who launched the execution
- state

Database performance suffers as new indices are added (ref [[1](https://use-the-index-luke.com/sql/dml/insert)] [[2](https://www.timescale.com/learn/postgresql-performance-tuning-optimizing-database-indexes)])

## 6 Alternatives

### Scheduling
This proposal purposefully uses random scheduling. But this does not preclude defining other scheduling orders or catch-up policies in the future.

To accomplish this, we can extend the `ConcurrenyPolicy` proto message to encapsulate scheduling behavior

```protobuf
message Concurrency {
  // Defines how many executions with this launch plan can run in parallel
  uint32 max = 1;
  
  // Defines how to handle the execution when the max concurrency is reached.
  ConcurrencyPolicy policy = 2;

  ConcurrencyScheduling scheduling = 3;
}


type ConcurrencyScheduling enum {
  FIFO = 0;
  FILO = 1;
  ...
}
```

When we process the pending executions in the Concurrency Controller, we can sort the pending executions by creation time in ascending or descending order based on the scheduling policy.

Furthermore, we may want to introduce a max pending period to fail executions that have been in `PENDING` for too long


## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions

- Should we always attempt to schedule pending executions in ascending order of creation time?
    - Decision: We'll use FIFO scheduling by default but can extend scheduling behavior with an enum going forward.
- Should we propagate concurrency policies to child executions?
    - Decision: no. Child executions can define concurrency at the child launch plan level if necessary.

## 9 Conclusion

This is a simple and lightweight means for limiting execution concurrency that we can build upon, for flexible scheduling policies and even limiting task execution concurrency.


**Checklist:**

- [x]  Copy template
- [x]  Draft RFC (think of it as a wireframe)
- [x]  Share as WIP with folks you trust to gut-check
- [x]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
