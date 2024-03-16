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

# Reference tasks

```{eval-rst}
.. tags:: Intermediate
```

A {py:func}`flytekit.reference_task` references the Flyte tasks that have already been defined, serialized, and registered.
You can reference tasks from other projects and create workflows that use tasks declared by others.
These tasks can be in their own containers, python runtimes, flytekit versions, and even different languages.

The following example illustrates how to use reference tasks.

:::{note}
Reference tasks cannot be run locally. You must mock them out.
:::

```{code-cell}
:lines_to_next_cell: 2

from typing import List

from flytekit import reference_task, workflow
from flytekit.types.file import FlyteFile


@reference_task(
    project="flytesnacks",
    domain="development",
    name="data_types_and_io.file.normalize_columns",
    version="{{ registration.version }}",
)
def normalize_columns(
    csv_url: FlyteFile,
    column_names: List[str],
    columns_to_normalize: List[str],
    output_location: str,
) -> FlyteFile:
    ...


@workflow
def wf() -> FlyteFile:
    return normalize_columns(
        csv_url="https://people.sc.fsu.edu/~jburkardt/data/csv/biostats.csv",
        column_names=["Name", "Sex", "Age", "Heights (in)", "Weight (lbs)"],
        columns_to_normalize=["Age"],
        output_location="",
    )
```

:::{note}
The macro `{{ registration.version }}` is populated by `flytectl register` during registration.
Generally, it is unnecessary for reference tasks, as it is preferable to bind to a specific version of the task or launch plan.
However, in this example, we are registering both the task `core.flyte_basics.files.normalize_columns` and the workflow that references it.
Therefore, we need the macro to be updated to the version of a specific Flytesnacks release.
This is why `{{ registration.version }}` is used.

A typical reference task would resemble the following:

```python
@reference_task(
     project="flytesnacks",
     domain="development",
     name="core.flyte_basics.files.normalize_columns",
     version="d06cebcfbeabc02b545eefa13a01c6ca992940c8", # If using GIT for versioning OR 0.16.0, if semver
 )
 def normalize_columns(...):
     ...
```
:::
