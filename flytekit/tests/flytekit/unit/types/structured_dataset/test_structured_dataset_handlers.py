import typing

import mock
import pyarrow as pa
import pytest

from flytekit.core import context_manager
from flytekit.core.base_task import kwtypes
from flytekit.models.literals import StructuredDatasetMetadata
from flytekit.models.types import StructuredDatasetType
from flytekit.types.structured import basic_dfs
from flytekit.types.structured.structured_dataset import (
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetTransformerEngine,
)

pd = pytest.importorskip("pandas")
my_cols = kwtypes(w=typing.Dict[str, typing.Dict[str, int]], x=typing.List[typing.List[int]], y=int, z=str)
fields = [("some_int", pa.int32()), ("some_string", pa.string())]
arrow_schema = pa.schema(fields)


def test_pandas():
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    encoder = basic_dfs.PandasToParquetEncodingHandler()
    decoder = basic_dfs.ParquetToPandasDecodingHandler()

    ctx = context_manager.FlyteContextManager.current_context()
    sd = StructuredDataset(dataframe=df)
    sd_type = StructuredDatasetType(format="parquet")
    sd_lit = encoder.encode(ctx, sd, sd_type)

    df2 = decoder.decode(ctx, sd_lit, StructuredDatasetMetadata(sd_type))
    assert df.equals(df2)


def test_csv():
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    encoder = basic_dfs.PandasToCSVEncodingHandler()
    decoder = basic_dfs.CSVToPandasDecodingHandler()

    ctx = context_manager.FlyteContextManager.current_context()
    sd = StructuredDataset(dataframe=df)
    sd_type = StructuredDatasetType(format="csv")
    sd_lit = encoder.encode(ctx, sd, sd_type)

    df2 = decoder.decode(ctx, sd_lit, StructuredDatasetMetadata(sd_type))
    assert df.equals(df2)


@mock.patch("pandas.DataFrame.to_parquet")
@mock.patch("pandas.read_parquet")
@mock.patch("flytekit.types.structured.basic_dfs.get_fsspec_storage_options")
def test_pandas_to_parquet_azure_storage_options(mock_get_fsspec_storage_options, mock_read_parquet, mock_to_parquet):
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    encoder = basic_dfs.PandasToParquetEncodingHandler()
    decoder = basic_dfs.ParquetToPandasDecodingHandler()

    mock_get_fsspec_storage_options.return_value = {"account_name": "accountname_from_storage_options"}
    ctx = context_manager.FlyteContextManager.current_context()
    sd = StructuredDataset(dataframe=df, uri="abfs://container/parquet_df")
    sd_type = StructuredDatasetType(format="parquet")
    sd_lit = encoder.encode(ctx, sd, sd_type)
    mock_to_parquet.assert_called_once()
    write_storage_options = mock_to_parquet.call_args.kwargs["storage_options"]
    assert write_storage_options == {"account_name": "accountname_from_storage_options"}

    decoder.decode(ctx, sd_lit, StructuredDatasetMetadata(sd_type))
    mock_read_parquet.assert_called_once()
    read_storage_options = mock_read_parquet.call_args.kwargs["storage_options"]
    read_storage_options == {"account_name": "accountname_from_storage_options"}


@mock.patch("pandas.DataFrame.to_csv")
@mock.patch("pandas.read_csv")
@mock.patch("flytekit.types.structured.basic_dfs.get_fsspec_storage_options")
def test_pandas_to_csv_azure_storage_options(mock_get_fsspec_storage_options, mock_read_parquet, mock_to_parquet):
    df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    encoder = basic_dfs.PandasToCSVEncodingHandler()
    decoder = basic_dfs.CSVToPandasDecodingHandler()

    mock_get_fsspec_storage_options.return_value = {"account_name": "accountname_from_storage_options"}
    ctx = context_manager.FlyteContextManager.current_context()
    sd = StructuredDataset(dataframe=df, uri="abfs://container/csv_df")
    sd_type = StructuredDatasetType(format="csv")
    sd_lit = encoder.encode(ctx, sd, sd_type)
    mock_to_parquet.assert_called_once()
    write_storage_options = mock_to_parquet.call_args.kwargs["storage_options"]
    assert write_storage_options == {"account_name": "accountname_from_storage_options"}

    decoder.decode(ctx, sd_lit, StructuredDatasetMetadata(sd_type))
    mock_read_parquet.assert_called_once()
    read_storage_options = mock_read_parquet.call_args.kwargs["storage_options"]
    read_storage_options == {"account_name": "accountname_from_storage_options"}


def test_base_isnt_instantiable():
    with pytest.raises(TypeError):
        StructuredDatasetEncoder(pd.DataFrame, "", "")

    with pytest.raises(TypeError):
        StructuredDatasetDecoder(pd.DataFrame, "", "")


def test_arrow():
    encoder = basic_dfs.ArrowToParquetEncodingHandler()
    decoder = basic_dfs.ParquetToArrowDecodingHandler()
    assert encoder.protocol is None
    assert decoder.protocol is None
    assert encoder.python_type is decoder.python_type
    d = StructuredDatasetTransformerEngine.DECODERS[encoder.python_type]["fsspec"]["parquet"]
    assert d is not None
