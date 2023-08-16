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

+++ {"lines_to_next_cell": 0}

# Quickstart

In this demo and following example, learn how to use `DoltTable` to annotate DataFrame inputs and outputs in the Flyte tasks.

```{eval-rst}
.. youtube:: TfBIuHYkyvU

```

+++ {"lines_to_next_cell": 0}

First, let's import the libraries.

```{code-cell}
import os
import sys

import pandas as pd
from flytekit import task, workflow
from flytekitplugins.dolt.schema import DoltConfig, DoltTable
```

+++ {"lines_to_next_cell": 0}

Next, we initialize Dolt's config.

```{code-cell}
doltdb_path = os.path.join(os.path.dirname(__file__), "foo")

rabbits_conf = DoltConfig(
    db_path=doltdb_path,
    tablename="rabbits",
)
```

We define a task to create a DataFrame and store the table in Dolt.

```{code-cell}
@task
def populate_rabbits(a: int) -> DoltTable:
    rabbits = [("George", a), ("Alice", a * 2), ("Sugar Maple", a * 3)]
    df = pd.DataFrame(rabbits, columns=["name", "count"])
    return DoltTable(data=df, config=rabbits_conf)
```

+++ {"lines_to_next_cell": 0}

`unwrap_rabbits` task does the exact opposite -- reading the table from Dolt and returning a DataFrame.

```{code-cell}
@task
def unwrap_rabbits(table: DoltTable) -> pd.DataFrame:
    return table.data
```

+++ {"lines_to_next_cell": 0}

Our workflow combines the above two tasks:

```{code-cell}
@workflow
def wf(a: int) -> pd.DataFrame:
    rabbits = populate_rabbits(a=a)
    df = unwrap_rabbits(table=rabbits)
    return df


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    if len(sys.argv) != 2:
        raise ValueError("Expected 1 argument: a (int)")
    a = int(sys.argv[1])
    result = wf(a=a)
    print(f"Running wf(), returns dataframe\n{result}\n{result.dtypes}")
```

Run this task by issuing the following command:

```{eval-rst}
.. prompt:: $

  python quickstart_example.py 1
```
