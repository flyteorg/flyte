---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(integrations_sql_sqlite3)=

# Sqlite3

The following example shows how you can write SQLite3 queries using the SQLite3Task, which is bundled as part of the
core flytekit. Since SQL Queries are portable across workflows and Flyte installations (as long as the data exists),
this task will always run with a pre-built container, specifically the [flytekit container](https://github.com/flyteorg/flytekit/blob/v0.19.0/Dockerfile.py38)
itself. Therefore, users are not required to build a container for SQLite3. You can simply implement the task - register and
execute it immediately.

In some cases local execution is not possible - e.g. Snowflake. But for SQLlite3 local execution is also supported.

```{code-cell}
import pandas
from flytekit import kwtypes, task, workflow
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
```

+++ {"lines_to_next_cell": 0}

SQLite3 queries in flyte produce a Schema output. The data in this example is
taken from [here](https://www.sqlitetutorial.net/sqlite-sample-database/).

```{code-cell}
from flytekit.types.schema import FlyteSchema

EXAMPLE_DB = "https://www.sqlitetutorial.net/wp-content/uploads/2018/03/chinook.zip"
```

+++ {"lines_to_next_cell": 0}

the task is declared as a regular task. Alternatively it can be declared within a workflow just at the point of using
it (example later)

```{code-cell}
sql_task = SQLite3Task(
    name="sqlite3.sample",
    query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
    inputs=kwtypes(limit=int),
    output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
    task_config=SQLite3Config(uri=EXAMPLE_DB, compressed=True),
)
```

+++ {"lines_to_next_cell": 0}

As described elsewhere FlyteSchemas can be easily be received as pandas Dataframe and Flyte will autoconvert them

```{code-cell}
@task
def print_and_count_columns(df: pandas.DataFrame) -> int:
    return len(df[df.columns[0]])
```

+++ {"lines_to_next_cell": 0}

The task can be used normally in the workflow, passing the declared inputs

```{code-cell}
@workflow
def wf() -> int:
    return print_and_count_columns(df=sql_task(limit=100))
```

+++ {"lines_to_next_cell": 0}

It can also be executed locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running main {wf()}")
```

+++ {"lines_to_next_cell": 0}

As mentioned earlier it is possible to also write the SQL Task inline as follows

```{code-cell}
@workflow
def query_wf() -> int:
    df = SQLite3Task(
        name="sqlite3.sample_inline",
        query_template="select TrackId, Name from tracks limit {{.inputs.limit}}",
        inputs=kwtypes(limit=int),
        output_schema_type=FlyteSchema[kwtypes(TrackId=int, Name=str)],
        task_config=SQLite3Config(uri=EXAMPLE_DB, compressed=True),
    )(limit=100)
    return print_and_count_columns(df=df)
```

+++ {"lines_to_next_cell": 0}

It can also be executed locally.

```{code-cell}
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running main {query_wf()}")
```
