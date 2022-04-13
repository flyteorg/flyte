"""
.. _integrations_sql_sqlite3:

Sqlite3
#######

The following example shows how you can write SQLite3 queries using the SQLite3Task, which is bundled as part of the
core flytekit. Since SQL Queries are portable across workflows and Flyte installations (as long as the data exists),
this task will always run with a pre-built container, specifically the `flytekit container <https://github.com/flyteorg/flytekit/blob/v0.19.0/Dockerfile.py38>`__
itself. Therefore, users are not required to build a container for SQLite3. You can simply implement the task - register and
execute it immediately.

In some cases local execution is not possible - e.g. Snowflake. But for SQLlite3 local execution is also supported.
"""
import pandas
from flytekit import kwtypes, task, workflow
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task

# %%
# SQLite3 queries in flyte produce a Schema output.
# https://www.sqlitetutorial.net/sqlite-sample-database/
from flytekit.types.schema import FlyteSchema

EXAMPLE_DB = "https://www.sqlitetutorial.net/wp-content/uploads/2018/03/chinook.zip"

# %%
# the task is declared as a regular task. Alternatively it can be declared within a workflow just at the point of using
# it (example later)
sql_task = SQLite3Task(
    name="cookbook.sqlite3.sample",
    query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
    inputs=kwtypes(limit=int),
    output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
    task_config=SQLite3Config(uri=EXAMPLE_DB, compressed=True),
)


# %%
# As described elsewhere FlyteSchemas can be easily be received as pandas Dataframe and Flyte will autoconvert them
@task
def print_and_count_columns(df: pandas.DataFrame) -> int:
    return len(df[df.columns[0]])


# %%
# The task can be used normally in the workflow, passing the declared inputs
@workflow
def wf() -> int:
    return print_and_count_columns(df=sql_task(limit=100))


# %%
# It can also be executed locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running main {wf()}")


# %%
# As mentioned earlier it is possible to also write the SQL Task inline as follows
@workflow
def query_wf() -> int:
    df = SQLite3Task(
        name="cookbook.sqlite3.sample_inline",
        query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
        task_config=SQLite3Config(uri=EXAMPLE_DB, compressed=True),
    )(limit=100)
    return print_and_count_columns(df=df)


# %%
# It can also be executed locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running main {query_wf()}")
