# Interruptible instances

:::{note}
In AWS the terms *interruptible instance* and  *spot instance* are used.
In GCP the equivalent term is *preemptible instance*.
Here we use the term *interruptible instance* generically for both providers.
:::

An interruptible instance is a machine instance made available to your cluster by your cloud provider that is not guaranteed to be always available.
As a result, interruptible instances are cheaper than regular instances.
In order to use an interruptible instance for a compute workload you have to be prepared for the possibility that an attempt to run the workload could fail due to lack of available resources and will need to be retried.

When onboarding your organization onto Union you [specify the configuration of your cluster](../../../data-plane-setup/configuring-your-data-plane).
Among the options is the choice of whether to use interruptible instances.

For each interruptible instance node group that you specify, an additional on-demand node group (though identical in every other respect to the interruptible one) will also be configured.
This on-demand node group will be used as a fallback when attempts to complete the task on the interruptible instance have failed.

## Configuring tasks to use interruptible instances

To schedule tasks on interruptible instances and retry them if they fail, specify the `interruptible` and `retries` parameters in the `@task` decorator.
For example:

```{code-block} python
@task(interruptible=True, retries=3)
```

* A task will only be scheduled on a interruptible instance if it has the parameter `interruptible=True` (or if its workflow has the parameter `interruptible=True` and the task does not have an explicit `interruptible` parameter).
* An interruptible task, like any other task, can have a `retries` parameter.
* If an interruptible task does not have an explicitly set `retries` parameter, then the `retries` value defaults to `1`.
* An interruptible task with `retries=n` will be attempted `n-1` times on a interruptible instance.
  If it still fails after `n-1` attempts the final retry will be done on the fallback on-demand instance.

## Workflow level interruptible


Interruptible is also available [at the workflow level](../../workflows/index). If you set it there it will apply to all tasks in the workflow that do not themselves have an explicit value set. Task level interruptible setting always overrides whatever the the workflow level setting is.

## Advantages and disadvantages of interruptible instances

The advantage of using interruptible instance for a task is simply that is less costly than using an on-demand instance (all other parameters being equal).
However, there are two main disadvantages:

1. The task is successfully scheduled on an interruptible instance but is interrupted.
In the worst case scenario, for `retries=n` the task may be interrupted `n-1` times until, finally, the fallback on-demand instance is used.
Clearly, this may be problem for time-critical tasks.

2. Interruptible instances of the selected node type may simply be unavailable on the initially attempt to schedule.
When this happens, the task may hang indefinitely until an interruptible instance becomes available.
Note that this is a distinct failure mode from the previous one where an interruptible node is successfully scheduled but is then interrupted.

In general, Union recommends that you use interruptible instances whenever available, but only for tasks that are not time-critical.


# Spot instances (copied from Flyte docs)

```{eval-rst}
.. tags:: AWS, GCP, Intermediate

```

## What are spot instances?

Spot instances are unused EC2 capacity in AWS. [Spot instances](https://aws.amazon.com/ec2/spot/?cards.sort-by=item.additionalFields.startDateTime&cards.sort-order=asc) can result in up to 90% savings on on-demand prices. The caveat is that these instances can be preempted at any point and no longer be available for use. This can happen due to:

- Price – The spot price is greater than your maximum price.
- Capacity – If there are not enough unused EC2 instances to meet the demand for spot instances, Amazon EC2 interrupts spot instances. Amazon EC2 determines the order in which the instances are interrupted.
- Constraints – If your request includes a constraint such as a launch group or an Availability Zone group, these spot instances are terminated as a group when the constraint can no longer be met.

Generally, most spot instances are obtained for around 2 hours (median), with the floor being about 20 minutes and the ceiling of unbounded duration.

:::{note}
Spot Instances are called `Preemptible Instances` in the GCP terminology.
:::

### Setting up spot instances

- AWS: <https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-requests.html>
- GCP: <https://cloud.google.com/compute/docs/instances/create-start-preemptible-instance>

If an auto-scaling group (ASG) is set up, you may want to isolate the tasks you want to trigger on spot/preemptible instances from the regular workloads.
This can be done by setting taints and tolerations using the [config](https://github.com/flyteorg/flyteplugins/blob/60b94c688ef2b98aa53a9224b529ac672af04540/go/tasks/pluginmachinery/flytek8s/config/config.go#L84-L92) available at `flyteorg/flyteplugins` repo.

:::{admonition} What's an ASG for a spot/preemptible instance?
When your spot/preemptible instance is terminated, ASG attempts to launch a replacement instance to maintain the desired capacity for the group.
:::


## What are interruptible tasks?

If specified, the `interruptible flag` is added to the task definition and signals to the Flyte engine that it may be scheduled on machines that may be preempted, such as AWS spot instances. This is low-hanging fruit for any cost-savings initiative.

### Setting interruptible tasks

To run your workload on a spot/preemptible instance, you can set interruptible to `True`. In case you would like to automatically retry in case the node gets preemted, please also make sure to set at least one retry. For example:

```python
@task(cache_version='1', interruptible=True, retries=1)
def add_one_and_print(value_to_print: int) -> int:
    return value_to_print + 1
```

By setting this value, Flyte will schedule your task on an auto-scaling group (ASG) with only spot instances.

:::{note}
If you set `retries=n`, for instance, and the task gets preempted repeatedly, Flyte will retry on a preemptible/spot instance `n-1` times and for the last attempt will retry your task on a non-spot (regular) instance. Please note that tasks will only be retried if at least one retry is allowed using the `retries` parameter in the `task` decorator.
:::

### Which tasks should be set to interruptible?

Most Flyte workloads should be good candidates for spot instances.
If your task does NOT exhibit the following properties, you can set `interruptible` to true.

- Time-sensitive: It needs to run now and can not have any unexpected delays.
- Side Effects: The task is not idempotent, and retrying will cause issues.
- Long-Running Tasks: The task takes > 2 hours. Having an interruption during this time frame could potentially waste a lot of computation.

In a nutshell, you should use spot/preemptible instances when you want to reduce the total cost of running jobs at the expense of potential delays in execution due to restarts.

% TODO: Write "How to Recover From Interruptions?" section

