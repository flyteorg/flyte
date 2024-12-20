import os
import sys

import pytest

from flytekit import kwtypes, task, workflow
from flytekit.configuration import DefaultImages
from flytekit.core import context_manager
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task

# https://www.sqlitetutorial.net/sqlite-sample-database/
from flytekit.types.schema import FlyteSchema

ctx = context_manager.FlyteContextManager.current_context()
EXAMPLE_DB = os.path.join(os.path.dirname(os.path.realpath(__file__)), "chinook.zip")

# This task belongs to test_task_static but is intentionally here to help test tracking
tk = SQLite3Task(
    "test",
    query_template="select * from tracks",
    task_config=SQLite3Config(
        uri=EXAMPLE_DB,
        compressed=True,
    ),
)


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_task_static():
    assert tk.output_columns is None

    df = tk()
    assert df is not None


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_task_schema():
    # sqlite3_start

    sql_task = SQLite3Task(
        "test",
        query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
        task_config=SQLite3Config(
            uri=EXAMPLE_DB,
            compressed=True,
        ),
    )
    # sqlite3_end

    assert sql_task.output_columns is not None
    df = sql_task(limit=1)
    assert df is not None


@pytest.mark.skipif("pandas" not in sys.modules, reason="Pandas is not installed.")
def test_workflow():
    import pandas as pd

    @task
    def my_task(df: pd.DataFrame) -> int:
        return len(df[df.columns[0]])

    sql_task = SQLite3Task(
        "test",
        query_template="select * from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        task_config=SQLite3Config(
            uri=EXAMPLE_DB,
            compressed=True,
        ),
    )

    @workflow
    def wf(limit: int) -> int:
        return my_task(df=sql_task(limit=limit))

    assert wf(limit=5) == 5


def test_task_serialization():
    sql_task = SQLite3Task(
        "test",
        query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
        task_config=SQLite3Config(
            uri=EXAMPLE_DB,
            compressed=True,
        ),
    )

    tt = sql_task.serialize_to_model(sql_task.SERIALIZE_SETTINGS)

    assert tt.container.args == [
        "pyflyte-execute",
        "--inputs",
        "{{.input}}",
        "--output-prefix",
        "{{.outputPrefix}}",
        "--raw-output-data-prefix",
        "{{.rawOutputDataPrefix}}",
        "--resolver",
        "flytekit.core.python_customized_container_task.default_task_template_resolver",
        "--",
        "{{.taskTemplatePath}}",
        "flytekit.extras.sqlite3.task.SQLite3TaskExecutor",
    ]

    assert tt.custom["query_template"] == "select TrackId, Name from tracks limit {{.inputs.limit}}"
    assert tt.container.image == DefaultImages.default_image()

    image = "xyz.io/docker2:latest"
    sql_task._container_image = image
    tt = sql_task.serialize_to_model(sql_task.SERIALIZE_SETTINGS)
    assert tt.container.image == image


@pytest.mark.parametrize(
    "query_template, expected_query",
    [
        (
            """
select *
from tracks
limit {{.inputs.limit}}""",
            "select * from tracks limit {{.inputs.limit}}",
        ),
        (
            """ \
select * \
from tracks \
limit {{.inputs.limit}}""",
            "select * from tracks limit {{.inputs.limit}}",
        ),
        ("select * from abc", "select * from abc"),
    ],
)
def test_query_sanitization(query_template, expected_query):
    sql_task = SQLite3Task(
        "test",
        query_template=query_template,
        inputs=kwtypes(limit=int),
        task_config=SQLite3Config(
            uri=EXAMPLE_DB,
            compressed=True,
        ),
    )
    assert sql_task.query_template == expected_query
