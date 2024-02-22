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

(dataclass)=

# Dataclass

```{eval-rst}
.. tags:: Basic
```

When you've multiple values that you want to send across Flyte entities, you can use a `dataclass`.

Flytekit uses the [Mashumaro library](https://github.com/Fatal1ty/mashumaro)
to serialize and deserialize dataclasses.

:::{important}
If you're using Flytekit version below v1.10, you'll need to decorate with `@dataclass_json` using
`from dataclass_json import dataclass_json` instead of inheriting from Mashumaro's `DataClassJSONMixin`.
:::

To begin, import the necessary dependencies.

```{code-cell}
import os
import tempfile
from dataclasses import dataclass

import pandas as pd
from flytekit import task, workflow
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile
from flytekit.types.structured import StructuredDataset
from mashumaro.mixins.json import DataClassJSONMixin
```

+++ {"lines_to_next_cell": 0}

## Python types
We define a `dataclass` with `int`, `str` and `dict` as the data types.

```{code-cell}
@dataclass
class Datum(DataClassJSONMixin):
    x: int
    y: str
    z: dict[int, str]
```

+++ {"lines_to_next_cell": 0}

You can send a `dataclass` between different tasks written in various languages, and input it through the Flyte console as raw JSON.

:::{note}
All variables in a data class should be **annotated with their type**. Failure to do should will result in an error.
:::

Once declared, a dataclass can be returned as an output or accepted as an input.

```{code-cell}
@task
def stringify(s: int) -> Datum:
    """
    A dataclass return will be treated as a single complex JSON return.
    """
    return Datum(x=s, y=str(s), z={s: str(s)})


@task
def add(x: Datum, y: Datum) -> Datum:
    """
    Flytekit automatically converts the provided JSON into a data class.
    If the structures don't match, it triggers a runtime failure.
    """
    x.z.update(y.z)
    return Datum(x=x.x + y.x, y=x.y + y.y, z=x.z)
```

+++ {"lines_to_next_cell": 0}

## Flyte types
We also define a data class that accepts {std:ref}`StructuredDataset <structured_dataset>`,
{std:ref}`FlyteFile <files>` and {std:ref}`FlyteDirectory <folder>`.

```{code-cell}
@dataclass
class FlyteTypes(DataClassJSONMixin):
    dataframe: StructuredDataset
    file: FlyteFile
    directory: FlyteDirectory


@task
def upload_data() -> FlyteTypes:
    """
    Flytekit will upload FlyteFile, FlyteDirectory and StructuredDataset to the blob store,
    such as GCP or S3.
    """
    # 1. StructuredDataset
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})

    # 2. FlyteDirectory
    temp_dir = tempfile.mkdtemp(prefix="flyte-")
    df.to_parquet(temp_dir + "/df.parquet")

    # 3. FlyteFile
    file_path = tempfile.NamedTemporaryFile(delete=False)
    file_path.write(b"Hello, World!")

    fs = FlyteTypes(
        dataframe=StructuredDataset(dataframe=df),
        file=FlyteFile(file_path.name),
        directory=FlyteDirectory(temp_dir),
    )
    return fs


@task
def download_data(res: FlyteTypes):
    assert pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]}).equals(res.dataframe.open(pd.DataFrame).all())
    f = open(res.file, "r")
    assert f.read() == "Hello, World!"
    assert os.listdir(res.directory) == ["df.parquet"]
```

+++ {"lines_to_next_cell": 0}

A data class supports the usage of data associated with Python types, data classes,
flyte file, flyte directory and structured dataset.

We define a workflow that calls the tasks created above.

```{code-cell}
@workflow
def dataclass_wf(x: int, y: int) -> (Datum, FlyteTypes):
    o1 = add(x=stringify(s=x), y=stringify(s=y))
    o2 = upload_data()
    download_data(res=o2)
    return o1, o2
```

+++ {"lines_to_next_cell": 0}

You can run the workflow locally as follows:

```{code-cell}
if __name__ == "__main__":
    dataclass_wf(x=10, y=20)
```

To trigger a task that accepts a dataclass as an input with `pyflyte run`, you can provide a JSON file as an input:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/dataclass.py \
  add --x dataclass_input.json --y dataclass_input.json
```
