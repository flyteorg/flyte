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

(structured_dataset)=

# StructuredDataset

```{eval-rst}
.. tags:: Basic, DataFrame
```
```{currentmodule} flytekit.types.structured
```

As with most type systems, Python has primitives, container types like maps and tuples, and support for user-defined structures.
However, while there’s a rich variety of dataframe classes (Pandas, Spark, Pandera, etc.), there’s no native Python type that
represents a dataframe in the abstract. This is the gap that the {py:class}`StructuredDataset` type is meant to fill.
It offers the following benefits:

- Eliminate boilerplate code you would otherwise need to write to serialize/deserialize from file objects into dataframe instances,
- Eliminate additional inputs/outputs that convey metadata around the format of the tabular data held in those files,
- Add flexibility around how dataframe files are loaded,
- Offer a range of dataframe specific functionality - enforce compatibility of different schemas
  (not only at compile time, but also runtime since type information is carried along in the literal),
   store third-party schema definitions, and potentially in the future, render sample data, provide summary statistics, etc.

## Usage

To use the `StructuredDataset` type, import `pandas` and define a task that returns a Pandas Dataframe.
Flytekit will detect the Pandas DataFrame return signature and convert the interface for the task to
the {py:class}`StructuredDataset` type.

## Example

This example demonstrates how to work with a structured dataset using Flyte entities.

```{note}
To use the `StructuredDataset` type, you only need to import `pandas`.
The other imports specified below are only necessary for this specific example.
```

To begin, import the dependencies for the example:

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

Define a task that returns a Pandas DataFrame.

```{code-cell}
@task
def generate_pandas_df(a: int) -> pd.DataFrame:
    return pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [a, 22], "Height": [160, 178]})
```

+++ {"lines_to_next_cell": 0}

Using this simplest form, however, the user is not able to set the additional dataframe information alluded to above,

- Column type information
- Serialized byte format
- Storage driver and location
- Additional third party schema information

This is by design as we wanted the default case to suffice for the majority of use-cases, and to require
as few changes to existing code as possible. Specifying these is simple, however, and relies on Python variable annotations,
which is designed explicitly to supplement types with arbitrary metadata.

## Column type information
If you want to extract a subset of actual columns of the dataframe and specify their types for type validation,
you can just specify the column names and their types in the structured dataset type annotation.

First, initialize column types you want to extract from the `StructuredDataset`.

```{code-cell}
all_cols = kwtypes(Name=str, Age=int, Height=int)
col = kwtypes(Age=int)
```

+++ {"lines_to_next_cell": 0}

Define a task that opens a structured dataset by calling `all()`.
When you invoke `all()` with ``pandas.DataFrame``, the Flyte engine downloads the parquet file on S3, and deserializes it to `pandas.DataFrame`.
Keep in mind that you can invoke ``open()`` with any dataframe type that's supported or added to structured dataset.
For instance, you can use ``pa.Table`` to convert the Pandas DataFrame to a PyArrow table.

```{code-cell}
@task
def get_subset_pandas_df(df: Annotated[StructuredDataset, all_cols]) -> Annotated[StructuredDataset, col]:
    df = df.open(pd.DataFrame).all()
    df = pd.concat([df, pd.DataFrame([[30]], columns=["Age"])])
    return StructuredDataset(dataframe=df)


@workflow
def simple_sd_wf(a: int = 19) -> Annotated[StructuredDataset, col]:
    pandas_df = generate_pandas_df(a=a)
    return get_subset_pandas_df(df=pandas_df)
```

+++ {"lines_to_next_cell": 0}

The code may result in runtime failures if the columns do not match.
The input ``df`` has ``Name``, ``Age`` and ``Height`` columns, whereas the output structured dataset will only have the ``Age`` column.

## Serialized byte format
You can use a custom serialization format to serialize your dataframes.
Here's how you can register the Pandas to CSV handler, which is already available,
and enable the CSV serialization by annotating the structured dataset with the CSV format:

