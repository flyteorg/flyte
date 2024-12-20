import os
import pathlib
import sqlite3
import typing

import pandas as pd
import pytest
from flytekitplugins.great_expectations import BatchRequestConfig, GreatExpectationsFlyteConfig, GreatExpectationsType
from great_expectations.exceptions import InvalidBatchRequestError, ValidationError

from flytekit import task, workflow
from flytekit.types.file import CSVFile
from flytekit.types.schema import FlyteSchema

this_dir = pathlib.Path(__file__).resolve().parent
os.chdir(this_dir)


def test_ge_type():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )

    s = GreatExpectationsType[str, ge_config]
    assert s.config()[1] == ge_config
    assert s.config()[1].datasource_name == "data"


def test_ge_schema_task():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )

    @task
    def get_length(dataframe: pd.DataFrame) -> int:
        return dataframe.shape[0]

    @task
    def my_task(csv_file: GreatExpectationsType[str, ge_config]) -> pd.DataFrame:
        df = pd.read_csv(os.path.join("data", csv_file))
        df.drop(5, axis=0, inplace=True)
        return df

    @workflow
    def valid_wf(dataset: str = "yellow_tripdata_sample_2019-01.csv") -> int:
        dataframe = my_task(csv_file=dataset)
        return get_length(dataframe=dataframe)

    @workflow
    def invalid_wf(dataset: str = "yellow_tripdata_sample_2019-02.csv") -> int:
        dataframe = my_task(csv_file=dataset)
        return get_length(dataframe=dataframe)

    valid_result = valid_wf()
    assert valid_result == 9999

    with pytest.raises(ValidationError):
        invalid_wf()


def test_ge_schema_multiple_args():
    ge_config_one = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )
    ge_config_two = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test1.demo",
        data_connector_name="data_example_data_connector",
    )

    @task
    def get_file_name(
        dataset_one: GreatExpectationsType[str, ge_config_one], dataset_two: GreatExpectationsType[str, ge_config_two]
    ) -> typing.Tuple[int, int]:
        df_one = pd.read_csv(os.path.join("data", dataset_one))
        df_two = pd.read_csv(os.path.join("data", dataset_two))
        return len(df_one), len(df_two)

    @workflow
    def wf(
        dataset_one: str = "yellow_tripdata_sample_2019-01.csv", dataset_two: str = "yellow_tripdata_sample_2019-02.csv"
    ) -> typing.Tuple[int, int]:
        return get_file_name(dataset_one=dataset_one, dataset_two=dataset_two)

    assert wf() == (10000, 10000)


def test_ge_schema_batchrequest_pandas_config():
    @task
    def my_task(
        directory: GreatExpectationsType[
            str,
            GreatExpectationsFlyteConfig(
                datasource_name="data",
                expectation_suite_name="test.demo",
                data_connector_name="my_data_connector",
                batch_request_config=BatchRequestConfig(
                    data_connector_query={
                        "batch_filter_parameters": {
                            "year": "2019",
                            "month": "01",  # noqa: F722
                        },
                        "limit": 10,
                    },
                ),
            ),
        ],
    ) -> str:
        return directory

    @workflow
    def my_wf():
        my_task(directory="my_assets")

    my_wf()


def test_invalid_ge_schema_batchrequest_pandas_config():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="my_data_connector",
        batch_request_config=BatchRequestConfig(
            data_connector_query={
                "batch_filter_parameters": {
                    "year": "2020",
                },
            }
        ),
    )

    @task
    def my_task(directory: GreatExpectationsType[str, ge_config]) -> str:
        return directory

    @workflow
    def my_wf():
        my_task(directory="my_assets")

    # Capture IndexError
    with pytest.raises(InvalidBatchRequestError):
        my_wf()


