import os
import typing

import numpy as np
import pyarrow as pa
import pyarrow.parquet as pq
import pytest
from typing_extensions import Annotated

from flytekit import FlyteContext, FlyteContextManager, kwtypes, task, workflow
from flytekit.models import literals
from flytekit.models.literals import StructuredDatasetMetadata
from flytekit.models.types import StructuredDatasetType
from flytekit.types.structured.basic_dfs import CSVToPandasDecodingHandler, PandasToCSVEncodingHandler
from flytekit.types.structured.structured_dataset import (
    CSV,
    DF,
    PARQUET,
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetTransformerEngine,
)

pd = pytest.importorskip("pandas")

PANDAS_PATH = FlyteContextManager.current_context().file_access.get_random_local_directory()
NUMPY_PATH = FlyteContextManager.current_context().file_access.get_random_local_directory()
BQ_PATH = "bq://flyte-dataset:flyte.table"

my_cols = kwtypes(Name=str, Age=int)
fields = [("Name", pa.string()), ("Age", pa.int32())]
arrow_schema = pa.schema(fields)
pd_df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})


class MockBQEncodingHandlers(StructuredDatasetEncoder):
    def __init__(self):
        super().__init__(pd.DataFrame, "bq", "")

    def encode(
        self,
        ctx: FlyteContext,
        structured_dataset: StructuredDataset,
        structured_dataset_type: StructuredDatasetType,
    ) -> literals.StructuredDataset:
        return literals.StructuredDataset(
            uri="bq://bucket/key", metadata=StructuredDatasetMetadata(structured_dataset_type)
        )


class MockBQDecodingHandlers(StructuredDatasetDecoder):
    def __init__(self):
        super().__init__(pd.DataFrame, "bq", "")

    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
        current_task_metadata: StructuredDatasetMetadata,
    ) -> pd.DataFrame:
        return pd_df


StructuredDatasetTransformerEngine.register(MockBQEncodingHandlers(), False, True)
StructuredDatasetTransformerEngine.register(MockBQDecodingHandlers(), False, True)


class NumpyRenderer:
    """
    The Polars DataFrame summary statistics are rendered as an HTML table.
    """

    def to_html(self, array: np.ndarray) -> str:
        return pd.DataFrame(array).describe().to_html()


@pytest.fixture(autouse=True)
def numpy_type():
    class NumpyEncodingHandlers(StructuredDatasetEncoder):
        def encode(
            self,
            ctx: FlyteContext,
            structured_dataset: StructuredDataset,
            structured_dataset_type: StructuredDatasetType,
        ) -> literals.StructuredDataset:
            path = typing.cast(str, structured_dataset.uri)
            if not path:
                path = ctx.file_access.join(
                    ctx.file_access.raw_output_prefix,
                    ctx.file_access.get_random_string(),
                )
            df = typing.cast(np.ndarray, structured_dataset.dataframe)
            name = ["col" + str(i) for i in range(len(df))]
            table = pa.Table.from_arrays(df, name)
            local_dir = ctx.file_access.get_random_local_directory()
            local_path = os.path.join(local_dir, f"{0:05}")
            pq.write_table(table, local_path)
            ctx.file_access.upload_directory(local_dir, path)
            structured_dataset_type.format = PARQUET
            return literals.StructuredDataset(uri=path, metadata=StructuredDatasetMetadata(structured_dataset_type))

    class NumpyDecodingHandlers(StructuredDatasetDecoder):
        def decode(
            self,
            ctx: FlyteContext,
            flyte_value: literals.StructuredDataset,
            current_task_metadata: StructuredDatasetMetadata,
        ) -> typing.Union[DF, typing.Generator[DF, None, None]]:
            path = flyte_value.uri
            local_dir = ctx.file_access.get_random_local_directory()
            ctx.file_access.get_data(path, local_dir, is_multipart=True)
            table = pq.read_table(local_dir)
            return table.to_pandas().to_numpy()

    StructuredDatasetTransformerEngine.register(NumpyEncodingHandlers(np.ndarray))
    StructuredDatasetTransformerEngine.register(NumpyDecodingHandlers(np.ndarray))
    StructuredDatasetTransformerEngine.register_renderer(np.ndarray, NumpyRenderer())


@task
def t1(dataframe: pd.DataFrame) -> Annotated[pd.DataFrame, my_cols]:
    # S3 (parquet) -> Pandas -> S3 (parquet) default behaviour
    return dataframe


@task
def t1a(dataframe: pd.DataFrame) -> Annotated[StructuredDataset, my_cols, PARQUET]:
    # S3 (parquet) -> Pandas -> S3 (parquet)
    return StructuredDataset(dataframe=dataframe, uri=PANDAS_PATH)


