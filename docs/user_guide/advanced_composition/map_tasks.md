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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:lines: 1
```

Here's a simple workflow that uses {py:func}`map_task <flytekit:flytekit.map_task>`:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:lines: 4-19
```

To customize resource allocations, such as memory usage for individual map tasks,
you can leverage `with_overrides`. Here's an example using the `detect_anomalies` map task within a workflow:

```python
from flytekit import Resources


@workflow
def map_workflow_with_resource_overrides(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    return map_task(detect_anomalies)(data_point=data).with_overrides(requests=Resources(mem="2Gi"))
```

You can use {py:class}`~flytekit.TaskMetadata` to set attributes such as `cache`, `cache_version`, `interruptible`, `retries` and `timeout`:
```python
from flytekit import TaskMetadata


@workflow
def map_workflow_with_metadata(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[bool]:
    return map_task(detect_anomalies, metadata=TaskMetadata(cache=True, cache_version="0.1", retries=1))(
        data_point=data
    )
```
When `cache` and `cache_version` are used in `TaskMetadata` for a map task, the cache hits occur on individual tasks being mapped over, rather than the parent map task operation. This means that if one input item in a list changes, each previously executed task is read from cache and only the task for the changed item is actually executed, rather than the task being re-executed for every item.  Note that this has the same effect as adding `cache` and `cache_version` in the `@task` decorator for a task being mapped over.

You can also configure `concurrency` and `min_success_ratio` for a map task:
- `concurrency` limits the number of mapped tasks that can run in parallel to the specified batch size.
If the input size exceeds the concurrency value, multiple batches will run serially until all inputs are processed. If left unspecified, it implies unbounded concurrency.
- `min_success_ratio` determines the minimum fraction of total jobs that must complete successfully before terminating the map task and marking it as successful.

```python
@workflow
def map_workflow_with_additional_params(data: list[int] = [10, 12, 11, 10, 13, 12, 100, 11, 12, 10]) -> list[typing.Optional[bool]]:
    return map_task(detect_anomalies, concurrency=1, min_success_ratio=0.75)(data_point=data)
```

:::{note}
Notice the return type of the list has been set to `Optional` when a `min_success_ratio` is added. This is due to the fact we are now tolerating failures, meaning the expected return type from the mapped task may in fact not get returned.
:::

A map task internally uses a compression algorithm (bitsets) to handle every Flyte workflow node’s metadata,
which would have otherwise been in the order of 100s of bytes.

When defining a map task, avoid calling other tasks in it. Flyte can't accurately register tasks that call other tasks. While Flyte will correctly execute a task that calls other tasks, it will not be able to give full performance advantages. This is especially true for map tasks.

In this example, the map task `suboptimal_mappable_task` would not give you the best performance:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:lines: 31-40
```

By default, the map task utilizes the Kubernetes array plugin for execution.
However, map tasks can also be run on alternate execution backends.
For example, you can configure the map task to run on
[AWS Batch](https://docs.flyte.org/en/latest/deployment/plugin_setup/aws/batch.html#deployment-plugin-setup-aws-array), a provisioned service that offers scalability for handling large-scale tasks.

## Map a task with multiple inputs

You might need to map a task with multiple inputs.

For instance, consider a task that requires three inputs:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:pyobject: multi_input_task
```

You may want to map this task with only the ``quantity`` input, while keeping the other inputs unchanged.
Since a map task accepts only one input, you can achieve this by partially binding values to the map task.
This can be done using the {py:func}`functools.partial` function:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:lines: 52-58
```

Another possibility is to bind the outputs of a task to partials:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:lines: 63-72
```

You can also provide multiple lists as input to a `map_task`:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py
:caption: advanced_composition/map_task.py
:pyobject: map_workflow_with_lists
```

```{note}
It is important to note that you cannot provide a list as an input to a partial task.
```

## Run the example on the Flyte cluster

To run the provided workflows on the Flyte cluster, use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_with_additional_params
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py \
  multiple_inputs_map_workflow
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_partial_with_task_output
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/map_task.py \
  map_workflow_with_lists
```

## ArrayNode

Flyte originally introduced map tasks to enable parallelization of homogeneous operations,
offering efficient evaluation and a user-friendly API. Because it’s implemented as a backend plugin,
its evaluation is independent of core Flyte logic, which generates subtask executions that lack full Flyte functionality.
ArrayNode tackled this issue by offering robust support for subtask executions.
It also extends mapping capabilities across all plugins and Flyte node types.
Starting with `flytekit` version 1.12.0, ArrayNode is the default `map_task` importable via `from flytekit import map_task`.

In contrast to the original map tasks, an ArrayNode provides the following enhancements:

- **Wider mapping support**. ArrayNode extends mapping capabilities beyond Kubernetes tasks, encompassing tasks such as Python tasks, container tasks and pod tasks.
- **Cache management**. It supports both cache serialization and cache overwriting for subtask executions.
- **Intra-task checkpointing**. ArrayNode enables intra-task checkpointing, contributing to improved execution reliability.
- **Workflow recovery**. Subtasks remain recoverable during the workflow recovery process. (This is a work in progress.)
- **Subtask failure handling**. The mechanism handles subtask failures effectively, ensuring that running subtasks are appropriately aborted.
- **Multiple input values**. Subtasks can be defined with multiple input values, enhancing their versatility.

We expect the performance of ArrayNode map tasks to compare closely to standard map tasks.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
