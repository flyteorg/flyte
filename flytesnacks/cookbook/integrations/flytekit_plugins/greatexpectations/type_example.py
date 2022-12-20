"""
Type Example
------------

``GreatExpectationsType`` when accompanied with data can be used for data validation.
It essentially is the type we attach to the data we want to validate.
In this example, we'll implement a simple task, followed by Great Expectations data validation on ``FlyteFile``, ``FlyteSchema``, and finally,
the :py:class:`RuntimeBatchRequest <great_expectations.core.batch.RuntimeBatchRequest>`. The following video is a demo of the Great Expectations type example.

.. youtube:: pDFTt_g76Wc

"""

# %%
# First, let's import the required libraries.
import os

import pandas as pd
from flytekit import Resources, task, workflow
from flytekit.types.file import CSVFile
from flytekit.types.schema import FlyteSchema
from flytekitplugins.great_expectations import BatchRequestConfig, GreatExpectationsFlyteConfig, GreatExpectationsType

# %%
# .. note::
#   ``BatchRequestConfig`` is useful in giving additional batch request parameters to construct both Great Expectations'
#   ``RuntimeBatchRequest`` and ``BatchRequest``.
#   Moreover, there's ``GreatExpectationsFlyteConfig`` that encapsulates the essential initialization parameters of the plugin.

# %%
# Next, we define variables that we use throughout the code.
DATASET_LOCAL = "yellow_tripdata_sample_2019-01.csv"
DATASET_REMOTE = "https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv"
CONTEXT_ROOT_DIR = "greatexpectations/great_expectations"

# %%
# Simple Type
# ===========
#
# We define a ``GreatExpectationsType`` that checks if the requested ``batch_filter_parameters`` can be used to fetch files from a directory.
# The directory that's being used is defined in ``my_assets``. You can find ``my_assets`` in the Great Expectations config file.
#
# The parameters within the ``data_connector_query`` convey that we're fetching all those files that have "2019" and "01" in the file names.


@task(limits=Resources(mem="500Mi"))
def simple_task(
    directory: GreatExpectationsType[
        str,
        GreatExpectationsFlyteConfig(
            datasource_name="data",  # noqa: F821
            expectation_suite_name="test.demo",  # noqa: F821
            data_connector_name="my_data_connector",  # noqa: F821
            batch_request_config=BatchRequestConfig(
                data_connector_query={
                    "batch_filter_parameters": {  # noqa: F821
                        "year": "2019",  # noqa: F821
                        "month": "01",  # noqa: F821, F722
                    },
                    "limit": 10,  # noqa: F821
                },
            ),
            context_root_dir=CONTEXT_ROOT_DIR,
        ),
    ]
) -> str:
    return f"Validation works for {directory}!"


# %%
# Finally, we define a workflow to call our task.
@workflow
def simple_wf(directory: str = "my_assets") -> str:
    return simple_task(directory=directory)


# %%
# FlyteFile
# =========
#
# First, we define ``GreatExpectationsFlyteConfig`` to initialize all our parameters. Here, we're validating a ``FlyteFile``.
great_expectations_config = GreatExpectationsFlyteConfig(
    datasource_name="data",
    expectation_suite_name="test.demo",
    data_connector_name="data_flytetype_data_connector",
    local_file_path="/tmp",
    context_root_dir=CONTEXT_ROOT_DIR,
)

# %%
# Next, we map ``dataset`` parameter to ``GreatExpectationsType``.
# Under the hood, ``GreatExpectationsType`` validates data in accordance with the ``GreatExpectationsFlyteConfig`` defined previously.
# This ``GreatExpectationsFlyteConfig`` is being fetched under the name ``great_expectations_config``.
#
# The first value that's being sent within ``GreatExpectationsType`` is ``CSVFile`` (this is a pre-formatted FlyteFile type).
# This means that we want to validate the ``FlyteFile`` data.


@task(limits=Resources(mem="500Mi"))
def file_task(
    dataset: GreatExpectationsType[CSVFile, great_expectations_config]
) -> pd.DataFrame:
    return pd.read_csv(dataset)


# %%
# Next, we define a workflow to call our task.
@workflow
def file_wf() -> pd.DataFrame:
    return file_task(dataset=DATASET_REMOTE)


# %%
# FlyteSchema
# ===========
#
# We define a ``GreatExpectationsType`` to validate ``FlyteSchema``. The ``local_file_path`` is where we would have our parquet file.
#
# .. note::
#   ``local_file_path``'s directory and ``base_directory`` in Great Expectations config ought to be the same.
@task(limits=Resources(mem="500Mi"))
def schema_task(
    dataframe: GreatExpectationsType[
        FlyteSchema,  # noqa: F821
        GreatExpectationsFlyteConfig(  # noqa: F821
            datasource_name="data",  # noqa: F821
            expectation_suite_name="test.demo",  # noqa: F821
            data_connector_name="data_flytetype_data_connector",  # noqa: F821
            batch_request_config=BatchRequestConfig(
                data_connector_query={"limit": 10}  # noqa : F841
            ),  # noqa: F821
            local_file_path="/tmp/test.parquet",  # noqa: F722
            context_root_dir=CONTEXT_ROOT_DIR,
        ),
    ]
) -> int:
    return dataframe.shape[0]


# %%
# Finally, we define a workflow to call our task.
# We're using DataFrame returned by the ``file_task`` that we defined in the ``FlyteFile`` section.
@workflow
def schema_wf() -> int:
    return schema_task(dataframe=file_wf())


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
# We instantiate ``data_asset_name`` to associate it with the ``RuntimeBatchRequest``.
# The typical Great Expectations' batch_data (or) query is automatically populated with the dataset.
#
# .. note::
#   If you want to load a database table as a batch, your dataset has to be a SQL query.
runtime_ge_config = GreatExpectationsFlyteConfig(
    datasource_name="my_pandas_datasource",
    expectation_suite_name="test.demo",
    data_connector_name="my_runtime_data_connector",
    data_asset_name="validate_pandas_data",
    batch_request_config=BatchRequestConfig(
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
# We define a task to validate the DataFrame.
@task
def runtime_validation(
    dataframe: GreatExpectationsType[FlyteSchema, runtime_ge_config]
) -> int:
    return len(dataframe)


# %%
# Finally, we define a workflow to run our tasks.
@workflow
def runtime_wf(dataset: str = DATASET_LOCAL) -> int:
    dataframe = runtime_to_df_task(csv_file=dataset)
    return runtime_validation(dataframe=dataframe)


# %%
# Lastly, this particular block of code helps us in running the code locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print("Simple Great Expectations Type...")
    print(simple_wf())
    print("Great Expectations Type with FlyteFile...")
    print(file_wf())
    print("Great Expectations Type with FlyteSchema...")
    print(schema_wf())
    print("Great Expectations Type with RuntimeBatchRequest...")
    print(runtime_wf())
