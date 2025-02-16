import os
import pathlib
import sqlite3
import typing

import pandas as pd
import pytest
from flytekitplugins.great_expectations import BatchRequestConfig, GreatExpectationsTask
from flytekitplugins.spark import Spark
from great_expectations.exceptions import InvalidBatchRequestError, ValidationError

import flytekit
from flytekit import kwtypes, task, workflow
from flytekit.types.file import CSVFile, FlyteFile
from flytekit.types.schema import FlyteSchema

this_dir = pathlib.Path(__file__).resolve().parent
os.chdir(this_dir)


def test_ge_simple_task():
    task_object = GreatExpectationsTask(
        name="test1",
        datasource_name="data",
        inputs=kwtypes(dataset=str),
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )

    # valid data
    result = task_object(dataset="yellow_tripdata_sample_2019-01.csv")

    assert result["success"] is True
    assert result["statistics"]["evaluated_expectations"] == result["statistics"]["successful_expectations"]

    # invalid data
    with pytest.raises(ValidationError):
        invalid_result = task_object(dataset="yellow_tripdata_sample_2019-02.csv")
        assert invalid_result["success"] is False
        assert (
            invalid_result["statistics"]["evaluated_expectations"]
            != invalid_result["statistics"]["successful_expectations"]
        )

    assert task_object.python_interface.inputs == {"dataset": str}


def test_ge_batchrequest_pandas_config():
    task_object = GreatExpectationsTask(
        name="test2",
        datasource_name="data",
        inputs=kwtypes(data=str),
        expectation_suite_name="test.demo",
        data_connector_name="my_data_connector",
        task_config=BatchRequestConfig(
            data_connector_query={
                "batch_filter_parameters": {
                    "year": "2019",
                    "month": "01",
                },
                "limit": 10,
            },
        ),
    )

    # name of the asset -- can be found in great_expectations.yml file
    task_object(data="my_assets")


def test_invalid_ge_batchrequest_pandas_config():
    task_object = GreatExpectationsTask(
        name="test3",
        datasource_name="data",
        inputs=kwtypes(data=str),
        expectation_suite_name="test.demo",
        data_connector_name="my_data_connector",
        task_config=BatchRequestConfig(
            data_connector_query={
                "batch_filter_parameters": {
                    "year": "2020",
                },
            }
        ),
    )

    # Capture IndexError
    with pytest.raises(InvalidBatchRequestError):
        task_object(data="my_assets")


def test_ge_runtimebatchrequest_sqlite_config():
    task_object = GreatExpectationsTask(
        name="test4",
        datasource_name="sqlite_data",
        inputs=kwtypes(dataset=str),
        expectation_suite_name="sqlite.movies",
        data_connector_name="sqlite_data_connector",
        data_asset_name="sqlite_data",
        task_config=BatchRequestConfig(
            batch_identifiers={
                "pipeline_stage": "validation",
            },
        ),
    )

    @workflow
    def runtime_sqlite_wf():
        task_object(dataset="SELECT * FROM movies")

    runtime_sqlite_wf()


def test_ge_runtimebatchrequest_pandas_config():
    task_object = GreatExpectationsTask(
        name="test5",
        datasource_name="my_pandas_datasource",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="test.demo",
        data_connector_name="my_runtime_data_connector",
        data_asset_name="pandas_data",
        task_config=BatchRequestConfig(
            batch_identifiers={
                "pipeline_stage": "validation",
            },
        ),
    )

    @workflow
    def runtime_pandas_wf(df: pd.DataFrame):
        task_object(dataset=df)

    runtime_pandas_wf(df=pd.read_csv("data/yellow_tripdata_sample_2019-01.csv"))


def test_ge_with_task():
    task_object = GreatExpectationsTask(
        name="test6",
        datasource_name="data",
        inputs=kwtypes(dataset=str),
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )

    @task
    def my_task(csv_file: str) -> int:
        df = pd.read_csv(os.path.join("data", csv_file))
        return df.shape[0]

    @workflow
    def valid_wf(dataset: str = "yellow_tripdata_sample_2019-01.csv") -> int:
        task_object(dataset=dataset)
        return my_task(csv_file=dataset)

    @workflow
    def invalid_wf(dataset: str = "yellow_tripdata_sample_2019-02.csv") -> int:
        task_object(dataset=dataset)
        return my_task(csv_file=dataset)

    valid_result = valid_wf()
    assert valid_result == 10000

    with pytest.raises(ValidationError, match=r".*passenger_count -> expect_column_min_to_be_between.*"):
        invalid_wf()


def test_ge_workflow():
    task_object = GreatExpectationsTask(
        name="test7",
        datasource_name="data",
        inputs=kwtypes(dataset=str),
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
    )

    @workflow
    def valid_wf(dataset: str = "yellow_tripdata_sample_2019-01.csv") -> None:
        task_object(dataset=dataset)

    valid_wf()


def test_ge_checkpoint_params():
    task_object = GreatExpectationsTask(
        name="test8",
        datasource_name="data",
        inputs=kwtypes(dataset=str),
        expectation_suite_name="test.demo",
        data_connector_name="data_example_data_connector",
        checkpoint_params={
            "site_names": ["local_site"],
        },
    )

    task_object(dataset="yellow_tripdata_sample_2019-01.csv")


