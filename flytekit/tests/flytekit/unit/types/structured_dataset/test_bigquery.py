import mock
import pytest
from typing_extensions import Annotated

from flytekit import StructuredDataset, kwtypes, task, workflow

pd = pytest.importorskip("pandas")

pd_df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
my_cols = kwtypes(Name=str, Age=int)


@task
def gen_df() -> Annotated[pd.DataFrame, my_cols, "parquet"]:
    return pd_df


@task
def t1(df: pd.DataFrame) -> Annotated[StructuredDataset, my_cols]:
    return StructuredDataset(dataframe=df, uri="bq://project:flyte.table")


@task
def t2(sd: Annotated[StructuredDataset, my_cols]) -> pd.DataFrame:
    return sd.open(pd.DataFrame).all()


@workflow
def wf() -> pd.DataFrame:
    df = gen_df()
    sd = t1(df=df)
    return t2(sd=sd)


@mock.patch("google.cloud.bigquery.Client")
@mock.patch("google.cloud.bigquery_storage.BigQueryReadClient")
@mock.patch("google.cloud.bigquery_storage_v1.reader.ReadRowsStream")
def test_bq_wf(mock_read_rows_stream, mock_bigquery_read_client, mock_client):
    class mock_pages:
        def to_dataframe(self):
            return pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})

    class mock_rows:
        pages = [mock_pages()]

    mock_client.load_table_from_dataframe.return_value = None
    mock_read_rows_stream.rows.return_value = mock_rows
    mock_bigquery_read_client.read_rows.return_value = mock_read_rows_stream
    mock_bigquery_read_client.return_value = mock_bigquery_read_client

    assert wf().equals(pd_df)
