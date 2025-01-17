import sys
import typing

import pyarrow as pa
import pytest
from typing_extensions import Annotated

from flytekit import kwtypes, task


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_structured_dataset_wf():
    import pandas as pd

    cols = kwtypes(Name=str, Age=int)
    subset_cols = kwtypes(Name=str)

    @task
    def t1(
        df1: Annotated[pd.DataFrame, cols], df2: Annotated[pa.Table, cols]
    ) -> typing.Tuple[Annotated[pd.DataFrame, subset_cols], Annotated[pa.Table, subset_cols]]:
        return df1, df2

    pd_df = pd.DataFrame({"Name": ["Tom", "Joseph"], "Age": [20, 22]})
    pa_df = pa.Table.from_pandas(pd_df)

    subset_pd_df = pd.DataFrame({"Name": ["Tom", "Joseph"]})
    subset_pa_df = pa.Table.from_pandas(subset_pd_df)

    df1, df2 = t1(df1=pd_df, df2=pa_df)
    assert df1.equals(subset_pd_df)
    assert df2.equals(subset_pa_df)
