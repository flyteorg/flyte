import tempfile

import pandas as pd
import polars as pl
from flytekitplugins.polars.sd_transformers import PolarsDataFrameRenderer
from typing_extensions import Annotated

from flytekit import kwtypes, task, workflow
from flytekit.types.structured.structured_dataset import PARQUET, StructuredDataset

subset_schema = Annotated[StructuredDataset, kwtypes(col2=str), PARQUET]
full_schema = Annotated[StructuredDataset, PARQUET]


def test_polars_workflow_subset():
    @task
    def generate() -> subset_schema:
        df = pl.DataFrame({"col1": [1, 3, 2], "col2": list("abc")})
        return StructuredDataset(dataframe=df)

    @task
    def consume(df: subset_schema) -> subset_schema:
        df = df.open(pl.DataFrame).all()

        assert df["col2"][0] == "a"
        assert df["col2"][1] == "b"
        assert df["col2"][2] == "c"

        return StructuredDataset(dataframe=df)

    @workflow
    def wf() -> subset_schema:
        return consume(df=generate())

    result = wf()
    assert result is not None


def test_polars_workflow_full():
    @task
    def generate() -> full_schema:
        df = pl.DataFrame({"col1": [1, 3, 2], "col2": list("abc")})
        return StructuredDataset(dataframe=df)

    @task
    def consume(df: full_schema) -> full_schema:
        df = df.open(pl.DataFrame).all()

        assert df["col1"][0] == 1
        assert df["col1"][1] == 3
        assert df["col1"][2] == 2
        assert df["col2"][0] == "a"
        assert df["col2"][1] == "b"
        assert df["col2"][2] == "c"

        return StructuredDataset(dataframe=df.sort("col1"))

    @workflow
    def wf() -> full_schema:
        return consume(df=generate())

    result = wf()
    assert result is not None


def test_polars_renderer():
    df = pl.DataFrame({"col1": [1, 3, 2], "col2": list("abc")})
    assert PolarsDataFrameRenderer().to_html(df) == pd.DataFrame(
        df.describe().transpose(), columns=df.describe().columns
    ).to_html(index=False)


def test_parquet_to_polars():
    data = {"name": ["Alice"], "age": [5]}

    @task
    def create_sd() -> StructuredDataset:
        df = pl.DataFrame(data=data)
        return StructuredDataset(dataframe=df)

    sd = create_sd()
    polars_df = sd.open(pl.DataFrame).all()
    assert pl.DataFrame(data).frame_equal(polars_df)

    tmp = tempfile.mktemp()
    pl.DataFrame(data).write_parquet(tmp)

    @task
    def t1(sd: StructuredDataset) -> pl.DataFrame:
        return sd.open(pl.DataFrame).all()

    sd = StructuredDataset(uri=tmp)
    assert t1(sd=sd).frame_equal(polars_df)

    @task
    def t2(sd: StructuredDataset) -> StructuredDataset:
        return StructuredDataset(dataframe=sd.open(pl.DataFrame).all())

    sd = StructuredDataset(uri=tmp)
    assert t2(sd=sd).open(pl.DataFrame).all().frame_equal(polars_df)
