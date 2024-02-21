import sys
from dataclasses import dataclass
from typing import List

import pytest
from dataclasses_json import DataClassJsonMixin

from flytekit.core.task import task
from flytekit.core.workflow import workflow


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_dataclass():
    @dataclass
    class AppParams(DataClassJsonMixin):
        snapshotDate: str
        region: str
        preprocess: bool
        listKeys: List[str]

    @task
    def t1() -> AppParams:
        ap = AppParams(snapshotDate="4/5/2063", region="us-west-3", preprocess=False, listKeys=["a", "b"])
        return ap

    @workflow
    def wf() -> AppParams:
        return t1()

    res = wf()
    assert res.region == "us-west-3"
