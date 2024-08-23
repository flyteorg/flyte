(task)=

# Tasks

```{eval-rst}
.. tags:: Basic
```

A task serves as the fundamental building block and an extension point within Flyte.
It exhibits the following characteristics:

1. Versioned (typically aligned with the git sha)
2. Strong interfaces (annotated inputs and outputs)
3. Declarative
4. Independently executable
5. Suitable for unit testing

A Flyte task operates within its own container and runs on a [Kubernetes pod](https://kubernetes.io/docs/concepts/workloads/pods/).
It can be classified into two types:

1. A task associated with a Python function. Executing the task is the same as executing the function.
2. A task without a Python function, such as a SQL query or a portable task like prebuilt
   algorithms in SageMaker, or a service calling an API.

Flyte offers numerous plugins for tasks, including backend plugins like
[Athena](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-aws-athena/flytekitplugins/athena/task.py).

This example demonstrates how to write and execute a
[Python function task](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/python_function_task.py#L75).

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import `task` from the `flytekit` library:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/task.py
:caption: basics/task.py
:lines: 1
```

The use of the {py:func}`~flytekit.task` decorator is mandatory for a ``PythonFunctionTask``.
A task is essentially a regular Python function, with the exception that all inputs and outputs must be clearly annotated with their types.
Learn more about the supported types in the {ref}`type-system section <python_to_flyte_type_mapping>`.

We create a task that computes the slope of a regression line:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/task.py
:caption: basics/task.py
:pyobject: slope
```

:::{note}
Flytekit will assign a default name to the output variable like `out0`.
In case of multiple outputs, each output will be numbered in the order
starting with 0, e.g., -> `out0, out1, out2, ...`.
:::

You can execute a Flyte task just like any regular Python function:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/task.py
:caption: basics/task.py
:lines: 14-15
```

:::{note}
When invoking a Flyte task, you need to use keyword arguments to specify
the values for the corresponding parameters.
:::

(single_task_execution)=

To run it locally, you can use the following `pyflyte run` command:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/task.py \
  slope --x '[-3,0,3]' --y '[7,4,-2]'
```

If you want to run it remotely on the Flyte cluster,
simply add the `--remote flag` to the `pyflyte run` command:
```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/task.py \
  slope --x '[-3,0,3]' --y '[7,4,-2]'
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
