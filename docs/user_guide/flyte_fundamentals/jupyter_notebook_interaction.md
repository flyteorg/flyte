---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_jupyter_notebook_interaction)=

# Running and developing workflows in Jupyter notebooks

Flyte supports the development, running, and debugging of tasks and workflows in an interactive
Jupyter notebook environment, which accelerates the iteration speed when building data-
or machine learning-driven applications.

```{admonition} Attention
:class: attention

This feature requires the `flytekit` version `1.14.0` or higher.
```

```{admonition} Prerequisites
:class: important

This guide assumes that you've completed the previous guides for
{ref}`Running and Scheduling Workflows <getting_started_run_and_schedule>`.
The code snippets in this guide are intended to be run in a Jupyter notebook.
```

The code of this guide can be found in the [flytesnacks](https://github.com/flyteorg/flytesnacks/blob/master/examples/basics/basics/basic_interactive_mode.ipynb)

## Create an interactive `FlyteRemote` object

In {ref}`Running and Scheduling Workflows <getting_started_run_and_schedule>`, you learned
how to run registered Flyte workflows from a Python runtime using the
{py:class}`~flytekit.remote.remote.FlyteRemote` client.

When developing workflows in a Jupyter notebook, `FlyteRemote` provides an
interactive interface to register and run workflows on a Flyte cluster. Let's
create an interactive `FlyteRemote` object:

```{code-cell} ipython3
:tags: [remove-output]

from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

remote = FlyteRemote(
    config=Config.auto(),
    default_project="flytesnacks",
    default_domain="development",
    interactive_mode_enabled=True,
)
```

```{admonition} Note
:class: Note

The `interactive_mode_enabled` flag is automatically set to `True` when running
in a Jupyter notebook environment, enabling interactive registration and execution
of workflows.
```

## Running a task or a workflow

You can run entities (tasks or workflows) using the `FlyteRemote`
{py:meth}`~flytekit.remote.remote.FlyteRemote.execute` method.
During execution, `flytekit` first checks if the entity is registered with the
Flyte backend, and if not, registers it before execution.

```{code-block} python
execution = remote.execute(my_task, inputs={"name": "Flyte"})
execution = remote.execute(my_wf, inputs={"name": "Flyte"})
```

You can then fetch the inputs and outputs of the execution by following the steps
in {ref}`<getting_started_run_and_schedule_fetch_execution>`.

## When Does Interactive `FlyteRemote` Re-register an Entity?

The interactive `FlyteRemote` client re-registers an entity whenever it's
redefined in the notebook, including when you re-execute a cell containing the
entity definition, even if the entity remains unchanged. This behavior facilitates
iterative development and debugging of tasks and workflows in a Jupyter notebook.

## What's next?

In this guide, you learned how to develop and run tasks and workflows in a
Jupyter Notebook environment using interactive `FlyteRemote`.

In the next guide, you'll learn how to visualize tasks using Flyte Decks.
