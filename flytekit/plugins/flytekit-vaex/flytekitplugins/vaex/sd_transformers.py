import os
import typing

from flytekit import FlyteContext, StructuredDatasetType, lazy_module
from flytekit.models import literals
from flytekit.models.literals import StructuredDatasetMetadata
from flytekit.types.structured.structured_dataset import (
    PARQUET,
    StructuredDataset,
    StructuredDatasetDecoder,
    StructuredDatasetEncoder,
    StructuredDatasetTransformerEngine,
)

pd = lazy_module("pandas")
vaex = lazy_module("vaex")


class VaexDataFrameToParquetEncodingHandler(StructuredDatasetEncoder):
    def __init__(self):
        super().__init__(vaex.dataframe.DataFrameLocal, None, PARQUET)

    def encode(
        self,
        ctx: FlyteContext,
        structured_dataset: StructuredDataset,
        structured_dataset_type: StructuredDatasetType,
    ) -> literals.StructuredDataset:
        df = typing.cast(vaex.dataframe.DataFrameLocal, structured_dataset.dataframe)
        local_dir = ctx.file_access.get_random_local_directory()
        local_path = os.path.join(local_dir, f"{0:05}")
        df.export_parquet(local_path)
        path = ctx.file_access.put_raw_data(local_dir)
        return literals.StructuredDataset(
            uri=path,
            metadata=StructuredDatasetMetadata(structured_dataset_type=structured_dataset_type),
        )


class ParquetToVaexDataFrameDecodingHandler(StructuredDatasetDecoder):
    def __init__(self):
        super().__init__(vaex.dataframe.DataFrameLocal, None, PARQUET)

    def decode(
        self,
        ctx: FlyteContext,
        flyte_value: literals.StructuredDataset,
        current_task_metadata: StructuredDatasetMetadata,
    ) -> vaex.dataframe.DataFrameLocal:
        local_dir = ctx.file_access.get_random_local_directory()
        ctx.file_access.get_data(flyte_value.uri, local_dir, is_multipart=True)
        path = f"{local_dir}/00000"
        if current_task_metadata.structured_dataset_type and current_task_metadata.structured_dataset_type.columns:
            columns = [c.name for c in current_task_metadata.structured_dataset_type.columns]
            return vaex.open(path)[columns]
        return vaex.open(path)


class VaexDataFrameRenderer:
    """
    Render a Vaex dataframe schema as an HTML table.
    """

    def to_html(self, df: vaex.dataframe.DataFrameLocal) -> str:
        assert isinstance(df, vaex.dataframe.DataFrameLocal)
        describe_df = df.describe()
        return pd.DataFrame(describe_df.transpose(), columns=describe_df.columns).to_html(index=False)


StructuredDatasetTransformerEngine.register(VaexDataFrameToParquetEncodingHandler())
StructuredDatasetTransformerEngine.register(ParquetToVaexDataFrameDecodingHandler())
StructuredDatasetTransformerEngine.register_renderer(vaex.dataframe.DataFrameLocal, VaexDataFrameRenderer())
