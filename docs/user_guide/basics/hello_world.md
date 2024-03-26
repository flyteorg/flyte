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


# Hello, World!

```{eval-rst}
.. tags:: Basic
```

Let's write a Flyte {py:func}`~flytekit.workflow` that invokes a
{py:func}`~flytekit.task` to generate the output "Hello, World!".

Flyte tasks are the core building blocks of larger, more complex workflows.
Workflows compose multiple tasks – or other workflows –
into meaningful steps of computation to produce some useful set of outputs or outcomes.

To begin, import `task` and `workflow` from the `flytekit` library.

```{code-cell}
from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

Define a task that produces the string "Hello, World!".
Simply using the `@task` decorator to annotate the Python function.

```{code-cell}
@task
def say_hello() -> str:
    return "Hello, World!"
```

+++ {"lines_to_next_cell": 0}

You can handle the output of a task in the same way you would with a regular Python function.
Store the output in a variable and use it as a return value for a Flyte workflow.

```{code-cell}
@workflow
def hello_world_wf() -> str:
    res = say_hello()
    return res
```

+++ {"lines_to_next_cell": 0}

Run the workflow by simply calling it like a Python function.

```{code-cell}
:lines_to_next_cell: 2

if __name__ == "__main__":
    print(f"Running hello_world_wf() {hello_world_wf()}")
```

Next, let's delve into the specifics of {ref}`tasks <task>`,
{ref}`workflows <workflow>` and {ref}`launch plans <launch_plan>`.
