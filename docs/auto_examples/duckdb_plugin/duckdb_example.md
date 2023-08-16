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

# DuckDB Example

In this example, we will see how to run SQL queries using DuckDB.

First, import the necessary libraries.

```{code-cell}
import json
from typing import List

import pandas as pd
import pyarrow as pa
from flytekit import kwtypes, task, workflow
from flytekit.types.structured.structured_dataset import StructuredDataset
from flytekitplugins.duckdb import DuckDBQuery
from typing_extensions import Annotated
```

+++ {"lines_to_next_cell": 0}

## A simple `SELECT` query

To run a simple `SELECT` query, initialize a `DuckDBQuery` task.

```{code-cell}
simple_duckdb_query = DuckDBQuery(
    name="simple_task",
    query="SELECT SUM(a) FROM mydf",
    inputs=kwtypes(mydf=pd.DataFrame),
)
```

+++ {"lines_to_next_cell": 0}

:::{note}
The default output type for the `DuckDBQuery` task is `StructuredDataset`.
Hence, it is possible to retrieve any compatible type of Structured Dataset such as a Pandas dataframe, Vaex dataframe, and others.
:::

You can invoke the task from within a {py:func}`~flytekit:flytekit.workflow` and return both a Pandas dataframe and a PyArrow table.
The query will be executed on a Pandas dataframe, and the resulting output can belong to any StructuredDataset-compatible type.

```{code-cell}
@task
def get_pandas_df() -> pd.DataFrame:
    return pd.DataFrame({"a": [1, 2, 3]})


@workflow
def pandas_wf() -> pd.DataFrame:
    return simple_duckdb_query(mydf=get_pandas_df())


@workflow
def arrow_wf() -> pa.Table:
    return simple_duckdb_query(mydf=get_pandas_df())


if __name__ == "__main__":
    print(f"Running pandas_wf()... {pandas_wf()}")
    print(f"Running arrow_wf()... {arrow_wf()}")
```

+++ {"lines_to_next_cell": 0}

## SQL query on Parquet file

DuckDB enables direct querying of a parquet file without the need for intermediate conversions to a database.
If you wish to execute a SQL query on a parquet file stored in a public S3 bucket, you can use the `httpfs` library by installing and loading it.
Simply send the parquet file as a parameter to the `SELECT` query.

It is important to note that multiple commands can be executed within a single `DuckDBQuery`.
However, it is essential to ensure that the last command within the query is a "SELECT" query to retrieve data in the end.

```{code-cell}
parquet_duckdb_query = DuckDBQuery(
    name="parquet_query",
    query=[
        "INSTALL httpfs",
        "LOAD httpfs",
        """SELECT hour(lpep_pickup_datetime) AS hour, count(*) AS count FROM READ_PARQUET(?) GROUP BY hour""",
    ],
    inputs=kwtypes(params=List[str]),
)


@workflow
def parquet_wf(parquet_file: str) -> pd.DataFrame:
    return parquet_duckdb_query(params=[parquet_file])


if __name__ == "__main__":
    parquet_file = "https://d37ci6vzurychx.cloudfront.net/trip-data/green_tripdata_2022-02.parquet"
    print(f"Running parquet_wf()... {parquet_wf(parquet_file=parquet_file)}")
```

+++ {"lines_to_next_cell": 0}

## SQL query on StructuredDataset

To execute a SQL query on a structured dataset, you can simply run a query just like any other query on a Pandas dataframe or PyArrow table.

```{code-cell}
sd_duckdb_query = DuckDBQuery(
    name="sd_query",
    query="SELECT * FROM sd_df WHERE i = 2",
    inputs=kwtypes(sd_df=StructuredDataset),
)


@task
def get_sd() -> StructuredDataset:
    return StructuredDataset(
        dataframe=pd.DataFrame.from_dict({"i": [1, 2, 3, 4], "j": ["one", "two", "three", "four"]})
    )


@workflow
def sd_wf() -> pd.DataFrame:
    sd_df = get_sd()
    return sd_duckdb_query(sd_df=sd_df)


if __name__ == "__main__":
    print(f"Running sd_wf()... {sd_wf()}")
```

+++ {"lines_to_next_cell": 0}

## Send parameters to multiple queries

To send parameters to multiple queries, use list of lists.

:::{note}
Sometimes, the annotation of parameter types can be somewhat complicated.
In such situations, you can convert the list to a string using `json.dumps`.
The string will be automatically loaded into a list under the hood.
If the length of the query list is 3 and the length of the parameter list is 2,
the plugin will search for parameter acceptance symbols ("?" or "\$") in each query
to determine whether to include or exclude the parameters before executing the query.
Therefore, it is necessary to provide the query parameters in the same order as the queries listed.
:::

```{code-cell}
duckdb_params_query = DuckDBQuery(
    name="params_query",
    query=[
        "CREATE TABLE items(item VARCHAR, value DECIMAL(10,2), count INTEGER)",
        "INSERT INTO items VALUES (?, ?, ?)",
        "SELECT $1 AS one, $1 AS two, $2 AS three",
    ],
    inputs=kwtypes(params=str),
)


@task
def read_df(df: Annotated[StructuredDataset, kwtypes(one=str)]) -> pd.DataFrame:
    return df.open(pd.DataFrame).all()


@workflow
def params_wf(
    params: str = json.dumps(
        [
            [["chainsaw", 500, 10], ["iphone", 300, 2]],
            ["duck", "goose"],
        ]
    )
) -> pd.DataFrame:
    return read_df(df=duckdb_params_query(params=params))


if __name__ == "__main__":
    print(f"Running params_wf()... {params_wf()}")
```
