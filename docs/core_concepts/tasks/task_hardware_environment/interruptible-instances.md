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
