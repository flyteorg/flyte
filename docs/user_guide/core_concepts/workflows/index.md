(workflow)=

# Workflows

```{eval-rst}
.. tags:: Basic
```

Workflows link multiple tasks together. They can be written as Python functions,
but it's important to distinguish tasks and workflows.

A task's body executes at run-time on a Kubernetes cluster, in a Query Engine like BigQuery,
or on hosted services like AWS Batch or SageMaker.

In contrast, a workflow's body doesn't perform computations; it's used to structure tasks.
A workflow's body executes at registration time, during the workflow's registration process.
Registration involves uploading the packaged (serialized) code to the Flyte backend,
enabling the workflow to be triggered.

For more information, see the {std:ref}`registration documentation <flyte:divedeep-registration>`.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import {py:func}`~flytekit.task` and {py:func}`~flytekit.workflow` from the flytekit library:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py
:caption: basics/workflow.py
:lines: 1
```

We define `slope` and `intercept` tasks to compute the slope and
intercept of the regression line, respectively:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py
:caption: basics/workflow.py
:lines: 6-19
```

Define a workflow to establish the task dependencies.
Just like a task, a workflow is also strongly typed:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py
:caption: basics/workflow.py
:pyobject: simple_wf
```

The {py:func}`~flytekit.workflow` decorator encapsulates Flyte tasks, essentially representing lazily evaluated promises. During parsing, function calls are deferred until execution time. These function calls generate {py:class}`~flytekit.extend.Promise`s that can be propagated to downstream functions, yet remain inaccessible within the workflow itself. The actual evaluation occurs when the workflow is executed.

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

You can run a workflow by calling it as you would with a Python function and providing the necessary inputs:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py
:caption: basics/workflow.py
:lines: 33-34
```

To run the workflow locally, you can use the following `pyflyte run` command:

```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py \
  simple_wf --x '[-3,0,3]' --y '[7,4,-2]'
```

If you want to run it remotely on the Flyte cluster,
simply add the `--remote flag` to the `pyflyte run` command:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py \
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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/workflow.py
:caption: basics/workflow.py
:lines: 39-45
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