```{code-cell}
from flytekit.types.structured import register_csv_handlers
from flytekit.types.structured.structured_dataset import CSV

register_csv_handlers()


@task
def pandas_to_csv(df: pd.DataFrame) -> Annotated[StructuredDataset, CSV]:
    return StructuredDataset(dataframe=df)


@workflow
def pandas_to_csv_wf() -> Annotated[StructuredDataset, CSV]:
    pandas_df = generate_pandas_df(a=19)
    return pandas_to_csv(df=pandas_df)
```

+++ {"lines_to_next_cell": 0}

## Storage driver and location
By default, the data will be written to the same place that all other pointer-types (FlyteFile, FlyteDirectory, etc.) are written to.
This is controlled by the output data prefix option in Flyte which is configurable on multiple levels.

That is to say, in the simple default case, Flytekit will,

- Look up the default format for say, Pandas dataframes,
- Look up the default storage location based on the raw output prefix setting,
- Use these two settings to select an encoder and invoke it.

So what's an encoder? To understand that, let's look into how the structured dataset plugin works.

## Inner workings of a structured dataset plugin

Two things need to happen with any dataframe instance when interacting with Flyte:

- Serialization/deserialization from/to the Python instance to bytes (in the format specified above).
- Transmission/retrieval of those bits to/from somewhere.

Each structured dataset plugin (called encoder or decoder) needs to perform both of these steps.
Flytekit decides which of the loaded plugins to invoke based on three attributes:

- The byte format
- The storage location
- The Python type in the task or workflow signature.

These three keys uniquely identify which encoder (used when converting a dataframe in Python memory to a Flyte value,
e.g. when a task finishes and returns a dataframe) or decoder (used when hydrating a dataframe in memory from a Flyte value,
e.g. when a task starts and has a dataframe input) to invoke.

However, it is awkward to require users to use `typing.Annotated` on every signature.
Therefore, Flytekit has a default byte-format for every Python dataframe type registered with flytekit.

## The `uri` argument

BigQuery `uri` allows you to load and retrieve data from cloud using the `uri` argument.
The `uri` comprises of the bucket name and the filename prefixed with `gs://`.
If you specify BigQuery `uri` for structured dataset, BigQuery creates a table in the location specified by the `uri`.
The `uri` in structured dataset reads from or writes to S3, GCP, BigQuery or any storage.

Before writing DataFrame to a BigQuery table,

