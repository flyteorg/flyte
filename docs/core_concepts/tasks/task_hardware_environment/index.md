# Task hardware environment

You can customize the hardware environment in which your task code executes.

Depending on your needs, there are two different of ways to define and register tasks with their own custom hardware requirements:

* Configuration in the `@task` decorator
* Defining a PodTemplate

## Using the `@task` decorator

Control request and limits on

* CPU number
* GPU number
* Memory size
* Storage size
* Ephemeral storage size

See [Customizing task resources](customizing-task-resources) for details.

## Using PodTemplate

If your needs are more complex, you can use Kubernetes-level configuration to constrain a task to only run on a specific machine type.

This requires that you coordinate with Union to set up the required machine types and node groups with the appropriate node assignment configuration (node selector labels, node affinities, taints, tolerations, etc.)

In your task definition you then use a `PodTemplate` that that uses the matching node assignment configuration to make sure that the task will only be scheduled on the appropriate machine type.

### `pod_template` and `pod_template_name` @task parameters

The `pod_template` parameter can be used to supply a custom Kubernetes `PodTemplate` to the task.
This can be used to define details about node selectors, affinity, tolerations, and other Kubernetes-specific settings.

The `pod_template_name` is a related parameter that can be used to specify the name of an already existing `PodTemplate` resource which will be used in this task.

For details see [Configuring task pods with K8s PodTemplates&#x2B00;](https://docs.flyte.org/en/latest/deployment/configuration/general.html#deployment-configuration-general).

```{toctree}
:maxdepth: 2
:hidden:

customizing_task_resources
accelerators
interruptible_instances
configuring_access_to_gpus
```