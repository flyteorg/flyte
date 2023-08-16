---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Map Tasks

```{tags} Intermediate
```

```{image} https://img.shields.io/badge/Blog%20Post-Map%20Tasks-blue?style=for-the-badge
:alt: Map Task Blog Post
:target: https://blog.flyte.org/map-tasks-in-flyte
```

A map task allows you to execute a pod task or a regular task on a series of inputs within a single workflow node.
This enables you to execute numerous instances of the task without having to create a node for each instance, resulting in significant performance improvements.

Map tasks find application in various scenarios, including:

- When multiple inputs require running through the same code logic.
- Processing multiple data batches concurrently.
- Conducting hyperparameter optimization.

Now, let's delve into an example!

+++ {"lines_to_next_cell": 0}

First, import the libraries.

```{code-cell}
from typing import List

from flytekit import Resources, map_task, task, workflow
```

+++ {"lines_to_next_cell": 0}

Define a task to be used in the map task.

:::{note}
A map task can only accept one input and produce one output.
:::

```{code-cell}
@task
def a_mappable_task(a: int) -> str:
    inc = a + 2
    stringified = str(inc)
    return stringified
```

+++ {"lines_to_next_cell": 0}

Also define a task to reduce the mapped output to a string.

```{code-cell}
@task
def coalesce(b: List[str]) -> str:
    coalesced = "".join(b)
    return coalesced
```

To repeat the execution of the `a_mappable_task` across a collection of inputs, use the {py:func}`~flytekit:flytekit.map_task` function from flytekit.
In this example, the input `a` is of type `List[int]`.
The `a_mappable_task` is executed for each element in the list.

You can utilize the `with_overrides` function to set resources specifically for individual map tasks.
This allows you to customize resource allocations such as memory usage.

```{code-cell}
@workflow
def my_map_workflow(a: List[int]) -> str:
    mapped_out = map_task(a_mappable_task)(a=a).with_overrides(
        requests=Resources(mem="300Mi"),
        limits=Resources(mem="500Mi"),
        retries=1,
    )
    coalesced = coalesce(b=mapped_out)
    return coalesced
```

+++ {"lines_to_next_cell": 0}

Finally, you can run the workflow locally.

```{code-cell}
:lines_to_next_cell: 1

if __name__ == "__main__":
    result = my_map_workflow(a=[1, 2, 3, 4, 5])
    print(f"{result}")
```

+++ {"lines_to_next_cell": 0}

When defining a map task, avoid calling other tasks in it. Flyte
can't accurately register tasks that call other tasks.  While Flyte
will correctly execute a task that calls other tasks, it will not be
able to give full performance advantages. This is
especially true for map tasks.

In this example, the map task `suboptimal_mappable_task` would not
give you the best performance.

```{code-cell}
@task
def upperhalf(a: int) -> int:
    return a / 2 + 1


@task
def suboptimal_mappable_task(a: int) -> str:
    inc = upperhalf(a=a)
    stringified = str(inc)
    return stringified
```

By default, the map task utilizes the Kubernetes Array plugin for execution.
However, map tasks can also be run on alternate execution backends.
For example, you can configure the map task to run on [AWS Batch](https://docs.flyte.org/en/latest/deployment/plugin_setup/aws/batch.html#deployment-plugin-setup-aws-array),
a provisioned service that offers scalability for handling large-scale tasks.

+++

## Map a Task with Multiple Inputs

You might need to map a task with multiple inputs.

For instance, consider a task that requires three inputs.

```{code-cell}
@task
def multi_input_task(quantity: int, price: float, shipping: float) -> float:
    return quantity * price * shipping
```

In some cases, you may want to map this task with only the ``quantity`` input, while keeping the other inputs unchanged.
Since a map task accepts only one input, you can achieve this by partially binding values to the map task.
This can be done using the {py:func}`functools.partial` function.

```{code-cell}
import functools


@workflow
def multiple_workflow(list_q: List[int] = [1, 2, 3, 4, 5], p: float = 6.0, s: float = 7.0) -> List[float]:
    partial_task = functools.partial(multi_input_task, price=p, shipping=s)
    return map_task(partial_task)(quantity=list_q)
```

Another possibility is to bind the outputs of a task to partials.

```{code-cell}
@task
def get_price() -> float:
    return 7.0


@workflow
def multiple_workflow_with_task_output(list_q: List[int] = [1, 2, 3, 4, 5], s: float = 6.0) -> List[float]:
    p = get_price()
    partial_task = functools.partial(multi_input_task, price=p, shipping=s)
    return map_task(partial_task)(quantity=list_q)
```

You can also provide multiple lists as input to a ``map_task``.

```{code-cell}
:lines_to_next_cell: 2

@workflow
def multiple_workflow_with_lists(
    list_q: List[int] = [1, 2, 3, 4, 5], list_p: List[float] = [6.0, 9.0, 8.7, 6.5, 1.2], s: float = 6.0
) -> List[float]:
    partial_task = functools.partial(multi_input_task, shipping=s)
    return map_task(partial_task)(quantity=list_q, price=list_p)
```

```{note}
It is important to note that you cannot provide a list as an input to a partial task.
```
