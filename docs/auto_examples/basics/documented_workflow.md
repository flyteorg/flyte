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

# Add Docstrings to Workflows

```{eval-rst}
.. tags:: Basic
```

Documented code helps enhance the readability of the code. Flyte supports docstrings to document your code.
Docstrings are stored in FlyteAdmin and shown on the UI in the launch form.

This example demonstrates the various ways in which you can document your Flyte workflow.
The example workflow accepts a DataFrame and data class. We send a record that needs to be appended to the DataFrame through a data class.

+++ {"lines_to_next_cell": 0}

Let's import the libraries.

```{code-cell}
from dataclasses import dataclass

import pandas as pd
from dataclasses_json import dataclass_json
from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

We define a dataclass.

```{code-cell}
@dataclass_json
@dataclass
class PandasData(object):
    id: int = 3
    name: str = "Bonnie"
```

+++ {"lines_to_next_cell": 0}

Next, we define a task that appends data to the DataFrame.

```{code-cell}
@task
def add_data(df: pd.DataFrame, data: PandasData) -> pd.DataFrame:
    df = df.append({"id": data.id, "name": data.name}, ignore_index=True)
    return df
```

+++ {"lines_to_next_cell": 0}

## Sphinx-style Docstring

An example to demonstrate Sphinx-style docstring.

The first block of the docstring is a one-liner about the workflow.
The second block of the docstring consists of a detailed description.
The third block of the docstring describes the parameters and return type.

```{code-cell}
@workflow
def sphinx_docstring(df: pd.DataFrame, data: PandasData = PandasData()) -> pd.DataFrame:
    """
    Showcase Sphinx-style docstring.

    This workflow accepts a DataFrame and data class.
    It calls a task that appends the user-sent record to the DataFrame.

    :param df: Pandas DataFrame
    :param data: A data class pertaining to the new record to be stored in the DataFrame
    :return: Pandas DataFrame
    """
    return add_data(df=df, data=data)
```

+++ {"lines_to_next_cell": 0}

## NumPy-style Docstring

An example to demonstrate NumPy-style docstring.

The first block of the docstring is a one-liner about the workflow.
The second block of the docstring consists of a detailed description.
The third block of the docstring describes all the parameters along with their data types.
The fourth block of the docstring describes the return type along with its data type.

```{code-cell}
@workflow
def numpy_docstring(df: pd.DataFrame, data: PandasData = PandasData()) -> pd.DataFrame:
    """
    Showcase NumPy-style docstring.

    This workflow accepts a DataFrame and data class.
    It calls a task that appends the user-sent record to the DataFrame.

    Parameters
    ----------
    df: pd.DataFrame
        Pandas DataFrame
    data: Dataclass
        A data class pertaining to the new record to be stored in the DataFrame

    Returns
    -------
    out : pd.DataFrame
        Pandas DataFrame
    """
    return add_data(df=df, data=data)
```

+++ {"lines_to_next_cell": 0}

## Google-style Docstring

An example to demonstrate Google-style docstring.

The first block of the docstring is a one-liner about the workflow.
The second block of the docstring consists of a detailed description.
The third block of the docstring describes the parameters and return type along with their data types.

```{code-cell}
@workflow
def google_docstring(df: pd.DataFrame, data: PandasData = PandasData()) -> pd.DataFrame:
    """
    Showcase Google-style docstring.

    This workflow accepts a DataFrame and data class.
    It calls a task that appends the user-sent record to the DataFrame.

    Args:
        df(pd.DataFrame): Pandas DataFrame
        data(Dataclass): A data class pertaining to the new record to be stored in the DataFrame
    Returns:
        pd.DataFrame: Pandas DataFrame
    """
    return add_data(df=df, data=data)
```

+++ {"lines_to_next_cell": 0}

Lastly, we can run the workflow locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(
        f"Running sphinx_docstring(), modified DataFrame is {sphinx_docstring(df=pd.DataFrame(data={'id': [1, 2], 'name': ['John', 'Meghan']}),data=PandasData(id=3, name='Bonnie'))}"
    )
    print(
        f"Running numpy_docstring(), modified DataFrame is {numpy_docstring(df=pd.DataFrame(data={'id': [1, 2], 'name': ['John', 'Meghan']}),data=PandasData(id=3, name='Bonnie'))}"
    )
    print(
        f"Running google_docstring(), modified DataFrame is {google_docstring(df=pd.DataFrame(data={'id': [1, 2], 'name': ['John', 'Meghan']}),data=PandasData(id=3, name='Bonnie'))}"
    )
```
