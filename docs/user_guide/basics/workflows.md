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

(workflow)=

# Workflows

```{eval-rst}
.. tags:: Basic
```

Workflows link multiple tasks together. They can be written as Python functions,
but it's important to distinguish tasks and workflows.

A task's body executes at run-time on a Kubernetes cluster, in a Query Engine like BigQuery,
or on hosted services like AWS Batch or Sagemaker.

In contrast, a workflow's body doesn't perform computations; it's used to structure tasks.
A workflow's body executes at registration time, during the workflow's registration process.
Registration involves uploading the packaged (serialized) code to the Flyte backend,
enabling the workflow to be triggered.

For more information, see the {std:ref}`registration documentation <flyte:divedeep-registration>`.

To begin, import {py:func}`~flytekit.task` and {py:func}`~flytekit.workflow` from the flytekit library.

```{code-cell}
from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

We define `slope` and `intercept` tasks to compute the slope and
intercept of the regression line, respectively.

```{code-cell}
@task
def slope(x: list[int], y: list[int]) -> float:
    sum_xy = sum([x[i] * y[i] for i in range(len(x))])
    sum_x_squared = sum([x[i] ** 2 for i in range(len(x))])
    n = len(x)
    return (n * sum_xy - sum(x) * sum(y)) / (n * sum_x_squared - sum(x) ** 2)


@task
def intercept(x: list[int], y: list[int], slope: float) -> float:
    mean_x = sum(x) / len(x)
    mean_y = sum(y) / len(y)
    intercept = mean_y - slope * mean_x
    return intercept
```

+++ {"lines_to_next_cell": 0}

Define a workflow to establish the task dependencies.
Just like a task, a workflow is also strongly typed.

```{code-cell}
@workflow
def simple_wf(x: list[int], y: list[int]) -> float:
    slope_value = slope(x=x, y=y)
    intercept_value = intercept(x=x, y=y, slope=slope_value)
    return intercept_value
```

+++ {"lines_to_next_cell": 0}

The {py:func}`~flytekit.workflow` decorator encapsulates Flyte tasks,
essentially representing lazily evaluated promises.
During parsing, function calls are deferred until execution time.
These function calls generate {py:class}`~flytekit.extend.Promise`s that can be propagated to downstream functions,
yet remain inaccessible within the workflow itself.
The actual evaluation occurs when the workflow is executed.

Workflows can be executed locally, resulting in immediate evaluation, or through tools like
[`pyflyte`](https://docs.flyte.org/projects/flytekit/en/latest/pyflyte.html),
[`flytectl`](https://docs.flyte.org/projects/flytectl/en/latest/index.html) or UI, triggering evaluation.
While workflows decorated with `@workflow` resemble Python functions,
they function as python-esque Domain Specific Language (DSL).
When encountering a @task-decorated Python function, a promise object is created.
This promise doesn't store the task's actual output. Its fulfillment only occurs during execution.
Additionally, the inputs to a workflow are also promises, you can only pass promises into
tasks, workflows and other Flyte constructs.

:::{note}
You can learn more about creating dynamic Flyte workflows by referring
to {ref}`dynamic workflows <dynamic_workflow>`.
In a dynamic workflow, unlike a simple workflow, the inputs are pre-materialized.
However, each task invocation within the dynamic workflow still generates a promise that is evaluated lazily.
Bear in mind that a workflow can have tasks, other workflows and dynamic workflows.
:::

You can run a workflow by calling it as you would with a Python function and providing the necessary inputs.

```{code-cell}
if __name__ == "__main__":
    print(f"Running simple_wf() {simple_wf(x=[-3, 0, 3], y=[7, 4, -2])}")
```

+++ {"lines_to_next_cell": 0}

To run the workflow locally, you can use the following `pyflyte run` command:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/basics/basics/workflow.py \
  simple_wf --x '[-3,0,3]' --y '[7,4,-2]'
```

If you want to run it remotely on the Flyte cluster,
simply add the `--remote flag` to the `pyflyte run` command:
```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/basics/basics/workflow.py \
  simple_wf --x '[-3,0,3]' --y '[7,4,-2]'
```

While workflows are usually constructed from multiple tasks with dependencies established through
shared inputs and outputs, there are scenarios where isolating the execution of a single task
proves advantageous during the development and iteration of its logic.
Crafting a new workflow definition each time for this purpose can be cumbersome.
However, {ref}`executing an individual task <single_task_execution>` independently,
without the confines of a workflow, offers a convenient approach for iterating on task logic effortlessly.

## Use `partial` to provide default arguments to tasks
You can use the {py:func}`functools.partial` function to assign default or constant values to the parameters of your tasks.

```{code-cell}
import functools


@workflow
def simple_wf_with_partial(x: list[int], y: list[int]) -> float:
    partial_task = functools.partial(slope, x=x)
    return partial_task(y=y)
```
