(eager_workflows)=

# Eager workflows

```{eval-rst}
.. tags:: Intermediate
```

```{important}
This feature is in beta and the API is still subject to minor changes.
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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 1-21
```

As we can see in the code above, we're defining an `async` function called
`simple_eager_workflow` that takes an integer as input and returns an integer.
By decorating this function with `@eager`, we now have the ability to invoke
tasks, static subworkflows, and even other eager subworkflows in an _eager_
fashion such that we can materialize their outputs and use them inside the
parent eager workflow itself.

In the `simple_eager_workflow` function, we can call the `add_one` task and assigning it to the `out` variable. If
`out` is a negative integer, the workflow will return `-1`. Otherwise, it
will double the output of `add_one` and return it.

Unlike in static and dynamic workflows, this variable is actually
the Python integer that is the result of `x + 1` and not a promise.

## How it works

### Parallels to Python `asyncio`
The eager paradigm was written around Python's native `async` functionality. As such, it follows the same rules and
constructs and if you're used to working with async functions, you should be able to apply the exact same understanding to work with eager tasks.

In the example above, the tasks `add_one` and `double` are normal Flyte tasks and the functions being decorated are normal Python functions. This means that in the execution of the async task `simple_eager_workflow` will block on each of these functions just like Python would if these were simply just Python functions.

If you want to run functions in parallel, you will need to use `async` marked tasks, just like you would in normal Python.

Note that `eager` tasks share the same limitation as Python async functions.  You can only call an `async` function inside another `async` function, or within a special handler (like `asyncio.run`). This means that until the `@workflow` decorator supports async workflow function definitions, which is doesn't today, you will not be able to call eager tasks or other async Python function tasks, inside workflows. This functionality is slated to be added in future releases. For the time being, you will need to run the tasks directly, either from FlyteRemote or the Flyte UI.

Unlike Python async however, when an `eager` task runs `async` sub-tasks in a real backend execution (not a local execution), it is doing real, wall-clock time parallelism, not just concurrency (assuming your K8s cluster is appropriately sized).

```{important}
With eager workflows, you basically have access to the Python `asyncio`
interface to define extremely flexible execution graphs! The trade-off is that
you lose the compile-time type safety that you get with regular static workflows
and to a lesser extent, dynamic workflows.

We're leveraging Python's native `async` capabilities in order to:

1. Materialize the output of flyte tasks and subworkflows so you can operate
   on them without spinning up another pod and also determine the shape of the
   workflow graph in an extremely flexible manner.
2. Provide an alternative way of achieving wall-time parallelism in Flyte. Flyte has
   concurrency built into it, so all tasks/subworkflows will execute concurrently
   assuming that they don't have any dependencies on each other. However, `eager`
   tasks provide a Python-native way of doing this, with the main downside
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

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:pyobject: another_eager_workflow
```

Since out is an actual Python integer and not a promise, we can do operations
on it at runtime, inside the eager workflow function body. This is not possible
with static or dynamic workflows.

### Pythonic conditionals

As you saw in the `simple_eager_workflow` workflow above, you can use regular
Python conditionals in your eager workflows. Let's look at a more complicated
example:

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 36-53
```

In the above example, we're using the eager workflow's Python runtime
to check if `out` is negative, but we're also using the `gt_100` task in the
`elif` statement, which will be executed in a separate Flyte task.

### Loops

You can also gather the outputs of multiple tasks or subworkflows into a list. Keep in mind that in this case, you will need to use an `async` function task, since normal tasks will block.

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 58-69
```

### Static subworkflows

You can also invoke static workflows from within an eager workflow:

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 74-84
```

### Eager subworkflows

You can have nest eager subworkflows inside a parent eager workflow:

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 89-97
```

### Catching exceptions

You can also catch exceptions in eager workflows through `EagerException`:

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 102-117
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

```{literalinclude} /examples/advanced_composition/advanced_composition/eager_workflows.py
:caption: advanced_composition/eager_workflows.py
:lines: 123-125
```

This just uses the `asyncio.run` function to execute the eager workflow just
like any other Python async code. This is useful for local debugging as you're
developing your workflows and tasks.

(eager_workflows_remote)=

### Setting up remote Flyte cluster access

Under the hood, `@eager` workflows use the {py:class}`~flytekit.remote.remote.FlyteRemote`
object to kick off task, static workflow, and eager workflow executions. In order to create a `FlyteRemote` instance, `Config.auto()` is run and the resulting config object is passed into `FlyteRemote`. 

This means that you just need to ensure eager workflow pods are run with the environment variables that including any mounted secrets for a client secret. For instance, the following three are sufficient in the most basic setup.

```{code-block} bash
FLYTE_PLATFORM_URL
FLYTE_CREDENTIALS_CLIENT_ID
FLYTE_CREDENTIALS_CLIENT_SECRET
```

See the relevant authentication docs for creating client credentials if using Flyte's internal authorization server.

### Sandbox Flyte cluster execution

When using a sandbox cluster started with `flytectl demo start` no authentication is needed and eager workflows should
just work out of the box.

```{important}
When executing eager workflows on a remote Flyte cluster, it will execute the
same version of tasks, static workflows, and eager workflows that are on
the `project` and `domain` as the eager task itself. If an entity is not found, FlyteRemote will attempt to register 
it. Please be aware of this and potential naming errors due to difference in folder paths when running in the container.
Future work may be done to allow execution in another project/domain, but reference entities should always be
correctly reflected and invoked.
```

### Registering and running

Assuming that your `flytekit` code is configured correctly, you should
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

Since eager workflows are still in beta, there is currently no
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

As this feature is still in beta, there are a few limitations that you
need to keep in mind:

- [Context managers](https://docs.python.org/3/library/contextlib.html) will
  only work on locally executed functions within the eager workflow, i.e. using a
  context manager to modify the behavior of a task or subworkflow will not work
  because they are executed on a completely different pod.
- All exceptions raised by Flyte tasks or workflows will be caught and raised
  as an {py:class}`~flytekit.exceptions.eager.EagerException` at runtime.
- All task/subworkflow outputs are materialized as Python values, which includes
  offloaded types like `FlyteFile`, `FlyteDirectory`, `StructuredDataset`, and
  `pandas.DataFrame` will be fully downloaded into the pod running the eager workflow.
  This prevents you from incrementally downloading or streaming very large datasets
  in eager workflows. (Please reach out to the team if you are interested in improving this.)
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

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
