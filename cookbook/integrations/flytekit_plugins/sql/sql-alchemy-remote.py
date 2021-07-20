"""
SQLAlchemy
----------

SQLAlchemy is the Python SQL toolkit and Object Relational Mapper that gives application developers the full power and flexibility of SQL.

That being said, Flyte provides an easy-to-use interface to utilize SQLAlchemy to connect to various SQL Databases.

The SQLAlchemy task will run with a pre-built container, and thus users needn't build one.
"""

# %%
# Let's import the libraries.
import pandas
from flytekit import kwtypes, task, workflow
from flytekitplugins.sqlalchemy import SQLAlchemyConfig, SQLAlchemyTask


# %%
# We define an SQLAlchemyTask to fetch limited records from a table. Finally, we return the length of the returned DataFrame.
#
# .. note::
#
#   The output of SQLAlchemyTask is a :py:class:`~flytekit.types.schema.FlyteSchema` by default.
@task
def get_length(df: pandas.DataFrame) -> int:
    return len(df)


sql_task = SQLAlchemyTask(
    name="sqlalchemy_task",
    query_template="select * from <table> limit {{.inputs.limit}}",
    inputs=kwtypes(limit=int),
    task_config=SQLAlchemyConfig(uri="<uri>"),
)


@workflow
def my_wf(limit: int) -> int:
    return get_length(df=sql_task(limit=limit))


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(my_wf(limit=3))