def test_ge_remote_flytefile():
    task_object = GreatExpectationsTask(
        name="test9",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteFile),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    task_object(
        dataset="https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv"
    )


def test_ge_remote_flytefile_with_task():
    task_object = GreatExpectationsTask(
        name="test10",
        datasource_name="data",
        inputs=kwtypes(dataset=CSVFile),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @task
    def my_task(dataset: CSVFile) -> int:
        return len(pd.read_csv(dataset))

    @workflow
    def my_wf(dataset: CSVFile) -> int:
        task_object(dataset=dataset)
        return my_task(dataset=dataset)

    result = my_wf(
        dataset="https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv"
    )
    assert result == 10000


def test_ge_remote_flytefile_workflow():
    task_object = GreatExpectationsTask(
        name="test11",
        datasource_name="data",
        inputs=kwtypes(dataset=CSVFile),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @workflow
    def valid_wf(
        dataset: CSVFile = "https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv",
    ) -> None:
        task_object(dataset=dataset)

    valid_wf()


def test_ge_flytefile_workflow():
    task_object = GreatExpectationsTask(
        name="test12",
        datasource_name="data",
        inputs=kwtypes(dataset=CSVFile),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @workflow
    def valid_wf(
        dataset: CSVFile = "data/yellow_tripdata_sample_2019-01.csv",
    ) -> None:
        task_object(dataset=dataset)

    valid_wf()


def test_ge_flytefile_multiple_args():
    task_object_one = GreatExpectationsTask(
        name="test13",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteFile),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )
    task_object_two = GreatExpectationsTask(
        name="test14",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteFile),
        expectation_suite_name="test1.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp",
    )

    @task
    def get_file_name(dataset_one: FlyteFile, dataset_two: FlyteFile) -> typing.Tuple[int, int]:
        df_one = pd.read_csv(os.path.join("data", dataset_one))
        df_two = pd.read_csv(os.path.join("data", dataset_two))
        return len(df_one), len(df_two)

    @workflow
    def wf(
        dataset_one: FlyteFile = "https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-01.csv",
        dataset_two: FlyteFile = "https://raw.githubusercontent.com/superconductive/ge_tutorials/main/data/yellow_tripdata_sample_2019-02.csv",
    ) -> typing.Tuple[int, int]:
        task_object_one(dataset=dataset_one)
        task_object_two(dataset=dataset_two)
        return get_file_name(dataset_one=dataset_one, dataset_two=dataset_two)

    assert wf() == (10000, 10000)


def test_ge_flyteschema():
    task_object = GreatExpectationsTask(
        name="test15",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp/test.parquet",
    )

    df = pd.read_csv("data/yellow_tripdata_sample_2019-01.csv")
    task_object(dataset=df)


def test_ge_flyteschema_with_task():
    task_object = GreatExpectationsTask(
        name="test16",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp/test1.parquet",
    )

    @task
    def my_task(dataframe: pd.DataFrame) -> int:
        return dataframe.shape[0]

    @workflow
    def valid_wf(dataframe: pd.DataFrame) -> int:
        task_object(dataset=dataframe)
        return my_task(dataframe=dataframe)

    df = pd.read_csv("data/yellow_tripdata_sample_2019-01.csv")
    result = valid_wf(dataframe=df)
    assert result == 10000


def test_ge_flyteschema_sqlite():
    task_object = GreatExpectationsTask(
        name="test17",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="sqlite.movies",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp/test1.parquet",
    )

    @workflow
    def my_wf(dataset: FlyteSchema):
        task_object(dataset=dataset)

    con = sqlite3.connect(os.path.join("data", "movies.sqlite"))
    df = pd.read_sql_query("SELECT * FROM movies", con)
    con.close()
    my_wf(dataset=df)


def test_ge_flyteschema_workflow():
    task_object = GreatExpectationsTask(
        name="test18",
        datasource_name="data",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="test.demo",
        data_connector_name="data_flytetype_data_connector",
        local_file_path="/tmp/test1.parquet",
    )

    @workflow
    def my_wf(dataframe: pd.DataFrame):
        task_object(dataset=dataframe)

    df = pd.read_csv("data/yellow_tripdata_sample_2019-01.csv")
    my_wf(dataframe=df)


def test_ge_runtimebatchrequest_pyspark_config():
    task_object = GreatExpectationsTask(
        name="test19",
        datasource_name="my_pyspark_datasource",
        inputs=kwtypes(dataset=FlyteSchema),
        expectation_suite_name="test.demo_pyspark",
        data_connector_name="pyspark_runtime_data_connector",
        data_asset_name="pyspark_data",
        task_config=BatchRequestConfig(
            batch_identifiers={
                "pipeline_stage_pyspark": "validation",
            },
        ),
    )

    @task(task_config=Spark())
    def read_and_test():
        spark = flytekit.current_context().spark_session
        data_df = spark.read.option("inferSchema", "true").csv("data/yellow_tripdata_sample_2019-01.csv", header="true")
        task_object(dataset=data_df)

    @workflow
    def runtime_pandas_wf():
        read_and_test()

    runtime_pandas_wf()
