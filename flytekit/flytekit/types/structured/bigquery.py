import re
import typing

from google.cloud import bigquery, bigquery_storage
from google.cloud.bigquery_storage_v1 import types

from flytekit import FlyteContext, lazy_module
from flytekit.models import literals
from flytekit.models.types import StructuredDatasetType
from flytekit.types.structured.structured_dataset import (
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetMetadata,
)

if typing.TYPE_CHECKING:
    import pandas as pd
    import pyarrow as pa
else:
    pd = lazy_module("pandas")
    pa = lazy_module("pyarrow")

BIGQUERY = "bq"


def _write_to_bq(structured_dataset: StructuredDataset):
    table_id = typing.cast(str, structured_dataset.uri).split("://", 1)[1].replace(":", ".")
    client = bigquery.Client()
    df = structured_dataset.dataframe
    if isinstance(df, pa.Table):
        df = df.to_pandas()
    client.load_table_from_dataframe(df, table_id)


def _read_from_bq(
    flyte_value: literals.StructuredDataset, current_task_metadata: StructuredDatasetMetadata
) -> pd.DataFrame:
    path = flyte_value.uri
    _, project_id, dataset_id, table_id = re.split("\\.|://|:", path)
    client = bigquery_storage.BigQueryReadClient()
    table = f"projects/{project_id}/datasets/{dataset_id}/tables/{table_id}"
    parent = "projects/{}".format(project_id)

    read_options = None
    if current_task_metadata.structured_dataset_type and current_task_metadata.structured_dataset_type.columns:
        columns = [c.name for c in current_task_metadata.structured_dataset_type.columns]
        read_options = types.ReadSession.TableReadOptions(selected_fields=columns)

    requested_session = types.ReadSession(table=table, data_format=types.DataFormat.ARROW, read_options=read_options)
    read_session = client.create_read_session(parent=parent, read_session=requested_session)

    stream = read_session.streams[0]
    reader = client.read_rows(stream.name)
    frames = []
    for message in reader.rows().pages:
        frames.append(message.to_dataframe())
    return pd.concat(frames)


class PandasToBQEncodingHandlers(StructuredDatasetEncoder):
    def __init__(self):
        super().__init__(pd.DataFrame, BIGQUERY, supported_format="")

    def encode(
        self,
        ctx: FlyteContext,
        structured_dataset: StructuredDataset,
        structured_dataset_type: StructuredDatasetType,
    ) -> literals.StructuredDataset:
        _write_to_bq(structured_dataset)
        return literals.StructuredDataset(
            uri=typing.cast(str, structured_dataset.uri), metadata=StructuredDatasetMetadata(structured_dataset_type)
        )


class BQToPandasDecodingHandler(StructuredDatasetDecoder):
    def __init__(self):
        super().__init__(pd.DataFrame, BIGQUERY, supported_format="")

    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
        current_task_metadata: StructuredDatasetMetadata,
    ) -> pd.DataFrame:
        return _read_from_bq(flyte_value, current_task_metadata)


class ArrowToBQEncodingHandlers(StructuredDatasetEncoder):
    def __init__(self):
        super().__init__(pa.Table, BIGQUERY, supported_format="")

    def encode(
        self,
        ctx: FlyteContext,
        structured_dataset: StructuredDataset,
        structured_dataset_type: StructuredDatasetType,
    ) -> literals.StructuredDataset:
        _write_to_bq(structured_dataset)
        return literals.StructuredDataset(
            uri=typing.cast(str, structured_dataset.uri), metadata=StructuredDatasetMetadata(structured_dataset_type)
        )


class BQToArrowDecodingHandler(StructuredDatasetDecoder):
    def __init__(self):
        super().__init__(pa.Table, BIGQUERY, supported_format="")

    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
        current_task_metadata: StructuredDatasetMetadata,
    ) -> pa.Table:
        return pa.Table.from_pandas(_read_from_bq(flyte_value, current_task_metadata))
