import os
import pathlib

import pytest
from flytekitplugins.dbt.error import DBTUnhandledError
from flytekitplugins.dbt.schema import (
    DBTFreshnessInput,
    DBTFreshnessOutput,
    DBTRunInput,
    DBTRunOutput,
    DBTTestInput,
    DBTTestOutput,
)
from flytekitplugins.dbt.task import DBTFreshness, DBTRun, DBTTest

from flytekit import workflow
from flytekit.tools.subprocess import check_call

DBT_PROJECT_DIR = str(pathlib.Path(os.path.dirname(os.path.realpath(__file__)), "testdata", "jaffle_shop"))
DBT_PROFILES_DIR = str(pathlib.Path(os.path.dirname(os.path.realpath(__file__)), "testdata", "profiles"))
DBT_PROFILE = "jaffle_shop"


@pytest.fixture(scope="module", autouse=True)
def prepare_db():
    # Ensure path to sqlite database file exists
    dbs_path = pathlib.Path(DBT_PROJECT_DIR, "dbs")
    dbs_path.mkdir(exist_ok=True, parents=True)
    database_file = pathlib.Path(dbs_path, "database_name.db")
    database_file.touch()

    # Seed the database
    check_call(
        [
            "dbt",
            "--log-format",
            "json",
            "seed",
            "--project-dir",
            DBT_PROJECT_DIR,
            "--profiles-dir",
            DBT_PROFILES_DIR,
            "--profile",
            DBT_PROFILE,
        ]
    )

    yield

    # Delete the database file
    database_file.unlink()


class TestDBTRun:
    def test_simple_task(self):
        dbt_run_task = DBTRun(
            name="test-task",
        )

        @workflow
        def my_workflow() -> DBTRunOutput:
            # run all models
            return dbt_run_task(
                input=DBTRunInput(
                    project_dir=DBT_PROJECT_DIR,
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                    select=["tag:something"],
                    exclude=["tag:something-else"],
                )
            )

        result = my_workflow()
        assert isinstance(result, DBTRunOutput)

    def test_incorrect_project_dir(self):
        dbt_run_task = DBTRun(
            name="test-task",
        )

        with pytest.raises(DBTUnhandledError):
            dbt_run_task(
                input=DBTRunInput(
                    project_dir=".",
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                )
            )

    def test_task_output(self):
        dbt_run_task = DBTRun(
            name="test-task",
        )

        output = dbt_run_task.execute(
            input=DBTRunInput(project_dir=DBT_PROJECT_DIR, profiles_dir=DBT_PROFILES_DIR, profile=DBT_PROFILE)
        )

        assert output.exit_code == 0
        assert (
            output.command
            == f"dbt --log-format json run --project-dir {DBT_PROJECT_DIR} --profiles-dir {DBT_PROFILES_DIR} --profile {DBT_PROFILE}"
        )

        with open(f"{DBT_PROJECT_DIR}/target/run_results.json", "r") as fp:
            exp_run_result = fp.read()
        assert output.raw_run_result == exp_run_result

        with open(f"{DBT_PROJECT_DIR}/target/manifest.json", "r") as fp:
            exp_manifest = fp.read()
        assert output.raw_manifest == exp_manifest


class TestDBTTest:
    def test_simple_task(self):
        dbt_test_task = DBTTest(
            name="test-task",
        )

        @workflow
        def test_workflow() -> DBTTestOutput:
            # run all tests
            return dbt_test_task(
                input=DBTTestInput(
                    project_dir=DBT_PROJECT_DIR,
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                )
            )

        assert isinstance(test_workflow(), DBTTestOutput)

    def test_incorrect_project_dir(self):
        dbt_test_task = DBTTest(
            name="test-task",
        )

        with pytest.raises(DBTUnhandledError):
            dbt_test_task(
                input=DBTTestInput(
                    project_dir=".",
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                )
            )

    def test_task_output(self):
        dbt_test_task = DBTTest(
            name="test-task",
        )

        output = dbt_test_task.execute(
            input=DBTTestInput(project_dir=DBT_PROJECT_DIR, profiles_dir=DBT_PROFILES_DIR, profile=DBT_PROFILE)
        )

        assert output.exit_code == 0
        assert (
            output.command
            == f"dbt --log-format json test --project-dir {DBT_PROJECT_DIR} --profiles-dir {DBT_PROFILES_DIR} --profile {DBT_PROFILE}"
        )

        with open(f"{DBT_PROJECT_DIR}/target/run_results.json", "r") as fp:
            exp_run_result = fp.read()
        assert output.raw_run_result == exp_run_result

        with open(f"{DBT_PROJECT_DIR}/target/manifest.json", "r") as fp:
            exp_manifest = fp.read()
        assert output.raw_manifest == exp_manifest


class TestDBTFreshness:
    def test_simple_task(self):
        dbt_freshness_task = DBTFreshness(
            name="test-task",
        )

        @workflow
        def my_workflow() -> DBTFreshnessOutput:
            # run all models
            return dbt_freshness_task(
                input=DBTFreshnessInput(
                    project_dir=DBT_PROJECT_DIR,
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                    select=["tag:something"],
                    exclude=["tag:something-else"],
                )
            )

        result = my_workflow()
        assert isinstance(result, DBTFreshnessOutput)

    def test_incorrect_project_dir(self):
        dbt_freshness_task = DBTFreshness(
            name="test-task",
        )

        with pytest.raises(DBTUnhandledError):
            dbt_freshness_task(
                input=DBTFreshnessInput(
                    project_dir=".",
                    profiles_dir=DBT_PROFILES_DIR,
                    profile=DBT_PROFILE,
                )
            )

    def test_task_output(self):
        dbt_freshness_task = DBTFreshness(
            name="test-task",
        )

        output = dbt_freshness_task.execute(
            input=DBTFreshnessInput(project_dir=DBT_PROJECT_DIR, profiles_dir=DBT_PROFILES_DIR, profile=DBT_PROFILE)
        )

        assert output.exit_code == 0
        assert (
            output.command
            == f"dbt --log-format json source freshness --project-dir {DBT_PROJECT_DIR} --profiles-dir {DBT_PROFILES_DIR} --profile {DBT_PROFILE}"
        )

        with open(f"{DBT_PROJECT_DIR}/target/sources.json", "r") as fp:
            exp_sources = fp.read()

        assert output.raw_sources == exp_sources
