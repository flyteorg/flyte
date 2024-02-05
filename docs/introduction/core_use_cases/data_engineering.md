---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_data_engineering)=

# Data Engineering

Flyte is well-suited for data engineering use cases, where you can interleave
SQL queries with data processing logic implemented in Python with whichever
data processing tools you prefer.

In this example, we create an ETL workflow that extracts data from a public
[RNA database](https://rnacentral.org/help/public-database), performs some simple
transforms on the data, and loads it into a CSV file.

## Extract

First, we define an `extract_task` task using the
{ref}`flytekitplugins-sqlalchemy <sql_alchemy>` plugin, which provides an
interface to perform SQL queries via the
{py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyTask`
and {py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyConfig` classes.

```{code-cell} ipython3
import os

import flytekit
import pandas as pd
from flytekit import Resources, kwtypes, task, workflow
from flytekit.types.file import CSVFile
from flytekitplugins.sqlalchemy import SQLAlchemyConfig, SQLAlchemyTask

DATABASE_URI = (
    "postgresql://reader:NWDMCE5xdipIjRrp@hh-pgsql-public.ebi.ac.uk:5432/pfmegrnargs"
)

extract_task = SQLAlchemyTask(
    "extract_rna",
    query_template=(
        "select len as sequence_length, timestamp from rna "
        "where len >= {{ .inputs.min_length }} and len <= {{ .inputs.max_length }} "
        "limit {{ .inputs.limit }}"
    ),
    inputs=kwtypes(min_length=int, max_length=int, limit=int),
    output_schema_type=pd.DataFrame,
    task_config=SQLAlchemyConfig(uri=DATABASE_URI),
)
```

You can format the `query_template` with `{{ .inputs.<input_name> }}` to
parameterize your query with the `input` keyword type specification, which
maps task argument names to their expected types.

```{important}
You can request for access to secrets via the `secret_requests` of the
{py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyTask` constructor, then
pass in a `secret_connect_args` argument to the
{py:class}`~flytekitplugins.sqlalchemy.SQLAlchemyConfig` constructor, assuming
that you have connection credentials available in the configured
{ref}`Secrets Management System <secrets>`, which is K8s by default.
```

## Transform

Next, we parse the raw `timestamp`s and represent the time as separate `date`
and `time` columns. Notice that we can encode the assumptions we have about this
task's resource requirements with the {py:class}`~flytekit.Resources` object.
If those assumptions ever change, we can update the resource request here, or
override it at the workflow-level with the {ref}`with_overrides <resource_with_overrides>` method.

```{code-cell} ipython3
@task(requests=Resources(mem="700Mi"))
def transform(df: pd.DataFrame) -> pd.DataFrame:
    """Add date and time columns; drop timestamp column."""
    timestamp = pd.to_datetime(df["timestamp"])
    df["date"] = timestamp.dt.date
    df["time"] = timestamp.dt.time
    df.drop("timestamp", axis=1, inplace=True)
    return df
```

## Load

Finally, we load the transformed data into its final destination: a CSV file in
blob storage. Flyte has a built-in `CSVFile` type that automatically handles
serializing/deserializing and uploading/downloading the file as it's passed from
one task to the next. All you need to do is write the file to some local location
and pass that location to the `path` argument of `CSVFile`.

```{code-cell} ipython3
@task(requests=Resources(mem="700Mi"))
def load(df: pd.DataFrame) -> CSVFile:
    """Load the dataframe to a csv file."""
    csv_file = os.path.join(flytekit.current_context().working_directory, "rna_df.csv")
    df.to_csv(csv_file, index=False)
    return CSVFile(path=csv_file)
```

## ETL Workflow

Putting all the pieces together, we create an `etl_workflow` that produces a
dataset based on the parameters you give it.

```{code-cell} ipython3
@workflow
def etl_workflow(
    min_length: int = 50, max_length: int = 200, limit: int = 10
) -> CSVFile:
    """Build an extract, transform and load pipeline."""
    return load(
        df=transform(
            df=extract_task(min_length=min_length, max_length=max_length, limit=limit)
        )
    )
```

During local execution, this CSV file lives in a random local
directory, but when the workflow is run on a Flyte cluster, this file lives in
the configured blob store, like S3 or GCS.

Running this workflow locally, we can access the CSV file and read it into
a `pandas.DataFrame`.

```{code-cell} ipython3
csv_file = etl_workflow(limit=5)
pd.read_csv(csv_file)
```

## Workflows as Reusable Components

Because Flyte tasks and workflows are simply functions, we can embed
`etl_workflow` as part of a larger workflow, where it's used to create a
CSV file that's then consumed by downstream tasks or subworkflows:

```{code-cell} ipython3
@task
def aggregate(file: CSVFile) -> pd.DataFrame:
    data = pd.read_csv(file)
    ... # process the data further


@task
def plot(data: pd.DataFrame):
    ...  # create a plot


@workflow
def downstream_workflow(
    min_length: int = 50, max_length: int = 200, limit: int = 10
):
    """A downstream workflow that visualizes an aggregation of the data."""
    csv_file = etl_workflow(
        min_length=min_length,
        max_length=max_length,
        limit=limit,
    )
    return plot(data=aggregate(file=csv_file))
```

```{important}
Prefer other data processing frameworks? Flyte ships with
[Polars](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-polars),
{ref}`Dask <plugins-dask-k8s>`, {ref}`Modin <modin-integration>`, {ref}`Spark <plugins-spark-k8s>`,
[Vaex](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-vaex),
and [DBT](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-dbt)
integrations.

For database connectors, Flyte provides first-party support for {ref}`AWS Athena <aws-athena>`,
{ref}`Google Bigquery <big-query>`, {ref}`Snowflake <plugins-snowflake>`,
{ref}`SQLAlchemy <sql_alchemy>`, and {ref}`SQLite3 <integrations_sql_sqlite3>`.
```
