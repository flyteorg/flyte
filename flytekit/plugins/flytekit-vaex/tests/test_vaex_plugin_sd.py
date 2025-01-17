import pandas as pd
import vaex
from flytekitplugins.vaex.sd_transformers import VaexDataFrameRenderer
from typing_extensions import Annotated

from flytekit import kwtypes, task, workflow
from flytekit.types.structured.structured_dataset import PARQUET, StructuredDataset

full_schema = Annotated[StructuredDataset, kwtypes(x=int, y=str), PARQUET]
subset_schema = Annotated[StructuredDataset, kwtypes(y=str), PARQUET]
vaex_df = vaex.from_dict(dict(x=[1, 3, 2], y=["a", "b", "c"]))


def test_vaex_workflow_subset():
    @task
    def generate() -> subset_schema:
        return StructuredDataset(dataframe=vaex_df)

    @task
    def consume(df: subset_schema) -> subset_schema:
        subset_df = df.open(vaex.dataframe.DataFrameLocal).all()
        assert subset_df.column_names == ["y"]
        coly = subset_df.y.values.tolist()
        assert coly[0] == "a"
        assert coly[1] == "b"
        assert coly[2] == "c"
        return StructuredDataset(dataframe=subset_df)

    @workflow
    def wf() -> subset_schema:
        return consume(df=generate())

    result = wf()
    assert result is not None


def test_vaex_workflow_full():
    @task
    def generate() -> full_schema:
        return StructuredDataset(dataframe=vaex_df)

    @task
    def consume(df: full_schema) -> full_schema:
        full_df = df.open(vaex.dataframe.DataFrameLocal).all()
        assert full_df.column_names == ["x", "y"]
        colx = full_df.x.values.tolist()
        coly = full_df.y.values.tolist()

        assert colx[0] == 1
        assert colx[1] == 3
        assert colx[2] == 2
        assert coly[0] == "a"
        assert coly[1] == "b"
        assert coly[2] == "c"

        return StructuredDataset(dataframe=full_df.sort("x"))

    @workflow
    def wf() -> full_schema:
        return consume(df=generate())

    result = wf()
    assert result is not None


def test_vaex_renderer():
    vaex_df = vaex.from_dict(dict(x=[1, 3, 2], y=["a", "b", "c"]))
    assert VaexDataFrameRenderer().to_html(vaex_df) == pd.DataFrame(
        vaex_df.describe().transpose(), columns=vaex_df.columns
    ).to_html(index=False)


def test_vaex_type():
    @task
    def create_vaex_df() -> vaex.dataframe.DataFrameLocal:
        return vaex.from_pandas(pd.DataFrame(data={"column_1": [-1, 2, -3], "column_2": [1.5, 2.21, 3.9]}))

    @task
    def consume_vaex_df(vaex_df: vaex.dataframe.DataFrameLocal) -> vaex.dataframe.DataFrameLocal:
        df_negative = vaex_df[vaex_df.column_1 < 0]
        return df_negative

    @workflow
    def wf() -> vaex.dataframe.DataFrameLocal:
        return consume_vaex_df(vaex_df=create_vaex_df())

    result = wf()
    assert isinstance(result, vaex.dataframe.DataFrameLocal)
    assert len(result) == 2
