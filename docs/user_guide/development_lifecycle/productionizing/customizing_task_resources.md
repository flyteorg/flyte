# Customizing task resources

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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 1-3
```

Define a task and configure the resources to be allocated to it:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: count_unique_numbers
```

Define a task that computes the square of a number:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: square
```

You can use the tasks decorated with memory and storage hints like regular tasks in a workflow.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: my_workflow
```

You can execute the workflow locally.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:lines: 38-40
```

Define a task and configure the resources to be allocated to it.
You can use tasks decorated with memory and storage hints like regular tasks in a workflow.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: count_unique_numbers
```

Define a task that computes the square of a number:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: square_1
```

The `with_overrides` method overrides the old resource allocations:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
:caption: productionizing/customizing_resources.py
:pyobject: my_pipeline
```

You can execute the workflow locally:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/customizing_resources.py
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
