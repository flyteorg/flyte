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

(structured_dataset_example)=

# Structured Dataset

```{eval-rst}
.. tags:: DataFrame, Basic, Data
```

Structured dataset is a superset of Flyte Schema.

The `StructuredDataset` Transformer can write a dataframe to BigQuery, s3, or any storage by registering new structured dataset encoder and decoder.

Flytekit makes it possible to return or accept a {py:class}`pandas.DataFrame` which is automatically
converted into Flyte's abstract representation of a structured dataset object.

This example explains how a structured dataset can be used with the Flyte entities.

+++ {"lines_to_next_cell": 0}

Let's import the necessary dependencies.

```{code-cell}
import os
import typing

import numpy as np
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq
from flytekit import FlyteContext, StructuredDatasetType, kwtypes, task, workflow
from flytekit.models import literals
from flytekit.models.literals import StructuredDatasetMetadata
from flytekit.types.schema import FlyteSchema
from flytekit.types.structured.structured_dataset import (
    PARQUET,
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetTransformerEngine,
)
from typing_extensions import Annotated
```

+++ {"lines_to_next_cell": 0}

We define the columns types for schema and `StructuredDataset`.

```{code-cell}
superset_cols = kwtypes(Name=str, Age=int, Height=int)
subset_cols = kwtypes(Age=int)
```

+++ {"lines_to_next_cell": 0}

We define two tasks, one returns a pandas DataFrame and the other a `FlyteSchema`.
Flyte serializes the DataFrames to an intermediate format, a parquet file, before sending them to the other tasks.

```{code-cell}
@task
def get_df(a: int) -> Annotated[pd.DataFrame, superset_cols]:
    """
    Generate a sample dataframe
    """
    return pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [a, 22], "Height": [160, 178]})


@task
def get_schema_df(a: int) -> FlyteSchema[superset_cols]:
    """
    Generate a sample dataframe
    """
    s = FlyteSchema()
    s.open().write(pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [a, 22], "Height": [160, 178]}))
    return s
```

+++ {"lines_to_next_cell": 0}

Next, we define a task that opens a structured dataset by calling `all()`.
When we invoke `all()`, the Flyte engine downloads the parquet file on S3, and deserializes it to `pandas.dataframe`.

:::{note}
- Despite the input type of the task being `StructuredDataset`, it can also accept FlyteSchema as input.
- The code may result in runtime failures if the columns do not match.
:::

```{code-cell}
@task
def get_subset_df(df: Annotated[StructuredDataset, subset_cols]) -> Annotated[StructuredDataset, subset_cols]:
    df = df.open(pd.DataFrame).all()
    df = pd.concat([df, pd.DataFrame([[30]], columns=["Age"])])
    # On specifying BigQuery uri for StructuredDataset, flytekit writes a pandas dataframe to a BigQuery table
    return StructuredDataset(dataframe=df)
```

## StructuredDataset with `uri` Argument

BigQuery `uri` allows you to load and retrieve data from cloud using the `uri`. The `uri` comprises of the bucket name and the filename prefixed with `gs://`.
If you specify BigQuery `uri` for StructuredDataset, BigQuery creates a table in the location specified by the `uri`.
The `uri` in StructuredDataset reads from or writes to S3, GCP, BigQuery, or any storage.
Let's understand how to convert a pandas DataFrame to a BigQuery table and vice-versa through an example.

Before writing DataFrame to a BigQuery table,

