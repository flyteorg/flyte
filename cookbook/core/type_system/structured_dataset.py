"""
.. _structured_dataset_example:

Structured Dataset
------------------

Structured dataset is a superset of Flyte Schema.

The ``StructuredDataset`` Transformer can write a dataframe to BigQuery, s3, or any storage by registering new structured dataset encoder and decoder.

Flytekit makes it possible to return or accept a :py:class:`pandas.DataFrame` which is automatically
converted into Flyte's abstract representation of a structured dataset object.

This example explains how a structured dataset can be used with the Flyte entities.
"""

# %%
# Let's import the necessary dependencies.
import os
import typing

import numpy as np
import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq

from flytekit import task, workflow, kwtypes, FlyteContext, StructuredDatasetType
from flytekit.models import literals
from flytekit.models.literals import StructuredDatasetMetadata
from flytekit.types.schema import FlyteSchema
from flytekit.types.structured.structured_dataset import (
    StructuredDataset,
    StructuredDatasetEncoder,
    StructuredDatasetDecoder,
    StructuredDatasetTransformerEngine,
    PARQUET,
    S3,
    LOCAL,
)

try:
    from typing import Annotated
except ImportError:
    from typing_extensions import Annotated


# %%
# We define the columns types for schema and ``StructuredDataset``.
superset_cols = kwtypes(Name=str, Age=int, Height=int)
subset_cols = kwtypes(Age=int)


# %%
# We define two tasks — one returns a pandas DataFrame, and the other, a :py:class:`flytekit.types.schema.FlyteSchema`.
# The DataFrames themselves will be serialized to an intermediate format like parquet before sending them to the other tasks.
@task
def get_df(a: int) -> Annotated[pd.DataFrame, superset_cols]:
    """
    Generate a sample dataframe
    """
    return pd.DataFrame(
        {"Name": ["Tom", "Joseph"], "Age": [a, 22], "Height": [160, 178]}
    )


@task
def get_schema_df(a: int) -> FlyteSchema[superset_cols]:
    """
    Generate a sample dataframe
    """
    s = FlyteSchema()
    s.open().write(
        pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [a, 22], "Height": [160, 178]})
    )
    return s


# %%
# Next, we define a task that opens a structured dataset by calling ``all()``.
# When we invoke ``all()``, the Flyte engine will download parquet fine on s3, and deserializes it to ``pandas.dataframe``.
#
# .. note::
#   * Despite the input type of the task being ``StructuredDataset``, it can also accept FlyteSchema as input.
#   * The code may result in runtime failures if the columns do not match.
@task
def get_subset_df(
    df: Annotated[StructuredDataset, subset_cols]
) -> Annotated[StructuredDataset, subset_cols]:
    df = df.open(pd.DataFrame).all()
    df = pd.concat([df, pd.DataFrame([[30]], columns=["Age"])])
    # On specifying BigQuery uri for StructuredDataset, Flytekit will write pd.dataframe to a BigQuery table
    return StructuredDataset(dataframe=df)


# %%
# ``StructuredDataset`` ships with an encoder and a decoder that handles conversion of a Python value to Flyte literal and vice-versa, respectively.
# Let's understand how they need to be written by defining a Numpy encoder and decoder, which in turn helps use Numpy array as a valid type within structured datasets.


# %%
# Numpy Encoder
# =============
#
# We extend ``StructuredDatasetEncoder`` and implement the ``encode`` function.
# The ``encode`` function converts Numpy array to an intermediate format like parquet.
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
            metadata=StructuredDatasetMetadata(
                structured_dataset_type=StructuredDatasetType(format=PARQUET)
            ),
        )


# %%
# Numpy Decoder
# =============
#
# Next we extend ``StructuredDatasetDecoder`` and implement the ``decode`` function.
# The ``decode`` function converts the parquet file to a ``numpy.ndarray``.
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


# %%
# Finally, we register the encoder and decoder with the ``StructuredDatasetTransformerEngine``.
for protocol in [LOCAL, S3]:
    StructuredDatasetTransformerEngine.register(
        NumpyEncodingHandlers(np.ndarray, protocol, PARQUET)
    )
    StructuredDatasetTransformerEngine.register(
        NumpyDecodingHandlers(np.ndarray, protocol, PARQUET)
    )

# %%
# You can now use ``numpy.ndarray`` to deserialize the parquet file to Numpy and serialize a task's output (Numpy array) to a parquet file.

# %%
# Let's define a task to test the above functionality.
# We open a structured dataset of type ``numpy.ndarray`` and serialize it again.
@task
def to_numpy(
    ds: Annotated[StructuredDataset, subset_cols]
) -> Annotated[StructuredDataset, subset_cols, PARQUET]:
    numpy_array = ds.open(np.ndarray).all()
    return StructuredDataset(dataframe=numpy_array)


# %%
# Finally, we define two workflows that showcase how a ``pandas.DataFrame`` and ``FlyteSchema`` are accepted by the ``StructuredDataset``.
@workflow
def pandas_compatibility_wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_df(a=a)
    ds = get_subset_df(
        df=df
    )  # noqa: shown for demonstration; users should use the same types between tasks
    return to_numpy(ds=ds)


@workflow
def schema_compatibility_wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_schema_df(a=a)
    ds = get_subset_df(
        df=df
    )  # noqa: shown for demonstration; users should use the same types between tasks
    return to_numpy(ds=ds)


# %%
# You can run the code locally as follows:
if __name__ == "__main__":
    numpy_array_one = pandas_compatibility_wf(a=42).open(np.ndarray).all()
    print(f"pandas DataFrame compatibility check output: ", numpy_array_one)
    numpy_array_two = schema_compatibility_wf(a=42).open(np.ndarray).all()
    print(f"Schema compatibility check output: ", numpy_array_two)
