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

(decorating_workflows)=

# Decorating workflows

```{eval-rst}
.. tags:: Intermediate
```

The behavior of workflows can be modified in a light-weight fashion by using the built-in {py:func}`~functools.wraps`
decorator pattern, similar to using decorators to
{ref}`customize task behavior <decorating_tasks>`. However, unlike in the case of
tasks, we need to do a little extra work to make sure that the DAG underlying the workflow executes tasks in the
correct order.

## Setup-teardown pattern

The main use case of decorating `@workflow`-decorated functions is to establish a setup-teardown pattern to execute task
before and after your main workflow logic. This is useful when integrating with other external services
like [wandb](https://wandb.ai/site) or [clearml](https://clear.ml/), which enable you to track metrics of model
training runs.

To begin, import the necessary libraries.

```{code-cell}
from functools import partial, wraps
from unittest.mock import MagicMock

import flytekit
from flytekit import FlyteContextManager, task, workflow
from flytekit.core.node_creation import create_node
```

+++ {"lines_to_next_cell": 0}

Let's define the tasks we need for setup and teardown. In this example, we use the
{py:class}`unittest.mock.MagicMock` class to create a fake external service that we want to initialize at the
beginning of our workflow and finish at the end.

```{code-cell}
external_service = MagicMock()


@task
def setup():
    print("initializing external service")
    external_service.initialize(id=flytekit.current_context().execution_id)


@task
def teardown():
    print("finish external service")
    external_service.complete(id=flytekit.current_context().execution_id)
```

+++ {"lines_to_next_cell": 0}

As you can see, you can even use Flytekit's current context to access the `execution_id` of the current workflow
if you need to link Flyte with the external service so that you reference the same unique identifier in both the
external service and Flyte.

## Workflow decorator

We create a decorator that we want to use to wrap our workflow function.

```{code-cell}
def setup_teardown(fn=None, *, before, after):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        # get the current flyte context to obtain access to the compilation state of the workflow DAG.
        ctx = FlyteContextManager.current_context()

        # defines before node
        before_node = create_node(before)
        # ctx.compilation_state.nodes == [before_node]

        # under the hood, flytekit compiler defines and threads
        # together nodes within the `my_workflow` function body
        outputs = fn(*args, **kwargs)
        # ctx.compilation_state.nodes == [before_node, *nodes_created_by_fn]

        # defines the after node
        after_node = create_node(after)
        # ctx.compilation_state.nodes == [before_node, *nodes_created_by_fn, after_node]

        # compile the workflow correctly by making sure `before_node`
        # runs before the first workflow node and `after_node`
        # runs after the last workflow node.
        if ctx.compilation_state is not None:
            # ctx.compilation_state.nodes is a list of nodes defined in the
            # order of execution above
            workflow_node0 = ctx.compilation_state.nodes[1]
            workflow_node1 = ctx.compilation_state.nodes[-2]
            before_node >> workflow_node0
            workflow_node1 >> after_node
        return outputs

    if fn is None:
        return partial(setup_teardown, before=before, after=after)

    return wrapper
```

+++ {"lines_to_next_cell": 0}

There are a few key pieces to note in the `setup_teardown` decorator above:

1. It takes a `before` and `after` argument, both of which need to be `@task`-decorated functions. These
   tasks will run before and after the main workflow function body.
2. The [create_node](https://github.com/flyteorg/flytekit/blob/9e156bb0cf3d1441c7d1727729e8f9b4bbc3f168/flytekit/core/node_creation.py#L18) function
   to create nodes associated with the `before` and `after` tasks.
3. When `fn` is called, under the hood Flytekit creates all the nodes associated with the workflow function body
4. The code within the `if ctx.compilation_state is not None:` conditional is executed at compile time, which
   is where we extract the first and last nodes associated with the workflow function body at index `1` and `-2`.
5. The `>>` right shift operator ensures that `before_node` executes before the
   first node and `after_node` executes after the last node of the main workflow function body.

## Defining the DAG

We define two tasks that will constitute the workflow.

```{code-cell}
@task
def t1(x: float) -> float:
    return x - 1


@task
def t2(x: float) -> float:
    return x**2
```

+++ {"lines_to_next_cell": 0}

And then create our decorated workflow:

```{code-cell}
:lines_to_next_cell: 2

@workflow
@setup_teardown(before=setup, after=teardown)
def decorating_workflow(x: float) -> float:
    return t2(x=t1(x=x))


if __name__ == "__main__":
    print(decorating_workflow(x=10.0))
```

## Run the example on the Flyte cluster

To run the provided workflow on the Flyte cluster, use the following command:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/decorating_workflows.py \
  decorating_workflow --x 10.0
```

To define workflows imperatively, refer to {ref}`this example <imperative_workflow>`,
and to learn more about how to extend Flyte at a deeper level, for example creating custom types, custom tasks or
backend plugins, see {ref}`Extending Flyte <plugins_extend>`.
