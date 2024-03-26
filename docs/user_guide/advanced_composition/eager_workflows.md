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

(eager_workflows)=

# Eager workflows

```{eval-rst}
.. tags:: Intermediate
```

```{important}
This feature is experimental and the API is subject to breaking changes.
If you encounter any issues please consider submitting a
[bug report](https://github.com/flyteorg/flyte/issues/new?assignees=&labels=bug%2Cuntriaged&projects=&template=bug_report.yaml&title=%5BBUG%5D+).
```

So far, the two types of workflows you've seen are static workflows, which
are defined with `@workflow`-decorated functions or imperative `Workflow` class,
and dynamic workflows, which are defined with the `@dynamic` decorator.

{ref}`Static workflows <workflow>` are created at compile time when you call `pyflyte run`,
`pyflyte register`, or `pyflyte serialize`. This means that the workflow is static
and cannot change its shape at any point: all of the variables defined as an input
to the workflow or as an output of a task or subworkflow are promises.
{ref}`Dynamic workflows <dynamic_workflow>`, on the other hand, are compiled
at runtime so that they can materialize the inputs of the workflow as Python values
and use them to determine the shape of the execution graph.

In this guide you'll learn how to use eager workflows, which allow you to
create extremely flexible workflows that give you run-time access to
intermediary task/subworkflow outputs.

## Why eager workflows?

