# Customizing task resources

When defining a task function, you can specify resource requirements for the pod that runs the task.
Union will take this into account to ensure that the task pod is scheduled to run on a Kubernetes node that meets the specified resource profile.

Resources are specified in the `@task` decorator. Here is an example:

```{code-block} python
from flytekit.extras.accelerators import A100

@task(
    requests=Resources(mem="120Gi", cpu="44", gpu="8", storage="100Gi", ephemeral_storage="100Gi"),
    limits=Resources(mem="200Gi", cpu="100", gpu="12", storage="200Gi", ephemeral_storage="200Gi"),
    accelerator=A100
)
def my_task()
    ...
```

There are three separate resource-related settings:

* `requests`
* `limits`
* `accelerator`

## The `requests` and `limits` settings

The `requests` and `limits` settings each takes a [`Resource`](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.Resources.html#flytekit-resources) object, which itself has five possible attributes:

* `cpu`: Number of CPU cores (in whole numbers or millicores (`m`)).
* `gpu`: Number of GPU cores (in whole numbers or millicores (`m`)).
* `mem`: Main memory (in `Mi`, `Gi`, etc.).
* `storage`: Storage (in `Mi`,  `Gi` etc.).
* `ephemeral_storage`: Ephemeral storage (in `Mi`,  `Gi` etc.).

Note that CPU and GPU allocations can be specified either as whole numbers or in millicores (`m`). For example `cpu="2"` means 2 CPU cores and `gpu="3500m"`, meaning three and a half GPU cores.

The `requests` setting tells the system that the task requires _at least_ the resources specified and therefore the pod running this task should be scheduled only on a node that meets or exceeds the resource profile specified.

The `limits` setting serves as a hard upper bound on the resource profile of nodes to be scheduled to run the task.
The task will not be scheduled on a node that exceeds the resource profile specified (in any of the specified attributes).

:::{admonition} GPUs take only `limits`

GPUs should only be specified in the `limits` section of the task decorator:

* You should specify GPU requirements only in `limits`, not in `requests`, because Kubernetes will use the `limits` value as the `requests` value anyway.

* You *can* specify GPU in both `limits` and `requests` but the two values must be equal.

* You cannot specify GPU `requests` without specifying `limits`.

:::

## The `accelerator` setting

The `accelerator` setting further specifies the specifies the *type* of specialized hardware required for the task.
This may be a GPU, a specific variation of a GPU, a fractional GPU, or a different hardware device, such as a TPU.

See [Accelerators](accelerators) for more information.

## Task resource validation

If you attempt to execute a workflow with unsatisfiable resource requests, the execution will fail immediately rather than being allowed to queue forever.

To remedy such a failure, you should make sure that the appropriate node types are:
*  Physically available in your cluster, meaning you have arranged with the Union team to include them when [configuring your data plane](../../../data-plane-setup/configuring-your-data-plane).
* Specified in the task decorator (via the `requests`, `limits`, `accelerator`, or other parameters).

Go to the **Usage > Compute** dashboard to find the available node types and their resource profiles.
To make changes to your cluster configuration, go to the [Union Support Portal](https://get.support.union.ai/servicedesk/customer/portal/1/group/6/create/30).
This portal also accessible from **Usage > Compute** through the **Adjust Configuration** button:

![]/_static/images/adjust-configuration.png)

See also [Customizing Task Resources](https://docs.flyte.org/en/latest/deployment/configuration/customizable_resources.html#task-resources) in the Flyte OSS docs.

## The `with_overrides` method

When `requests`, `limits`, or `accelerator` are specified in the `@task` decorator, they apply every time that a task is invoked from a workflow.
In some cases, you may wish to change the resources specified from one invocation to another.
To do that, use the [`with_overrides` method](https://docs.flyte.org/en/latest/flytesnacks/examples/productionizing/customizing_resources.html#resource-with-overrides) of the task function.
For example:

```{code-block} python
@task
def my_task(ff: FlyteFile):
    ...

@workflow
def my_workflow():
    my_task(ff=smallFile)
    my_task(ff=bigFile).withoverrides(requests=Resources(mem="120Gi", cpu="10"))
```

# Customizing task resources (doc from user guide)

```{eval-rst}
.. tags:: Deployment, Infrastructure, Basic
```

One of the reasons to use a hosted Flyte environment is the potential of leveraging CPU, memory and storage resources, far greater than what's available locally.
Flytekit makes it possible to specify these requirements declaratively and close to where the task itself is declared.

In this example, the memory required by the function increases as the dataset size increases.
Large datasets may not be able to run locally, so we would want to provide hints to the Flyte backend to request for more memory.
This is done by decorating the task with the hints as shown in the following code sample.

Tasks can have `requests` and `limits` which mirror the native [equivalents in Kubernetes](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits).
A task can possibly be allocated more resources than it requests, but never more than its limit.
Requests are treated as hints to schedule tasks on nodes with available resources, whereas limits
are hard constraints.

For either a request or limit, refer to the {py:class}`flytekit:flytekit.Resources` documentation.

The following attributes can be specified for a `Resource`.

1. `cpu`
2. `mem`
3. `gpu`

To ensure that regular tasks that don't require GPUs are not scheduled on GPU nodes, a separate node group for GPU nodes can be configured with [taints](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).

To ensure that tasks that require GPUs get the needed tolerations on their pods, set up FlytePropeller using the following [configuration](https://github.com/flyteorg/flytepropeller/blob/v0.10.5/config.yaml#L51,L56). Ensure that this toleration config matches the taint config you have configured to protect your GPU providing nodes from dealing with regular non-GPU workloads (pods).

The actual values follow the [Kubernetes convention](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes). Let's look at an example to understand how to customize resources.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

Import the dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 1-3
```

Define a task and configure the resources to be allocated to it:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: count_unique_numbers
```

Define a task that computes the square of a number:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: square
```

You can use the tasks decorated with memory and storage hints like regular tasks in a workflow.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: my_workflow
```

You can execute the workflow locally.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 32-34
```

:::{note}
To alter the limits of the default platform configuration, change the [admin config](https://github.com/flyteorg/flyte/blob/b16ffd76934d690068db1265ac9907a278fba2ee/deployment/eks/flyte_helm_generated.yaml#L203-L213) and [namespace level quota](https://github.com/flyteorg/flyte/blob/b16ffd76934d690068db1265ac9907a278fba2ee/deployment/eks/flyte_helm_generated.yaml#L214-L240) on the cluster.
:::

(resource_with_overrides)=

## Using `with_overrides`

You can use the `with_overrides` method to override the resources allocated to the tasks dynamically.
Let's understand how the resources can be initialized with an example.

Import the dependencies.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 38-40
```

Define a task and configure the resources to be allocated to it.
You can use tasks decorated with memory and storage hints like regular tasks in a workflow.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: count_unique_numbers
```

Define a task that computes the square of a number:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: square_1
```

The `with_overrides` method overrides the old resource allocations:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: my_pipeline
```

You can execute the workflow locally:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 65-67
```

You can see the memory allocation below. The memory limit is `500Mi` rather than `350Mi`, and the
CPU limit is 4, whereas it should have been 6 as specified using `with_overrides`.
This is because the default platform CPU quota for every pod is 4.

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/core/resource_allocation.png
:alt: Resource allocated using "with_overrides" method

Resource allocated using "with_overrides" method
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
