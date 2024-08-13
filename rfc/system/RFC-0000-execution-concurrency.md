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
1. Query all pending executions by timestamp ascending (open question, should we prefer more recent executions instead? should we make this configurable?)
   1. as an optimization, could even parallelize this into goroutines, one per launch plan distinct launch plan ID that has any `PENDING` execution
2. Check the database to see if there are fewer than `MAX_CONCURRENCY` non-terminal executions with an identical launch plan ID
3. If there are none, select the oldest pending execution for that launch plan
   1. create the workflow CRD
   1. open question: also update its phase in the database to `QUEUED`?
   1. let execution proceed

We should consider adding an index to the executions table to include
- launch_plan_id
- phase
- created_at

#### Open Questions
- Should we always attempt to schedule pending executions in ascending order of creation time?
- Should we propagate concurrency policies to child executions?

## 4 Metrics & Dashboards

*What are the main metrics we should be measuring? For example, when interacting with an external system, it might be the external system latency. When adding a new table, how fast would it fill up?*

## 5 Drawbacks

*Are there any reasons why we should not do this? Here we aim to evaluate risk and check ourselves.*

## 6 Alternatives

*What are other ways of achieving the same outcome?*

## 7 Potential Impact and Dependencies

*Here, we aim to be mindful of our environment and generate empathy towards others who may be impacted by our decisions.*

- *What other systems or teams are affected by this proposal?*
- *How could this be exploited by malicious attackers?*

## 8 Unresolved questions

*What parts of the proposal are still being defined or not covered by this proposal?*

## 9 Conclusion

*Here, we briefly outline why this is the right decision to make at this time and move forward!*

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [ ]  Copy template
- [ ]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
