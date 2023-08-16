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
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Using Schemas

```{eval-rst}
.. tags:: DataFrame, Basic
```

This example explains how an untyped schema is passed between tasks using {py:class}`pandas.DataFrame`.
Flytekit makes it possible for users to directly return or accept a {py:class}`pandas.DataFrame`, which are automatically
converted into flyte's abstract representation of a schema object

+++

:::{warning}
`FlyteSchema` is deprecated, use {ref}`structured_dataset_example` instead.
:::

```{code-cell}
import pandas
from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

Flytekit allows users to directly use pandas.dataframe in their tasks as long as they import
Note: # noqa: F401. This is to ignore pylint complaining about unused imports

```{code-cell}
from flytekit.types import schema  # noqa: F401
```

+++ {"lines_to_next_cell": 0}

This task generates a pandas.DataFrame and returns it. The Dataframe itself will be serialized to an intermediate
format like parquet before passing between tasks

```{code-cell}
@task
def get_df(a: int) -> pandas.DataFrame:
    """
    Generate a sample dataframe
    """
    return pandas.DataFrame(data={"col1": [a, 2], "col2": [a, 4]})
```

+++ {"lines_to_next_cell": 0}

This task shows an example of transforming a dataFrame

```{code-cell}
@task
def add_df(df: pandas.DataFrame) -> pandas.DataFrame:
    """
    Append some data to the dataframe.
    NOTE: this may result in runtime failures if the columns do not match
    """
    return df.append(pandas.DataFrame(data={"col1": [5, 10], "col2": [5, 10]}))
```

+++ {"lines_to_next_cell": 0}

The workflow shows that passing DataFrame's between tasks is as simple as passing dataFrames in memory

```{code-cell}
@workflow
def df_wf(a: int) -> pandas.DataFrame:
    """
    Pass data between the dataframes
    """
    df = get_df(a=a)
    return add_df(df=df)
```

+++ {"lines_to_next_cell": 0}

The entire program can be run locally

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running df_wf(a=42) {df_wf(a=42)}")
```
