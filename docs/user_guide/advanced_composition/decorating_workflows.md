(decorating_workflows)=

# Decorating workflows

```{eval-rst}
.. tags:: Intermediate
```

The behavior of workflows can be modified in a light-weight fashion by using the built-in {py:func}`~functools.wraps`
decorator pattern, similar to using decorators to
{ref}`customize task behavior <decorating_tasks>`. However, unlike in the case of
tasks, we need to do a little extra work to make sure that the DAG underlying the workflow executes tasks in the correct order.

## Setup-teardown pattern

The main use case of decorating `@workflow`-decorated functions is to establish a setup-teardown pattern to execute task
before and after your main workflow logic. This is useful when integrating with other external services
like [wandb](https://wandb.ai/site) or [clearml](https://clear.ml/), which enable you to track metrics of model training runs.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary libraries.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py
:caption: advanced_composition/decorating_workflows.py
:lines: 1-6
```

Let's define the tasks we need for setup and teardown. In this example, we use the
{py:class}`unittest.mock.MagicMock` class to create a fake external service that we want to initialize at the
beginning of our workflow and finish at the end.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py
:caption: advanced_composition/decorating_workflows.py
:lines: 9-21
```

As you can see, you can even use Flytekit's current context to access the `execution_id` of the current workflow
if you need to link Flyte with the external service so that you reference the same unique identifier in both the
external service and Flyte.

## Workflow decorator

We create a decorator that we want to use to wrap our workflow function.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py
:caption: advanced_composition/decorating_workflows.py
:pyobject: setup_teardown
```

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

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py
:caption: advanced_composition/decorating_workflows.py
:lines: 63-70
```

And then create our decorated workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py
:caption: advanced_composition/decorating_workflows.py
:lines: 74-82
```

## Run the example on the Flyte cluster

To run the provided workflow on the Flyte cluster, use the following command:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_workflows.py \
  decorating_workflow --x 10.0
```

To define workflows imperatively, refer to {ref}`this example <imperative_workflow>`,
and to learn more about how to extend Flyte at a deeper level, for example creating custom types, custom tasks or
backend plugins, see {ref}`Extending Flyte <plugins_extend>`.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
