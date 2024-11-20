# [RFC Template] Title

**Authors:**

- @eapolinario
- @katrogan

## 1 Executive Summary

This is a proposal to implement workflow execution concurrency, defined at the launch plan level.

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
We propose adding a new IDL message to capture concurrency behavior at CreateExecutionTime

```protobuf
message Concurrency {
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

message LaunchPlanSpec {
  ...

  Concurrency concurrency = X; 
}

// embedded in the ExecutionClosure
message ExecutionStateChangeDetails  {
  ...
  
  // Includes the reason for the `PENDING` phase
  string description = X;
  
  
}

// Can also add to ExecutionSpec to specify execution time overrides

```

### FlyteAdmin
At a broad level
1. At CreateExecution time, if the launch plan in the ExecutionSpec has a concurrency policy
   1. Create the execution in the database with a new `PENDING` execution phase and reason populated in `ExecutionStateChangeDetails`.
      1. or fail the request when the concurrency policy is set to `ABORT`
   1. Do not create the workflow CRD

Introduce an async reconciliation loop in FlyteAdmin to poll for all pending executions:
1. 1x a minute: Query all pending executions by timestamp ascending, grouped by launch plan ID, roughly something like
```sql
SELECT e.*
FROM   executions AS e
WHERE  ( launch_plan_id, created_at ) IN (SELECT launch_plan_id,
                                                 Min(created_at)
                                          FROM   executions
                                          WHERE  phase = 'PENDING'
                                          GROUP  BY launch_plan_id); 
```
2. For each execution returned by the above query, `Add()` the pending execution to a [rate limiting workqueue](https://github.com/kubernetes/client-go/blob/master/util/workqueue/rate_limiting_queue.go#L27-L40) (as a suggestion)
3. In a separate goroutine, fetch items from the workqueue and individually process each execution entry
   1. Check the database to see if there are fewer than `MAX_CONCURRENCY` non-terminal executions matching the launch plan ID in the pending execution model
   ```sql
   select count(launch_plan_id) from executions where phase not in ('SUCCEEDED', 'FAILED', 'ABORTED', 'TIMED_OUT') group by launch_plan_id;
   ```
   1. If there are fewer than `MAX_CONCURRENCY` executions running
      1. check that the execution is still in `PENDING`
      1. create the workflow CRD
          1. if the CRD already exists because we've previously processed this pending execution before we had a chance to update the DB state, swallow the already exists error gracefully
      1. conditionally mark the execution as `QUEUED` in the db if it's already in `PENDING` or  `QUEUED` to not overwrite any events should the execution have already reported progress in the interim
      1. If creating the workflow CRD fails: mark the execution as `FAILED` in the db (and swallow any errors if the workflow is already in a terminal state should we have previously reported the failure after re-enqueuing the pending execution a previous loop). This will remove its eligiblity from the pending loop
   1. If there are already `MAX_CONCURRENCY` executions running, simply proceed to (iii.)
   1. Finally, always mark the queue item as [Done()](https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go#L33) 

If we wanted further parallelization here, we could introduce a worker pool rather than having one async process read from the workqueue.

We should consider adding an index to the executions table to include
- launch_plan_id
- phase==PENDING only (in order to safeguard for well-populated flyteadmin instances with lots of completed, historical executions)
- created_at

##### Concurrency across launch plan versions
Executions are always tied to the versioned launch plan that triggered them (see [here](https://github.com/flyteorg/flyte/blob/38883c721dac2875bdd2333f4cd56e757e81ea5f/flyteadmin/pkg/repositories/models/execution.go#L26))
Therefore, this proposal only applies concurrency at the versioned launch plan level.

If we wanted to support concurrency across launch plan versions:

We could update usage like so

```python
my_lp = LaunchPlan.get_or_create(
    name="my_serial_lp",
    workflow=my_wf,
    ...
    concurrency=Concurrency(
        max=1,  # defines how many executions with this launch plan can run in parallel
        policy=ConcurrencyPolicy.WAIT  # defines the policy to apply when the max concurrency is reached
        precision=ConcurrencyPrecision.LAUNCH_PLAN
    )
)
```

and by default, when the precision is omitted the SDK could register the launch plan using `ConcurrencyPrecision.LAUNCH_PLAN_VERSION`

We could update the concurrency protobuf definition like so:
```protobuf
message Concurrency {
   // Defines how many executions with this launch plan can run in parallel
   uint32 max = 1;

   // Defines how to handle the execution when the max concurrency is reached.
   ConcurrencyPolicy policy = 2;

   ConcurrencyLevel level = 3;
}

enum ConcurrencyLevel {
  UNSPECIFIED = 0;
  
  // Applies concurrency limits across all launch plan versions.
  LAUNCH_PLAN = 1;
  
  // Applies concurrency at the versioned launch plan level
  LAUNCH_PLAN_VERSION = 2;
}
```


We could add another index to the Executions table to include the launch plan named entity, that is the entry in the [NamedEntity table](https://github.com/flyteorg/flyte/blob/38883c721dac2875bdd2333f4cd56e757e81ea5f/flyteadmin/pkg/repositories/models/named_entity.go#L39-L42) corresponding to the launch plan project, domain & name

In models/execution.go:
```go

type Execution struct {
	
	...
	// Already exists
	LaunchPlanID uint   `gorm:"index"`
	// New field to make querying on the named entity 
	LaunchPlanNamedEntityID uint `gorm:"index"`
	// New field to make querying on concurrency policy by the reconciliation loop easier
	ConcurrencyLevel uint32
	
}


```

Then the reconciliation loop would query executions in a non-terminal phase matching the launch plan named entity ID instead of LaunchPlanID based on the ConcurrencyLevel.

```sql
SELECT e.*
FROM   executions AS e
WHERE  ( launch_plan_named_entity_id, created_at ) IN (SELECT launch_plan_named_entity_id,
                                                 Min(created_at)
                                          FROM   executions
                                          WHERE  phase = 'PENDING' AND concurrency_level = 2;
                                          GROUP  BY launch_plan_named_entity_id); 
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
We could as an alternative, repurpose the existing launch plan index to include (launch plan id, phase, created at) to optimize the query for pending executions and not significantly affect queries on launch plan id leveraging the existing index.


## 6 Alternatives

### Scheduling
This proposal purposefully uses FIFO scheduling. But this does not preclude defining other scheduling orders or catch-up policies in the future.

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

Furthermore, we may want to introduce a max pending period to fail executions that have been in `PENDING` for too long

### Other concurrency policies: Terminate priors on execution

What if we actually want to terminate existing executions when the concurrency limit is reached?

In practice this could work by adding a new `ConcurrencyPolicy` enum for `RUN_IMMEDIATELY`

And the reconciliation loop would now proceed like so

In a separate goroutine, fetch items from the workqueue and individually process each execution entry
1. Check the database to see if there are fewer than `MAX_CONCURRENCY` non-terminal executions matching the launch plan ID in the pending execution model
   ```sql
   select count(launch_plan_id) from executions where phase not in ('SUCCEEDED', 'FAILED', 'ABORTED', 'TIMED_OUT') group by launch_plan_id;
   ```
1. If there are fewer than `MAX_CONCURRENCY` executions running
   1. check that the execution is still in `PENDING`
   1. create the workflow CRD
      1. if the CRD already exists because we've previously processed this pending execution before we had a chance to update the DB state, swallow the already exists error gracefully
   1. conditionally mark the execution as `QUEUED` in the db if it's already in `PENDING` or  `QUEUED` to not overwrite any events should the execution have already reported progress in the interim
   1. If creating the workflow CRD fails: mark the execution as `FAILED` in the db (and swallow any errors if the workflow is already in a terminal state should we have previously reported the failure after re-enqueuing the pending execution a previous loop). This will remove its eligiblity from the pending loop
1. If there are already `MAX_CONCURRENCY` executions running
   1. Retrieve n executions where n = count(actively running executions) - MAX_CONCURRENCY (ordered by creation time, ascending so we kill the oldest executions first)
   2. Kill each execution
   3. Proceed to (1) above.
1. Finally, always mark the queue item as [Done()](https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go#L33)


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
