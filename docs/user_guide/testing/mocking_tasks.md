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
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Mocking tasks

A lot of the tasks that you write you can run locally, but some of them you will not be able to, usually because they
are tasks that depend on a third-party only available on the backend. Hive tasks are a common example, as most users
will not have access to the service that executes Hive queries from their development environment. However, it's still
useful to be able to locally run a workflow that calls such a task. In these instances, flytekit provides a couple
of utilities to help navigate this.

```{code-cell}
import datetime

import pandas
from flytekit import SQLTask, TaskMetadata, kwtypes, task, workflow
from flytekit.testing import patch, task_mock
from flytekit.types.schema import FlyteSchema
```

+++ {"lines_to_next_cell": 0}

This is a generic SQL task (and is by default not hooked up to any datastore nor handled by any plugin), and must
be mocked.

```{code-cell}
sql = SQLTask(
    "my-query",
    query_template="SELECT * FROM hive.city.fact_airport_sessions WHERE ds = '{{ .Inputs.ds }}' LIMIT 10",
    inputs=kwtypes(ds=datetime.datetime),
    outputs=kwtypes(results=FlyteSchema),
    metadata=TaskMetadata(retries=2),
)
```

+++ {"lines_to_next_cell": 0}

This is a task that can run locally

```{code-cell}
@task
def t1() -> datetime.datetime:
    return datetime.datetime.now()
```

+++ {"lines_to_next_cell": 0}

Declare a workflow that chains these two tasks together.

```{code-cell}
@workflow
def my_wf() -> FlyteSchema:
    dt = t1()
    return sql(ds=dt)
```

+++ {"lines_to_next_cell": 0}

Without a mock, calling the workflow would typically raise an exception, but with the `task_mock` construct, which
returns a `MagicMock` object, we can override the return value.

```{code-cell}
def main_1():
    with task_mock(sql) as mock:
        mock.return_value = pandas.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})
        assert (my_wf().open().all() == pandas.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()
```

+++ {"lines_to_next_cell": 0}

There is another utility as well called `patch` which offers the same functionality, but in the traditional Python
patching style, where the first argument is the `MagicMock` object.

```{code-cell}
def main_2():
    @patch(sql)
    def test_user_demo_test(mock_sql):
        mock_sql.return_value = pandas.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})
        assert (my_wf().open().all() == pandas.DataFrame(data={"x": [1, 2], "y": ["3", "4"]})).all().all()

    test_user_demo_test()


if __name__ == "__main__":
    main_1()
    main_2()
```
