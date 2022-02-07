"""
Using Structured Dataset
------------------

This example explains how an Structured Dataset is passed between tasks.
Flytekit makes it possible for users to directly return or accept a :py:class:`pandas.DataFrame`, which are automatically
converted into flyte's abstract representation of a Structured Dataset object

The structured dataset is a superset of flyte scheme. StructuredDataset Transformer can write dataframe to the BigQuery,
 S3, or any other storage by registering new structured dataset encoder/decoder.

"""
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
from flytekit.types.structured.structured_dataset import StructuredDataset, StructuredDatasetEncoder, \
    StructuredDatasetDecoder, FLYTE_DATASET_TRANSFORMER, PARQUET, S3, LOCAL

try:
    from typing import Annotated
except ImportError:
    from typing_extensions import Annotated


# %%
# Define columns type for schema and structuredDataset
superset_cols = kwtypes(Name=str, Age=int, Height=int)
subset_cols = kwtypes(Age=int)


# %%
# This task generates a pandas.DataFrame and returns it. The Dataframe itself will be serialized to an intermediate
# format like parquet before passing between tasks.
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


# %%
# This task shows an example of transforming a dataFrame, opening a Structured Dataset
# When we invoke all(), flyte engine will download parquet fine on s3, and deserialized to pandas.dataframe
#
# .. note::
# * Although the input type of task is StructuredDataset, it can also accepts FlyteSchema as input.
# * This may result in runtime failures if the columns do not match.
@task
def get_subset_df(df: Annotated[StructuredDataset, subset_cols]) -> Annotated[StructuredDataset, subset_cols]:
    df = df.open(pd.DataFrame).all()
    df = df.append(pd.DataFrame(data={"Age": [30]}))
    # You can also specify bigquery uri for StructuredDataset, then flytekit will write pd.dataframe to bigquery table
    return StructuredDataset(dataframe=df)


# %%
# Extend StructuredDatasetEncoder, implement the encode function, and register your concrete class with the
# FLYTE_DATASET_TRANSFORMER defined at this module level in order for the core flytekit type engine to handle
# dataframe libraries. This is the encoding interface, meaning it is used when there is a Python value that the
# flytekit type engine is trying to convert into a Flyte Literal.
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
        return literals.StructuredDataset(uri=path,
                                          metadata=StructuredDatasetMetadata(
                                              structured_dataset_type=StructuredDatasetType(format=PARQUET)))


# %%
# Extend StructuredDatasetDecoder, implement the decode function, and register your concrete class with the
# FLYTE_DATASET_TRANSFORMER. This is the decoding interface, meaning it is used when there is a Flyte Literal that the
# flytekit type engine is trying to convert into a python value, e.g. `pandas.Dataframe`
class NumpyDecodingHandlers(StructuredDatasetDecoder):
    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
    ) -> np.ndarray:
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.get_data(flyte_value.uri, local_dir, is_multipart=True)
        table = pq.read_table(local_dir)
        return table.to_pandas().to_numpy()


for protocol in [LOCAL, S3]:
    FLYTE_DATASET_TRANSFORMER.register_handler(NumpyEncodingHandlers(np.ndarray, protocol, PARQUET))
    FLYTE_DATASET_TRANSFORMER.register_handler(NumpyDecodingHandlers(np.ndarray, protocol, PARQUET))


# %%
# We just registered NumpyEncodingHandlers and NumpyDecodingHandlers, so we can deserialize the parquet file to numpy,
# and serialize task's output which is numpy array to s3.
@task
def to_numpy(ds: Annotated[StructuredDataset, subset_cols]) -> Annotated[StructuredDataset, subset_cols, PARQUET]:
    numpy_array = ds.open(np.ndarray).all()
    return StructuredDataset(dataframe=numpy_array)


# %%
# The workflow shows that passing DataFrame's between tasks is as simple as passing dataFrames in memory
@workflow
def wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_df(a=a)
    ds = get_subset_df(df=df)  # noqa, shown for demonstration, users should use the same types from one task to another
    return to_numpy(ds=ds)


@workflow
def schema_compatibility_wf(a: int) -> Annotated[StructuredDataset, subset_cols]:
    df = get_schema_df(a=a)
    ds = get_subset_df(df=df)  # noqa, shown for demonstration, users should use the same types from one task to another
    return to_numpy(ds=ds)


# %%
# The entire program can be run locally
if __name__ == "__main__":
    dataframe = wf(a=42).open(pd.DataFrame).all()
    print(f"Dataframe value:\n", dataframe)