Both static and dynamic workflows have a key limitation: while they provide
compile-time and run-time type safety, respectively, they both suffer from
inflexibility in expressing asynchronous execution graphs that many Python
programmers may be accustomed to by using, for example, the
[asyncio](https://docs.python.org/3/library/asyncio.html) library.

Unlike static and dynamic workflows, eager workflows allow you to use all of
the python constructs that you're familiar with via the `asyncio` API. To
understand what this looks like, let's define a very basic eager workflow
using the `@eager` decorator.

```{code-cell}
:lines_to_next_cell: 2

from flytekit import task, workflow
from flytekit.experimental import eager


@task
def add_one(x: int) -> int:
    return x + 1


@task
def double(x: int) -> int:
    return x * 2


@eager
async def simple_eager_workflow(x: int) -> int:
    out = await add_one(x=x)
    if out < 0:
        return -1
    return await double(x=out)
```

+++ {"lines_to_next_cell": 2}

As we can see in the code above, we're defining an `async` function called
`simple_eager_workflow` that takes an integer as input and returns an integer.
By decorating this function with `@eager`, we now have the ability to invoke
tasks, static subworkflows, and even other eager subworkflows in an _eager_
fashion such that we can materialize their outputs and use them inside the
parent eager workflow itself.

In the `simple_eager_workflow` function, we can see that we're `await`ing
the output of the `add_one` task and assigning it to the `out` variable. If
`out` is a negative integer, the workflow will return `-1`. Otherwise, it
will double the output of `add_one` and return it.

Unlike in static and dynamic workflows, this variable is actually
the Python integer that is the result of `x + 1` and not a promise.

## How it works

When you decorate a function with `@eager`, any function invoked within it
that's decorated with `@task`, `@workflow`, or `@eager` becomes
an [awaitable](https://docs.python.org/3/library/asyncio-task.html#awaitables)
object within the lifetime of the parent eager workflow execution. Note that
this happens automatically and you don't need to use the `async` keyword when
defining a task or workflow that you want to invoke within an eager workflow.

```{important}
With eager workflows, you basically have access to the Python `asyncio`
interface to define extremely flexible execution graphs! The trade-off is that
you lose the compile-time type safety that you get with regular static workflows
and to a lesser extent, dynamic workflows.

We're leveraging Python's native `async` capabilities in order to:

1. Materialize the output of flyte tasks and subworkflows so you can operate
   on them without spinning up another pod and also determine the shape of the
   workflow graph in an extremely flexible manner.
2. Provide an alternative way of achieving concurrency in Flyte. Flyte has
   concurrency built into it, so all tasks/subworkflows will execute concurrently
   assuming that they don't have any dependencies on each other. However, eager
   workflows provide a python-native way of doing this, with the main downside
   being that you lose the benefits of statically compiled workflows such as
   compile-time analysis and first-class data lineage tracking.
```

Similar to {ref}`dynamic workflows <dynamic_workflow>`, eager workflows are
actually tasks. The main difference is that, while dynamic workflows compile
a static workflow at runtime using materialized inputs, eager workflows do
not compile any workflow at all. Instead, they use the {py:class}`~flytekit.remote.remote.FlyteRemote`
object together with Python's `asyncio` API to kick off tasks and subworkflow
executions eagerly whenever you `await` on a coroutine. This means that eager
workflows can materialize an output of a task or subworkflow and use it as a
Python object in the underlying runtime environment. We'll see how to configure
`@eager` functions to run on a remote Flyte cluster
{ref}`later in this guide <eager_workflows_remote>`.

## What can you do with eager workflows?

In this section we'll cover a few of the use cases that you can accomplish
with eager workflows, some of which you can't accomplish with static or dynamic
workflows.

### Operating on task and subworkflow outputs

One of the biggest benefits of eager workflows is that you can now materialize
task and subworkflow outputs as Python values and do operations on them just
like you would in any other Python function. Let's look at another example:

```{code-cell}
@eager
async def another_eager_workflow(x: int) -> int:
    out = await add_one(x=x)

    # out is a Python integer
    out = out - 1

    return await double(x=out)
```

+++ {"lines_to_next_cell": 0}

Since out is an actual Python integer and not a promise, we can do operations
on it at runtime, inside the eager workflow function body. This is not possible
with static or dynamic workflows.

### Pythonic conditionals

As you saw in the `simple_eager_workflow` workflow above, you can use regular
Python conditionals in your eager workflows. Let's look at a more complicated
example:

```{code-cell}
:lines_to_next_cell: 2

@task
def gt_100(x: int) -> bool:
    return x > 100


@eager
async def eager_workflow_with_conditionals(x: int) -> int:
    out = await add_one(x=x)

    if out < 0:
        return -1
    elif await gt_100(x=out):
        return 100
    else:
        out = await double(x=out)

    assert out >= -1
    return out
```

In the above example, we're using the eager workflow's Python runtime
to check if `out` is negative, but we're also using the `gt_100` task in the
`elif` statement, which will be executed in a separate Flyte task.

### Loops

You can also gather the outputs of multiple tasks or subworkflows into a list:

```{code-cell}
import asyncio


@eager
async def eager_workflow_with_for_loop(x: int) -> int:
    outputs = []

    for i in range(x):
        outputs.append(add_one(x=i))

    outputs = await asyncio.gather(*outputs)
    return await double(x=sum(outputs))
```

+++ {"lines_to_next_cell": 0}

### Static subworkflows

You can also invoke static workflows from within an eager workflow:

```{code-cell}
:lines_to_next_cell: 2

@workflow
def subworkflow(x: int) -> int:
    out = add_one(x=x)
    return double(x=out)


@eager
async def eager_workflow_with_static_subworkflow(x: int) -> int:
    out = await subworkflow(x=x)
    assert out == (x + 1) * 2
    return out
```

+++ {"lines_to_next_cell": 0}

### Eager subworkflows

You can have nest eager subworkflows inside a parent eager workflow:

```{code-cell}
:lines_to_next_cell: 2

@eager
async def eager_subworkflow(x: int) -> int:
    return await add_one(x=x)


@eager
async def nested_eager_workflow(x: int) -> int:
    out = await eager_subworkflow(x=x)
    return await double(x=out)
```

+++ {"lines_to_next_cell": 0}

### Catching exceptions

You can also catch exceptions in eager workflows through `EagerException`:

```{code-cell}
:lines_to_next_cell: 2

from flytekit.experimental import EagerException


@task
def raises_exc(x: int) -> int:
    if x <= 0:
        raise TypeError
    return x


@eager
async def eager_workflow_with_exception(x: int) -> int:
    try:
        return await raises_exc(x=x)
    except EagerException:
        return -1
```

Even though the `raises_exc` exception task raises a `TypeError`, the
`eager_workflow_with_exception` runtime will raise an `EagerException` and
you'll need to specify `EagerException` as the exception type in your `try... except`
block.

```{note}
This is a current limitation in the `@eager` workflow implementation.
````

## Executing eager workflows

As with most Flyte constructs, you can execute eager workflows both locally
and remotely.

### Local execution

You can execute eager workflows locally by simply calling them like a regular
`async` function:

```{code-cell}
if __name__ == "__main__":
    result = asyncio.run(simple_eager_workflow(x=5))
    print(f"Result: {result}")  # "Result: 12"
```

This just uses the `asyncio.run` function to execute the eager workflow just
like any other Python async code. This is useful for local debugging as you're
developing your workflows and tasks.

(eager_workflows_remote)=

### Remote Flyte cluster execution

Under the hood, `@eager` workflows use the {py:class}`~flytekit.remote.remote.FlyteRemote`
object to kick off task, static workflow, and eager workflow executions.

In order to actually execute them on a Flyte cluster, you'll need to configure
eager workflows with a `FlyteRemote` object and secrets configuration that
allows you to authenticate into the cluster via a client secret key.

```{code-block} python
from flytekit.remote import FlyteRemote
from flytekit.configuration import Config

@eager(
    remote=FlyteRemote(
        config=Config.auto(config_file="config.yaml"),
        default_project="flytesnacks",
        default_domain="development",
    ),
    client_secret_group="<my_client_secret_group>",
    client_secret_key="<my_client_secret_key>",
)
async def eager_workflow_remote(x: int) -> int:
    ...
```

+++

Where `config.yaml` contains a
[flytectl](https://docs.flyte.org/projects/flytectl/en/latest/#configuration)-compatible
config file and `my_client_secret_group` and `my_client_secret_key` are the
{ref}`secret group and key <secrets>` that you've configured for your Flyte
cluster to authenticate via a client key.

+++

### Sandbox Flyte cluster execution

When using a sandbox cluster started with `flytectl demo start`, however, the
`client_secret_group` and `client_secret_key` are not required, since the
default sandbox configuration does not require key-based authentication.

```{code-cell}
:lines_to_next_cell: 2

from flytekit.configuration import Config
from flytekit.remote import FlyteRemote


@eager(
    remote=FlyteRemote(
        config=Config.for_sandbox(),
        default_project="flytesnacks",
        default_domain="development",
    )
)
async def eager_workflow_sandbox(x: int) -> int:
    out = await add_one(x=x)
    if out < 0:
        return -1
    return await double(x=out)
```

```{important}
When executing eager workflows on a remote Flyte cluster, it will execute the
latest version of tasks, static workflows, and eager workflows that are on
the `default_project` and `default_domain` as specified in the `FlyteRemote`
object. This means that you need to pre-register all Flyte entities that are
invoked inside of the eager workflow.
```

### Registering and running

Assuming that your `flytekit` code is configured correctly, you will need to
register all of the task and subworkflows that are used with your eager
workflow with `pyflyte register`:

```{prompt} bash
pyflyte --config <path/to/config.yaml> register \
 --project <project> \
 --domain <domain> \
 --image <image> \
 path/to/eager_workflows.py
```

And then run it with `pyflyte run`:

```{prompt} bash
pyflyte --config <path/to/config.yaml> run \
 --project <project> \
 --domain <domain> \
 --image <image> \
 path/to/eager_workflows.py simple_eager_workflow --x 10
```

```{note}
You need to register the tasks/workflows associated with your eager workflow
because eager workflows are actually flyte tasks under the hood, which means
that `pyflyte run` has no way of knowing what tasks and subworkflows are
invoked inside of it.
```

## Eager workflows on Flyte console

Since eager workflows are an experimental feature, there is currently no
first-class representation of them on Flyte Console, the UI for Flyte.
When you register an eager workflow, you'll be able to see it in the task view:

:::{figure} https://github.com/flyteorg/static-resources/blob/main/flytesnacks/user_guide/flyte_eager_workflow_ui_view.png?raw=true
:alt: Eager Workflow UI View
:class: with-shadow
:::

When you execute an eager workflow, the tasks and subworkflows invoked within
it **won't show up** on the node, graph, or timeline view. As mentioned above,
this is because eager workflows are actually Flyte tasks under the hood and
Flyte has no way of knowing the shape of the execution graph before actually
executing them.

:::{figure} https://github.com/flyteorg/static-resources/blob/main/flytesnacks/user_guide/flyte_eager_workflow_execution.png?raw=true
:alt: Eager Workflow Execution
:class: with-shadow
:::

However, at the end of execution, you'll be able to use {ref}`Flyte Decks <deck>`
to see a list of all the tasks and subworkflows that were executed within the
eager workflow:

:::{figure} https://github.com/flyteorg/static-resources/blob/main/flytesnacks/user_guide/flyte_eager_workflow_deck.png?raw=true
:alt: Eager Workflow Deck
:class: with-shadow
:::

## Limitations

As this feature is still experimental, there are a few limitations that you
need to keep in mind:

- You cannot invoke {ref}`dynamic workflows <dynamic_workflow>`,
  {ref}`map tasks <map_task>`, or {ref}`launch plans <launch_plan>` inside an
  eager workflow.
- [Context managers](https://docs.python.org/3/library/contextlib.html) will
  only work on locally executed functions within the eager workflow, i.e. using a
  context manager to modify the behavior of a task or subworkflow will not work
  because they are executed on a completely different pod.
- All exceptions raised by Flyte tasks or workflows will be caught and raised
  as an {py:class}`~flytekit.experimental.EagerException` at runtime.
- All task/subworkflow outputs are materialized as Python values, which includes
  offloaded types like `FlyteFile`, `FlyteDirectory`, `StructuredDataset`, and
  `pandas.DataFrame` will be fully downloaded into the pod running the eager workflow.
  This prevents you from incrementally downloading or streaming very large datasets
  in eager workflows.
- Flyte entities that are invoked inside of an eager workflow must be registered
  under the same project and domain as the eager workflow itself. The eager
  workflow will execute the latest version of these entities.
- Flyte console currently does not have a first-class way of viewing eager
  workflows, but it can be accessed via the task list view and the execution
  graph is viewable via Flyte Decks.

## Summary of workflows

Eager workflows are a powerful new construct that trades-off compile-time type
safety for flexibility in the shape of the execution graph. The table below
will help you to reason about the different workflow constructs in Flyte in terms
of promises and materialized values:

| Construct | Description | Flyte Promises | Pro | Con |
|--------|--------|--------|----|----|
| `@workflow` | Compiled at compile-time | All inputs and intermediary outputs are promises | Type errors caught at compile-time | Constrained by Flyte DSL |
| `@dynamic` | Compiled at run-time | Inputs are materialized, but outputs of all Flyte entities are Promises | More flexible than `@workflow`, e.g. can do Python operations on inputs | Can't use a lot of Python constructs (e.g. try/except) |
| `@eager` | Never compiled | Everything is materialized! | Can effectively use all Python constructs via `asyncio` syntax | No compile-time benefits, this is the wild west 🏜 |