@task
def t2(dataframe: pd.DataFrame) -> Annotated[pd.DataFrame, arrow_schema]:
    # S3 (parquet) -> Pandas -> S3 (parquet)
    return dataframe


@task
def t3(dataset: Annotated[StructuredDataset, my_cols]) -> Annotated[StructuredDataset, my_cols]:
    # s3 (parquet) -> pandas -> s3 (parquet)
    print(dataset.open(pd.DataFrame).all())
    # In the example, we download dataset when we open it.
    # Here we won't upload anything, since we're returning just the input object.
    return dataset


@task
def t3a(dataset: Annotated[StructuredDataset, my_cols]) -> Annotated[StructuredDataset, my_cols]:
    # This task will not do anything - no uploading, no downloading
    return dataset


@task
def t4(dataset: Annotated[StructuredDataset, my_cols]) -> pd.DataFrame:
    # s3 (parquet) -> pandas -> s3 (parquet)
    return dataset.open(pd.DataFrame).all()


@task
def t5(dataframe: pd.DataFrame) -> Annotated[StructuredDataset, my_cols]:
    # s3 (parquet) -> pandas -> bq
    return StructuredDataset(dataframe=dataframe, uri=BQ_PATH)


@task
def t6(dataset: Annotated[StructuredDataset, my_cols]) -> pd.DataFrame:
    # bq -> pandas -> s3 (parquet)
    df = dataset.open(pd.DataFrame).all()
    return df


@task
def t7(
    df1: pd.DataFrame, df2: pd.DataFrame
) -> (Annotated[StructuredDataset, my_cols], Annotated[StructuredDataset, my_cols]):
    # df1: pandas -> bq
    # df2: pandas -> s3 (parquet)
    return StructuredDataset(dataframe=df1, uri=BQ_PATH), StructuredDataset(dataframe=df2)


@task
def t8(dataframe: pa.Table) -> Annotated[StructuredDataset, my_cols]:
    # Arrow table -> s3 (parquet)
    print(dataframe.columns)
    return StructuredDataset(dataframe=dataframe)


@task
def t8a(dataframe: pa.Table) -> pa.Table:
    # Arrow table -> s3 (parquet)
    print(dataframe.columns)
    return dataframe


@task
def t9(dataframe: np.ndarray) -> Annotated[StructuredDataset, my_cols]:
    # numpy -> Arrow table -> s3 (parquet)
    return StructuredDataset(dataframe=dataframe, uri=NUMPY_PATH)


@task
def t10(dataset: Annotated[StructuredDataset, my_cols]) -> np.ndarray:
    # s3 (parquet) -> Arrow table -> numpy
    np_array = dataset.open(np.ndarray).all()
    return np_array


StructuredDatasetTransformerEngine.register(PandasToCSVEncodingHandler())
StructuredDatasetTransformerEngine.register(CSVToPandasDecodingHandler())


@task
def t11(dataframe: pd.DataFrame) -> Annotated[StructuredDataset, CSV]:
    # pandas -> csv
    return StructuredDataset(dataframe=dataframe, uri=PANDAS_PATH)


@task
def t12(dataset: Annotated[StructuredDataset, my_cols]) -> pd.DataFrame:
    # csv -> pandas
    df = dataset.open(pd.DataFrame).all()
    return df


@task
def generate_pandas() -> pd.DataFrame:
    return pd_df


@task
def generate_numpy() -> np.ndarray:
    return np.array([[1, 2], [4, 5]])


@task
def generate_arrow() -> pa.Table:
    return pa.Table.from_pandas(pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]}))


@workflow()
def wf():
    df = generate_pandas()
    np_array = generate_numpy()
    arrow_df = generate_arrow()
    t1(dataframe=df)
    t1a(dataframe=df)
    t2(dataframe=df)
    t3(dataset=StructuredDataset(uri=PANDAS_PATH))
    t3a(dataset=StructuredDataset(uri=PANDAS_PATH))
    t4(dataset=StructuredDataset(uri=PANDAS_PATH))
    t5(dataframe=df)
    t6(dataset=StructuredDataset(uri=BQ_PATH))
    t7(df1=df, df2=df)
    t8(dataframe=arrow_df)
    t8a(dataframe=arrow_df)
    t9(dataframe=np_array)
    t10(dataset=StructuredDataset(uri=NUMPY_PATH))
    t11(dataframe=df)
    t12(dataset=StructuredDataset(uri=PANDAS_PATH))


def test_structured_dataset_wf():
    wf()
