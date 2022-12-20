"""
Task Example
------------

``GreatExpectationsTask`` can be used to define data validation within the task.
In this example, we'll implement a simple task, followed by Great Expectations data validation on ``FlyteFile``, ``FlyteSchema``, and finally,
the :py:class:`RuntimeBatchRequest <great_expectations.core.batch.RuntimeBatchRequest>`.

The following video shows the inner workings of the Great Expectations plugin, plus a demo of the task example.

.. youtube:: wjiO7jassrw

"""

# %%
# First, let's import the required libraries.
import os
import typing

import pandas as pd
from flytekit import Resources, kwtypes, task, workflow
from flytekit.extras.sqlite3.task import SQLite3Config, SQLite3Task
from flytekit.types.file import CSVFile
from flytekit.types.schema import FlyteSchema
from flytekitplugins.great_expectations import BatchRequestConfig, GreatExpectationsTask

# %%
# .. note::
#   ``BatchRequestConfig`` is useful in giving additional batch request parameters to construct
#   both Great Expectations' ``RuntimeBatchRequest`` and ``BatchRequest``.

# %%
# Next, we define variables that we use throughout the code.
CONTEXT_ROOT_DIR = "greatexpectations/great_expectations"
DATASET_LOCAL = "yellow_tripdata_sample_2019-01.csv"
DATASET_REMOTE = f"https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/{DATASET_LOCAL}"
SQLITE_DATASET = "https://cdn.discordapp.com/attachments/545481172399030272/867254085426085909/movies.sqlite"


# %%
# Simple Task
# ===========
#
# We define a ``GreatExpectationsTask`` that validates a CSV file. This does pandas data validation.
simple_task_object = GreatExpectationsTask(
    name="great_expectations_task_simple",
    datasource_name="data",
    inputs=kwtypes(dataset=str),
    expectation_suite_name="test.demo",
    data_connector_name="data_example_data_connector",
    context_root_dir=CONTEXT_ROOT_DIR,
)

# %%
# Next, we define a task that validates the data before returning the shape of the DataFrame.
@task(limits=Resources(mem="500Mi"))
def simple_task(csv_file: str) -> int:
    # GreatExpectationsTask returns Great Expectations' checkpoint result.
    # You can print the result to know more about the data within it.
    # If the data validation fails, this will return a ValidationError.
    result = simple_task_object(dataset=csv_file)
    df = pd.read_csv(os.path.join("greatexpectations", "data", csv_file))
    return df.shape[0]


# %%
# Finally, we define a workflow.
@workflow
def simple_wf(dataset: str = DATASET_LOCAL) -> int:
    return simple_task(csv_file=dataset)


# %%
# FlyteFile
# =========
#
# We define a ``GreatExpectationsTask`` that validates a ``FlyteFile``.
# Here, we're using a different data connector owing to the different ``base_directory`` we're using within the Great Expectations config file.
# The ``local_file_path`` argument helps in copying the remote file to the user-given path.
#
# .. note::
#   ``local_file_path``'s directory and ``base_directory`` in Great Expectations config ought to be the same.
file_task_object = GreatExpectationsTask(
    name="great_expectations_task_flytefile",
    datasource_name="data",
    inputs=kwtypes(dataset=CSVFile),
    expectation_suite_name="test.demo",
    data_connector_name="data_flytetype_data_connector",
    local_file_path="/tmp",
    context_root_dir=CONTEXT_ROOT_DIR,
)

# %%
# Next, we define a task that calls the validation logic.
@task(limits=Resources(mem="500Mi"))
def file_task(
    dataset: CSVFile,
) -> int:
    file_task_object(dataset=dataset)
    return len(pd.read_csv(dataset))


# %%
# Finally, we define a workflow to run our task.
@workflow
def file_wf(
    dataset: CSVFile = DATASET_REMOTE,
) -> int:
    return file_task(dataset=dataset)


# %%
# FlyteSchema
# ===========
#
# We define a ``GreatExpectationsTask`` that validates FlyteSchema.
# The ``local_file_path`` here refers to the parquet file in which we want to store our DataFrame.
schema_task_object = GreatExpectationsTask(
    name="great_expectations_task_schema",
    datasource_name="data",
    inputs=kwtypes(dataset=FlyteSchema),
    expectation_suite_name="sqlite.movies",
    data_connector_name="data_flytetype_data_connector",
    local_file_path="/tmp/test.parquet",
    context_root_dir=CONTEXT_ROOT_DIR,
)

# %%
# Let's fetch the DataFrame from the SQL Database we've with us. To do so, we use the ``SQLite3Task`` available within Flyte.
sql_to_df = SQLite3Task(
    name="greatexpectations.task.sqlite3",
    query_template="select * from movies",
    output_schema_type=FlyteSchema,
    task_config=SQLite3Config(uri=SQLITE_DATASET),
)

# %%
# Next, we define a task that validates the data and returns the columns in it.
@task(limits=Resources(mem="500Mi"))
def schema_task(dataset: pd.DataFrame) -> typing.List[str]:
    schema_task_object(dataset=dataset)
    return list(dataset.columns)


# %%
# Finally, we define a workflow to fetch the DataFrame and validate it.
@workflow
def schema_wf() -> typing.List[str]:
    df = sql_to_df()
    return schema_task(dataset=df)


# %%
# RuntimeBatchRequest
# ===================
#
# The :py:class:`RuntimeBatchRequest <great_expectations.core.batch.RuntimeBatchRequest>` can wrap either an in-memory DataFrame,
# filepath, or SQL query, and must include batch identifiers that uniquely identify the data.
#
# Let's instantiate a ``RuntimeBatchRequest`` that accepts a DataFrame and thereby validates it.
#
# .. note::
#   The plugin determines the type of request as ``RuntimeBatchRequest`` by analyzing the user-given data connector.
#
# We give ``data_asset_name`` to associate it with the ``RuntimeBatchRequest``.
# The typical Great Expectations' ``batch_data`` (or) ``query`` is automatically populated with the dataset.
#
# .. note::
#   If you want to load a database table as a batch, your dataset has to be a SQL query.
runtime_task_obj = GreatExpectationsTask(
    name="greatexpectations.task.runtime",
    datasource_name="my_pandas_datasource",
    inputs=kwtypes(dataframe=FlyteSchema),
    expectation_suite_name="test.demo",
    data_connector_name="my_runtime_data_connector",
    data_asset_name="validate_pandas_data",
    task_config=BatchRequestConfig(
        batch_identifiers={
            "pipeline_stage": "validation",
        },
    ),
    context_root_dir=CONTEXT_ROOT_DIR,
)


# %%
# We define a task to generate DataFrame from the CSV file.
@task
def runtime_to_df_task(csv_file: str) -> pd.DataFrame:
    df = pd.read_csv(os.path.join("greatexpectations", "data", csv_file))
    return df


# %%
# Finally, we define a workflow to run our task.
@workflow
def runtime_wf(dataset: str = DATASET_LOCAL) -> None:
    dataframe = runtime_to_df_task(csv_file=dataset)
    runtime_task_obj(dataframe=dataframe)


# %%
# Lastly, this particular block of code helps us in running the code locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print("Simple Great Expectations Task...")
    print(simple_wf())
    print("Great Expectations Task with FlyteFile...")
    print(file_wf())
    print("Great Expectations Task with FlyteSchema...")
    print(schema_wf())
    print("Great Expectations Task with RuntimeBatchRequest...")
    runtime_wf()
