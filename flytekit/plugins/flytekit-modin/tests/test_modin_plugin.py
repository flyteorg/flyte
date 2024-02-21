import os

import ray
from flytekitplugins.modin import schema  # noqa F401
from modin import pandas as pd

from flytekit import task, workflow

os.environ["MODIN_ENGINE"] = "ray"
if not ray.is_initialized():
    ray.init(_plasma_directory="/tmp")  # setting to disable out of core in Ray


def test_modin_workflow():
    @task
    def generate() -> pd.DataFrame:
        df = pd.DataFrame({"col1": [1, 2, 3], "col2": list("abc")})
        return df

    @task
    def consume(df: pd.DataFrame) -> pd.DataFrame:
        df["col3"] = df["col1"].astype(str) + df["col2"]
        return df

    @workflow
    def wf() -> pd.DataFrame:
        return consume(df=generate())

    result = wf()
    assert result is not None
    assert isinstance(result, pd.DataFrame)