def test_ge_schema_runtimebatchrequest_sqlite_config():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="sqlite_data",
        expectation_suite_name="sqlite.movies",
        data_connector_name="sqlite_data_connector",
        data_asset_name="sqlite_data",
        batch_request_config=BatchRequestConfig(
            batch_identifiers={
                "pipeline_stage": "validation",
            },
        ),
    )

    @task
    def my_task(sqlite_db: GreatExpectationsType[str, ge_config]) -> int:
        # read sqlite query results into a pandas DataFrame
        con = sqlite3.connect(os.path.join("data/movies.sqlite"))
        df = pd.read_sql_query("SELECT * FROM movies", con)
        con.close()

        # verify that result of SQL query is stored in the dataframe
        return len(df)

    @workflow
    def my_wf() -> int:
        return my_task(sqlite_db="SELECT * FROM movies")

    result = my_wf()
    assert result == 2736


def test_ge_runtimebatchrequest_pandas_config():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="my_pandas_datasource",
        expectation_suite_name="test.demo",
        data_connector_name="my_runtime_data_connector",
        data_asset_name="pandas_data",
        batch_request_config=BatchRequestConfig(
            batch_identifiers={
                "pipeline_stage": "validation",
            },
        ),
    )

    @task
    def my_task(pandas_df: GreatExpectationsType[FlyteSchema, ge_config]) -> int:
        return len(pandas_df.open().all())

    @workflow
    def runtime_pandas_wf(df: pd.DataFrame):
        my_task(pandas_df=df)

    runtime_pandas_wf(df=pd.read_csv("data/yellow_tripdata_sample_2019-01.csv"))


def test_ge_schema_checkpoint_params():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
        checkpoint_params={
            "site_names": ["local_site"],
        },
    )

    @task
    def my_task(dataset: GreatExpectationsType[str, ge_config]) -> None:
        assert type(dataset) == str

    @workflow
    def my_wf() -> None:
        my_task(dataset="yellow_tripdata_sample_2019-01.csv")

    my_wf()


def test_ge_schema_flyteschema():
    @task
    def my_task(
        dataframe: GreatExpectationsType[
            FlyteSchema,
            GreatExpectationsFlyteConfig(
                datasource_name="data",
                expectation_suite_name="test.demo",
                data_connector_name="data_flytetype_data_connector",
                batch_request_config=BatchRequestConfig(data_connector_query={"limit": 10}),
                local_file_path="/tmp/test3.parquet",  # noqa: F722
            ),
        ],
    ) -> int:
        return dataframe.open().all().shape[0]

    @workflow
    def valid_wf(dataframe: FlyteSchema) -> int:
        return my_task(dataframe=dataframe)

    df = pd.read_csv("data/yellow_tripdata_sample_2019-01.csv")
    result = valid_wf(dataframe=df)
    assert result == 10000


def test_ge_schema_flyteschema_literal():
    @task
    def my_task(
        dataframe: GreatExpectationsType[
            FlyteSchema,
            GreatExpectationsFlyteConfig(
                datasource_name="data",
                expectation_suite_name="test.demo",
                data_connector_name="data_flytetype_data_connector",
                batch_request_config=BatchRequestConfig(data_connector_query={"limit": 10}),
                local_file_path="/tmp/test3.parquet",  # noqa: F722
            ),
        ],
    ) -> int:
        return dataframe.open().all().shape[0]

    @workflow
    def valid_wf() -> int:
        df = pd.read_csv("data/yellow_tripdata_sample_2019-01.csv")
        return my_task(dataframe=df)

    result = valid_wf()
    assert result == 10000


def test_ge_schema_remote_flytefile():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @task
    def my_task(dataset: GreatExpectationsType[CSVFile, ge_config]) -> int:
        return len(pd.read_csv(dataset))

    @workflow
    def my_wf(dataset: CSVFile) -> int:
        return my_task(dataset=dataset)

    result = my_wf(
        dataset="https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv"
    )
    assert result == 10000


def test_ge_schema_remote_flytefile_literal():
    ge_config = GreatExpectationsFlyteConfig(
        datasource_name="data",
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @task
    def my_task(dataset: GreatExpectationsType[CSVFile, ge_config]) -> int:
        return len(pd.read_csv(dataset))

    @workflow
    def my_wf() -> int:
        return my_task(
            dataset="https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv"
        )

    result = my_wf()
    assert result == 10000
