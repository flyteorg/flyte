(concurrency_control)=

# Concurrency Control

```{eval-rst}
.. tags:: Intermediate
```

Concurrency control allows you to limit the number of concurrently running workflow executions for a specific launch plan, identified by its unique `project`, `domain`, and `name`. This control is applied across all versions of that launch plan.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

## How It Works

When a new execution for a launch plan with a `ConcurrencyPolicy` is requested, flyteadmin will perform a check to count the number of currently active executions for that same launch plan (`project/domain/name`), irrespective of their versions.

This check is done using a database query that joins the `executions` table with the `launch_plans` table. It filters for executions that are in an active phase (e.g., `QUEUED`, `RUNNING`, `ABORTING`, etc.) and belong to the launch plan name being triggered.

If the number of active executions is already at or above the `max_concurrency` limit defined in the policy of the launch plan version being triggered, the new execution will be handled according to the specified `behavior`.

## Basic Usage

Here's an example of how to define a launch plan with concurrency control:

```python
from flytekit import ConcurrencyPolicy, ConcurrencyLimitBehavior, LaunchPlan, workflow

@workflow
def my_workflow() -> str:
    return "Hello, World!"

# Create a launch plan with concurrency control
concurrency_limited_lp = LaunchPlan.get_or_create(
    name="my_concurrent_lp",
    workflow=my_workflow,
    concurrency=ConcurrencyPolicy(
        max_concurrency=3,
        behavior=ConcurrencyLimitBehavior.SKIP,
    ),
)
```

## Scheduled Workflows with Concurrency Control

Concurrency control is particularly useful for scheduled workflows to prevent overlapping executions:

```python
from flytekit import ConcurrencyPolicy, ConcurrencyLimitBehavior, CronSchedule, LaunchPlan, workflow

@workflow
def scheduled_workflow() -> str:
    # This workflow might take a long time to complete
    return "Processing complete"

# Create a scheduled launch plan with concurrency control
scheduled_lp = LaunchPlan.get_or_create(
    name="my_scheduled_concurrent_lp",
    workflow=scheduled_workflow,
    concurrency=ConcurrencyPolicy(
        max_concurrency=1,  # Only allow one execution at a time
        behavior=ConcurrencyLimitBehavior.SKIP,
    ),
    schedule=CronSchedule(schedule="*/5 * * * *"),  # Runs every 5 minutes
)
```

## Defining the Policy

A `ConcurrencyPolicy` is defined with two main parameters:

- `max_concurrency` (integer): The maximum number of workflows that can be running concurrently for this launch plan name.
- `behavior` (enum): What to do when the `max_concurrency` limit is reached. Currently, only `SKIP` is supported, which means new executions will not be created if the limit is hit.

```python
from flytekit import ConcurrencyPolicy, ConcurrencyLimitBehavior

policy = ConcurrencyPolicy(
    max_concurrency=5,
    behavior=ConcurrencyLimitBehavior.SKIP
)
```

## Key Behaviors & Considerations

### Version-Agnostic Check, Version-Specific Enforcement

The concurrency check counts all active workflow executions of a given launch plan (`project/domain/name`). However, the enforcement (i.e., the `max_concurrency` limit and `behavior`) is based on the `ConcurrencyPolicy` defined in the specific version of the launch plan you are trying to launch.

**Example Scenario:**
1. Launch Plan `MyLP` version `v1` has a `ConcurrencyPolicy` with `max_concurrency = 3`.
2. Three executions of `MyLP` (they could be `v1` or any other version) are currently running.
3. You try to launch `MyLP` version `v2`, which has a `ConcurrencyPolicy` with `max_concurrency = 10`.
   - **Result**: This `v2` execution will launch successfully because its own limit (10) is not breached by the current 3 active executions.
4. Now, with 4 total active executions (3 original + the new `v2`), you try to launch `MyLP` version `v1` again.
   - **Result**: This `v1` execution will **fail**. The check sees 4 active executions, and `v1`'s policy only allows a maximum of 3.

### Manual Execution Example

When running workflows manually, you might encounter concurrency limits:

```bash
# If the concurrency limit is hit, you might see an error like:
# _InactiveRpcError: <_InactiveRpcError of RPC that terminated with:
#         status = StatusCode.RESOURCE_EXHAUSTED
#         details = "Concurrency limit (1) reached for launch plan my_workflow_lp. Skipping execution."
# >
```

### Scheduled Execution Behavior

When the scheduler attempts to trigger an execution and the concurrency limit is met, the creation will fail, and errors will be logged in FlyteAdmin. The execution will be skipped according to the `SKIP` behavior.

## Limitations

### "At Most" Enforcement

While the system aims to respect `max_concurrency`, it acts as an "at most" limit. Due to the nature of scheduling, workflow execution durations, and the timing of the concurrency check (at launch time), there might be periods where the number of active executions is below `max_concurrency` even if the system could theoretically run more.

For example, if `max_concurrency` is 5 and all 5 workflows finish before the next scheduled check/trigger, the count will drop. The system prevents exceeding the limit but doesn't actively try to always maintain `max_concurrency` running instances.

### Notifications for Skipped Executions

Currently, there is no built-in notification system for skipped executions. When a scheduled execution is skipped due to concurrency limits, it will be logged in FlyteAdmin but no user notification will be sent. This is an area for future enhancement.

## Best Practices

1. **Use with Scheduled Workflows**: Concurrency control is most beneficial for scheduled workflows that might take longer than the schedule interval to complete.

2. **Set Appropriate Limits**: Consider your system resources and the resource requirements of your workflows when setting `max_concurrency`.

3. **Monitor Skipped Executions**: Regularly check FlyteAdmin logs to monitor if executions are being skipped due to concurrency limits.

4. **Version Management**: Be aware that different versions of the same launch plan can have different concurrency policies, but the check is performed across all versions.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
