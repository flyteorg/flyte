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
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(map_task)=

# Map tasks

```{eval-rst}
.. tags:: Intermediate
```

Using a map task in Flyte allows for the execution of a pod task or a regular task across a series of inputs within a single workflow node.
This capability eliminates the need to create individual nodes for each instance, leading to substantial performance improvements.

Map tasks find utility in diverse scenarios, such as:

1. Executing the same code logic on multiple inputs
2. Concurrent processing of multiple data batches
3. Hyperparameter optimization

The following examples demonstrate how to use map tasks with both single and multiple inputs.

To begin, import the required libraries.

```{code-cell}
from flytekit import map_task, task, workflow
```

+++ {"lines_to_next_cell": 0}

Here's a simple workflow that uses {py:func}`map_task <flytekit:flytekit.map_task>`.

```{code-cell}
threshold = 11


@task
def detect_anomalies(data_point: int) -> bool:
    return data_point > threshold


@workflow
def map_workflow(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    # Use the map task to apply the anomaly detection function to each data point
    return map_task(detect_anomalies)(data_point=data)


if __name__ == "__main__":
    print(f"Anomalies Detected: {map_workflow()}")
```

+++ {"lines_to_next_cell": 0}

To customize resource allocations, such as memory usage for individual map tasks,
you can leverage `with_overrides`. Here's an example using the `detect_anomalies` map task within a workflow:

```python
from flytekit import Resources


@workflow
def map_workflow_with_resource_overrides(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    return map_task(detect_anomalies)(data_point=data).with_overrides(requests=Resources(mem="2Gi"))
```

You can use {py:class}`~flytekit.TaskMetadata` to set attributes such as `cache`, `cache_version`, `interruptible`, `retries` and `timeout`.
```python
from flytekit import TaskMetadata


@workflow
def map_workflow_with_metadata(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    return map_task(detect_anomalies, metadata=TaskMetadata(cache=True, cache_version="0.1", retries=1))(
        data_point=data
    )
```

You can also configure `concurrency` and `min_success_ratio` for a map task:
- `concurrency` limits the number of mapped tasks that can run in parallel to the specified batch size.
If the input size exceeds the concurrency value, multiple batches will run serially until all inputs are processed.
If left unspecified, it implies unbounded concurrency.
- `min_success_ratio` determines the minimum fraction of total jobs that must complete successfully before terminating
the map task and marking it as successful.

```python
@workflow
def map_workflow_with_additional_params(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    return map_task(detect_anomalies, concurrency=1, min_success_ratio=0.75)(data_point=data)
```

A map task internally uses a compression algorithm (bitsets) to handle every Flyte workflow node’s metadata,
which would have otherwise been in the order of 100s of bytes.

When defining a map task, avoid calling other tasks in it. Flyte
can't accurately register tasks that call other tasks. While Flyte
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

+++ {"lines_to_next_cell": 0}

By default, the map task utilizes the Kubernetes array plugin for execution.
However, map tasks can also be run on alternate execution backends.
For example, you can configure the map task to run on
[AWS Batch](https://docs.flyte.org/en/latest/deployment/plugin_setup/aws/batch.html#deployment-plugin-setup-aws-array),
a provisioned service that offers scalability for handling large-scale tasks.

## Map a task with multiple inputs

You might need to map a task with multiple inputs.

For instance, consider a task that requires three inputs.

```{code-cell}
@task
def multi_input_task(quantity: int, price: float, shipping: float) -> float:
    return quantity * price * shipping
```

+++ {"lines_to_next_cell": 0}

You may want to map this task with only the ``quantity`` input, while keeping the other inputs unchanged.
Since a map task accepts only one input, you can achieve this by partially binding values to the map task.
This can be done using the {py:func}`functools.partial` function.

```{code-cell}
import functools


@workflow
def multiple_inputs_map_workflow(list_q: list[int] = [1, 2, 3, 4, 5], p: float = 6.0, s: float = 7.0) -> list[float]:
    partial_task = functools.partial(multi_input_task, price=p, shipping=s)
    return map_task(partial_task)(quantity=list_q)
```

+++ {"lines_to_next_cell": 0}

Another possibility is to bind the outputs of a task to partials.

```{code-cell}
@task
def get_price() -> float:
    return 7.0


@workflow
def map_workflow_partial_with_task_output(list_q: list[int] = [1, 2, 3, 4, 5], s: float = 6.0) -> list[float]:
    p = get_price()
    partial_task = functools.partial(multi_input_task, price=p, shipping=s)
    return map_task(partial_task)(quantity=list_q)
```

+++ {"lines_to_next_cell": 0}

You can also provide multiple lists as input to a ``map_task``.

```{code-cell}
:lines_to_next_cell: 2

@workflow
def map_workflow_with_lists(
    list_q: list[int] = [1, 2, 3, 4, 5], list_p: list[float] = [6.0, 9.0, 8.7, 6.5, 1.2], s: float = 6.0
) -> list[float]:
    partial_task = functools.partial(multi_input_task, shipping=s)
    return map_task(partial_task)(quantity=list_q, price=list_p)
```

```{note}
It is important to note that you cannot provide a list as an input to a partial task.
```

## Run the example on the Flyte cluster

To run the provided workflows on the Flyte cluster, use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_with_additional_params
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/map_task.py \
  multiple_inputs_map_workflow
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_partial_with_task_output
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_with_lists
```

## ArrayNode

:::{important}
This feature is experimental and the API is subject to breaking changes.
If you encounter any issues please consider submitting a
[bug report](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2Cuntriaged&projects=&template=bug_report.yaml&title=%5BBUG%5D+).
:::

ArrayNode map tasks serve as a seamless substitution for regular map tasks, differing solely in the submodule
utilized to import the `map_task` function. Specifically, you will need to import `map_task` from the experimental module as illustrated below:

```python
from flytekit import task, workflow
from flytekit.experimental import map_task

@task
def t(a: int) -> int:
    ...

@workflow
def array_node_wf(xs: list[int]) -> list[int]:
    return map_task(t)(a=xs)
```

Flyte introduces map task to enable parallelization of homogeneous operations,
offering efficient evaluation and a user-friendly API. Because it’s implemented as a backend plugin,
its evaluation is independent of core Flyte logic, which generates subtask executions that lack full Flyte functionality.
ArrayNode tackles this issue by offering robust support for subtask executions.
It also extends mapping capabilities across all plugins and Flyte node types.
This enhancement will be a part of our move from the experimental phase to general availability.

In contrast to map tasks, an ArrayNode provides the following enhancements:

- **Wider mapping support**. ArrayNode extends mapping capabilities beyond Kubernetes tasks, encompassing tasks such as Python tasks, container tasks and pod tasks.
- **Cache management**. It supports both cache serialization and cache overwriting for subtask executions.
- **Intra-task checkpointing**. ArrayNode enables intra-task checkpointing, contributing to improved execution reliability.
- **Workflow recovery**. Subtasks remain recoverable during the workflow recovery process. (This is a work in progress.)
- **Subtask failure handling**. The mechanism handles subtask failures effectively, ensuring that running subtasks are appropriately aborted.
- **Multiple input values**. Subtasks can be defined with multiple input values, enhancing their versatility.

We expect the performance of ArrayNode map tasks to compare closely to standard map tasks.