1. Create a [GCP account](https://cloud.google.com/docs/authentication/getting-started) and create a service account.
2. Create a project and add the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to your `.bashrc` file.
3. Create a dataset in your project.

Here's how you can define a task that converts a pandas DataFrame to a BigQuery table:

```python
@task
def pandas_to_bq() -> StructuredDataset:
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    return StructuredDataset(dataframe=df, uri="gs://<BUCKET_NAME>/<FILE_NAME>")
```

Replace `BUCKET_NAME` with the name of your GCS bucket and `FILE_NAME` with the name of the file the dataframe should be copied to.

### Note that no format was specified in the structured dataset constructor, or in the signature. So how did the BigQuery encoder get invoked?
This is because the stock BigQuery encoder is loaded into Flytekit with an empty format.
The Flytekit `StructuredDatasetTransformerEngine` interprets that to mean that it is a generic encoder
(or decoder) and can work across formats, if a more specific format is not found.

And here's how you can define a task that converts the BigQuery table to a pandas DataFrame:

```python
@task
def bq_to_pandas(sd: StructuredDataset) -> pd.DataFrame:
   return sd.open(pd.DataFrame).all()
```

:::{note}
Flyte creates a table inside the dataset in the project upon BigQuery query execution.
:::

## How to return multiple dataframes from a task?
For instance, how would a task return say two dataframes:
- The first dataframe be written to BigQuery and serialized by one of their libraries,
- The second needs to be serialized to CSV and written at a specific location in GCS different from the generic pointer-data bucket

If you want the default behavior (which is itself configurable based on which plugins are loaded),
you can work just with your current raw dataframe classes.

```python
@task
def t1() -> typing.Tuple[StructuredDataset, StructuredDataset]:
   ...
   return StructuredDataset(df1, uri="bq://project:flyte.table"), \
          StructuredDataset(df2, uri="gs://auxiliary-bucket/data")
```

If you want to customize the Flyte interaction behavior, you'll need to wrap your dataframe in a `StructuredDataset` wrapper object.

## How to define a custom structured dataset plugin?

`StructuredDataset` ships with an encoder and a decoder that handles the conversion of a
Python value to a Flyte literal and vice-versa, respectively.
Here is a quick demo showcasing how one might build a NumPy encoder and decoder,
enabling the use of a 2D NumPy array as a valid type within structured datasets.

### NumPy encoder

Extend `StructuredDatasetEncoder` and implement the `encode` function.
The `encode` function converts NumPy array to an intermediate format (parquet file format in this case).

```{code-cell}
class NumpyEncodingHandler(StructuredDatasetEncoder):
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

### NumPy decoder

Extend {py:class}`StructuredDatasetDecoder` and implement the {py:meth}`~StructuredDatasetDecoder.decode` function.
The {py:meth}`~StructuredDatasetDecoder.decode` function converts the parquet file to a `numpy.ndarray`.

```{code-cell}
class NumpyDecodingHandler(StructuredDatasetDecoder):
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

### NumPy renderer

Create a default renderer for numpy array, then Flytekit will use this renderer to
display schema of NumPy array on the Flyte deck.

```{code-cell}
class NumpyRenderer:
    def to_html(self, df: np.ndarray) -> str:
        assert isinstance(df, np.ndarray)
        name = ["col" + str(i) for i in range(len(df))]
        table = pa.Table.from_arrays(df, name)
        return pd.DataFrame(table.schema).to_html(index=False)
```

+++ {"lines_to_next_cell": 0}

In the end, register the encoder, decoder and renderer with the `StructuredDatasetTransformerEngine`.
Specify the Python type you want to register this encoder with (`np.ndarray`),
the storage engine to register this against (if not specified, it is assumed to work for all the storage backends),
and the byte format, which in this case is `PARQUET`.

```{code-cell}
StructuredDatasetTransformerEngine.register(NumpyEncodingHandler(np.ndarray, None, PARQUET))
StructuredDatasetTransformerEngine.register(NumpyDecodingHandler(np.ndarray, None, PARQUET))
StructuredDatasetTransformerEngine.register_renderer(np.ndarray, NumpyRenderer())
```

+++ {"lines_to_next_cell": 0}

You can now use `numpy.ndarray` to deserialize the parquet file to NumPy and serialize a task's output (NumPy array) to a parquet file.

```{code-cell}
@task
def generate_pd_df_with_str() -> pd.DataFrame:
    return pd.DataFrame({"Name": ["Tom", "Joseph"]})


@task
def to_numpy(sd: StructuredDataset) -> Annotated[StructuredDataset, None, PARQUET]:
    numpy_array = sd.open(np.ndarray).all()
    return StructuredDataset(dataframe=numpy_array)


@workflow
def numpy_wf() -> Annotated[StructuredDataset, None, PARQUET]:
    return to_numpy(sd=generate_pd_df_with_str())
```

+++ {"lines_to_next_cell": 0}

:::{note}
`pyarrow` raises an `Expected bytes, got a 'int' object` error when the dataframe contains integers.
:::

You can run the code locally as follows:

```{code-cell}
if __name__ == "__main__":
    sd = simple_sd_wf()
    print(f"A simple Pandas dataframe workflow: {sd.open(pd.DataFrame).all()}")
    print(f"Using CSV as the serializer: {pandas_to_csv_wf().open(pd.DataFrame).all()}")
    print(f"NumPy encoder and decoder: {numpy_wf().open(np.ndarray).all()}")
```
