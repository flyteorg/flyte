import os
import sys

import pytest
from flytekitplugins.papermill import NotebookTask
from flytekitplugins.spark import Spark

from flytekit import kwtypes
from flytekit.types.schema import FlyteSchema


def _get_nb_path(name: str, suffix: str = "", abs: bool = True, ext: str = ".ipynb") -> str:
    """
    Creates a correct path no matter where the test is run from
    """
    _local_path = os.path.dirname(__file__)
    path = f"{_local_path}/testdata/{name}{suffix}{ext}"
    return os.path.abspath(path) if abs else path


def test_notebook_task_simple():
    if sys.version_info[:2] >= (3, 11):
        pytest.xfail(
            "pyspark serialization issues will be addressed in version 3.4.0, issue tracking at: "
            "https://issues.apache.org/jira/browse/SPARK-41125"
        )

    nb_name = "nb-spark"
    nb = NotebookTask(
        name="test",
        notebook_path=_get_nb_path(nb_name, abs=False),
        outputs=kwtypes(df=FlyteSchema[kwtypes(name=str, age=int)]),
        task_config=Spark(spark_conf={"x": "y"}),
    )
    n, out, render = nb.execute()
    assert nb.python_interface.outputs.keys() == {"df", "out_nb", "out_rendered_nb"}
    assert nb.output_notebook_path == out == _get_nb_path(nb_name, suffix="-out")
    assert nb.rendered_output_path == render == _get_nb_path(nb_name, suffix="-out", ext=".html")