1. Create a [GCP account](https://cloud.google.com/docs/authentication/getting-started) and create a service account.
2. Create a project and add the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to your .bashrc file.
3. Create a dataset in your project.

+++ {"lines_to_next_cell": 0}

Import the dependencies.

```{code-cell}
import pandas as pd  # noqa: E402
from flytekit import task  # noqa: E402
from flytekit.types.structured import StructuredDataset  # noqa: E402
```

+++ {"lines_to_next_cell": 0}

Define a task that converts a pandas DataFrame to a BigQuery table.

```{code-cell}
@task
def pandas_to_bq() -> StructuredDataset:
    # create a pandas dataframe
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    # convert the dataframe to StructuredDataset
    return StructuredDataset(dataframe=df, uri="bq://sample-project-1-352610.sample_352610.test1")
```

:::{note}
The BigQuery uri's format is `bq://<project_name>.<dataset_name>.<table_name>`.
:::

+++ {"lines_to_next_cell": 0}

Define a task that converts the BigQuery table to a pandas DataFrame.

```{code-cell}
@task
def bq_to_pandas(sd: StructuredDataset) -> pd.DataFrame:
    # convert to pandas dataframe
    return sd.open(pd.DataFrame).all()
```

:::{note}
Flyte creates the table inside the dataset in the project upon BigQuery query execution.
:::

+++ {"lines_to_next_cell": 0}

Trigger the tasks locally.

```{code-cell}
if __name__ == "__main__":
    o1 = bq_to_pandas(sd=StructuredDataset(uri="bq://sample-project-1-352610.sample_352610.test1"))
    o2 = pandas_to_bq()
```

## NumPy Encoder and Decoder

`StructuredDataset` ships with an encoder and a decoder that handles the conversion of a Python value to a Flyte literal and vice-versa, respectively.
Let's understand how to write them by defining a NumPy encoder and decoder, which helps use NumPy array as a valid type within structured datasets.

+++ {"lines_to_next_cell": 0}

### NumPy Encoder

We extend `StructuredDatasetEncoder` and implement the `encode` function.
The `encode` function converts NumPy array to an intermediate format (parquet file format in this case).

```{code-cell}
class NumpyEncodingHandlers(StructuredDatasetEncoder):
    def encode(
        self,
        ctx: FlyteContext,
        structured_dataset: StructuredDataset,
        structured_dataset_type: StructuredDatasetType,
    ) -> literals.StructuredDataset:
        df = typing.cast(np.ndarray, structured_dataset.dataframe)
        name = ["col" + str(i) for i in range(len(df))]
        table = pa.Table.from_arrays(df, name)
        path = ctx.file_access.get_random_remote_directory()
        local_dir = ctx.file_access.get_random_local_directory()
        local_path = os.path.join(local_dir, f"{0:05}")
        pq.write_table(table, local_path)
        ctx.file_access.upload_directory(local_dir, path)
        return literals.StructuredDataset(
            uri=path,
            metadata=StructuredDatasetMetadata(structured_dataset_type=StructuredDatasetType(format=PARQUET)),
        )
```

+++ {"lines_to_next_cell": 0}

### NumPy Decoder

Next we extend `StructuredDatasetDecoder` and implement the `decode` function.
The `decode` function converts the parquet file to a `numpy.ndarray`.

```{code-cell}
class NumpyDecodingHandlers(StructuredDatasetDecoder):
    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
        current_task_metadata: StructuredDatasetMetadata,
    ) -> np.ndarray:
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.get_data(flyte_value.uri, local_dir, is_multipart=True)
        table = pq.read_table(local_dir)
        return table.to_pandas().to_numpy()
```

+++ {"lines_to_next_cell": 0}

### NumPy Renderer

Create a default renderer for numpy array, then flytekit will use this renderer to
display schema of numpy array on flyte deck.

```{code-cell}
class NumpyRenderer:
    """
    The schema of Numpy array are rendered as an HTML table.
    """

    def to_html(self, df: np.ndarray) -> str:
        assert isinstance(df, np.ndarray)
        name = ["col" + str(i) for i in range(len(df))]
        table = pa.Table.from_arrays(df, name)
        return pd.DataFrame(table.schema).to_html(index=False)
```

+++ {"lines_to_next_cell": 0}

Finally, we register the encoder, decoder, and renderer with the `StructuredDatasetTransformerEngine`.

```{code-cell}
StructuredDatasetTransformerEngine.register(NumpyEncodingHandlers(np.ndarray, None, PARQUET))
StructuredDatasetTransformerEngine.register(NumpyDecodingHandlers(np.ndarray, None, PARQUET))
StructuredDatasetTransformerEngine.register_renderer(np.ndarray, NumpyRenderer())
```

You can now use `numpy.ndarray` to deserialize the parquet file to NumPy and serialize a task's output (NumPy array) to a parquet file.

+++

Let's define a task to test the above functionality.
We open a structured dataset of type `numpy.ndarray` and serialize it again.

```{code-cell}
@task
def to_numpy(ds: Annotated[StructuredDataset, subset_cols]) -> Annotated[StructuredDataset, subset_cols, PARQUET]:
    numpy_array = ds.open(np.ndarray).all()
    return StructuredDataset(dataframe=numpy_array)
```

+++ {"lines_to_next_cell": 0}

Finally, we define two workflows that showcase how a `pandas.DataFrame` and `FlyteSchema` are accepted by the `StructuredDataset`.

```{code-cell}
@workflow
def pandas_compatibility_wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_df(a=a)
    ds = get_subset_df(df=df)  # noqa: shown for demonstration; users should use the same types between tasks
    return to_numpy(ds=ds)


@workflow
def schema_compatibility_wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_schema_df(a=a)
    ds = get_subset_df(df=df)  # noqa: shown for demonstration; users should use the same types between tasks
    return to_numpy(ds=ds)
```

+++ {"lines_to_next_cell": 0}

You can run the code locally as follows:

```{code-cell}
if __name__ == "__main__":
    numpy_array_one = pandas_compatibility_wf(a=42).open(np.ndarray).all()
    print(f"pandas DataFrame compatibility check output: {numpy_array_one}")
    numpy_array_two = schema_compatibility_wf(a=42).open(np.ndarray).all()
    print(f"Schema compatibility check output: {numpy_array_two}")
```
