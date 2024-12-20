import os

from flytekitplugins.dbt.error import DBTHandledError, DBTUnhandledError
from flytekitplugins.dbt.schema import (
    DBTFreshnessInput,
    DBTFreshnessOutput,
    DBTRunInput,
    DBTRunOutput,
    DBTTestInput,
    DBTTestOutput,
)
from flytekitplugins.dbt.util import run_cli

from flytekit import kwtypes
from flytekit.core.interface import Interface
from flytekit.core.python_function_task import PythonInstanceTask
from flytekit.loggers import logger

SUCCESS = 0
HANDLED_ERROR_CODE = 1
UNHANDLED_ERROR_CODE = 2


class DBTRun(PythonInstanceTask):
    """
    Execute DBT Run CLI command.

    The task will execute ``dbt run`` CLI command in a subprocess.
    Input from :class:`flytekitplugins.dbt.schema.DBTRunInput` will be converted into the corresponding CLI flags
    and stored in :class:`flytekitplugins.dbt.schema.DBTRunOutput`'s command.

    Parameters
    ----------
    name : str
        Task name.
    """

    def __init__(
        self,
        name: str,
        **kwargs,
    ):
        super(DBTRun, self).__init__(
            task_type="dbt-run",
            name=name,
            task_config=None,
            interface=Interface(inputs=kwtypes(input=DBTRunInput), outputs=kwtypes(output=DBTRunOutput)),
            **kwargs,
        )

    def execute(self, **kwargs) -> DBTRunOutput:
        """
        This method will be invoked to execute the task.

        Example
        -------
        ::

            dbt_run_task = DBTRun(name="test-task")

            @workflow
            def my_workflow() -> DBTRunOutput:
                return dbt_run_task(
                    input=DBTRunInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                    )
                )


        Parameters
        ----------
        input : DBTRunInput
            DBT run input.

        Returns
        -------
        DBTRunOutput
            DBT run output.

        Raises
        ------
        DBTHandledError
            If the ``dbt run`` command returns ``exit code 1``.
        DBTUnhandledError
            If the ``dbt run`` command returns ``exit code 2``.
        """

        task_input: DBTRunInput = kwargs["input"]

        args = task_input.to_args()
        cmd = ["dbt", "--log-format", "json", "run"] + args
        full_command = " ".join(cmd)

        logger.info(f"Executing command: {full_command}")
        exit_code, logs = run_cli(cmd)
        logger.info(f"dbt exited with return code {exit_code}")

        if exit_code == HANDLED_ERROR_CODE and not task_input.ignore_handled_error:
            raise DBTHandledError(f"handled error while executing {full_command}", logs)

        if exit_code == UNHANDLED_ERROR_CODE:
            raise DBTUnhandledError(f"unhandled error while executing {full_command}", logs)

        output_dir = os.path.join(task_input.project_dir, task_input.output_path)
        run_result_path = os.path.join(output_dir, "run_results.json")
        with open(run_result_path) as file:
            run_result = file.read()

        # read manifest.json
        manifest_path = os.path.join(output_dir, "manifest.json")
        with open(manifest_path) as file:
            manifest = file.read()

        return DBTRunOutput(
            command=full_command,
            exit_code=exit_code,
            raw_run_result=run_result,
            raw_manifest=manifest,
        )


class DBTTest(PythonInstanceTask):
    """Execute DBT Test CLI command

    The task will execute ``dbt test`` CLI command in a subprocess.
    Input from :class:`flytekitplugins.dbt.schema.DBTTestInput` will be converted into the corresponding CLI flags
    and stored in :class:`flytekitplugins.dbt.schema.DBTTestOutput`'s command.

    Parameters
    ----------
    name : str
        Task name.
    """

    def __init__(
        self,
        name: str,
        **kwargs,
    ):
        super(DBTTest, self).__init__(
            task_type="dbt-test",
            name=name,
            task_config=None,
            interface=Interface(
                inputs={
                    "input": DBTTestInput,
                },
                outputs={"output": DBTTestOutput},
            ),
            **kwargs,
        )

    def execute(self, **kwargs) -> DBTTestOutput:
        """
        This method will be invoked to execute the task.

        Example
        -------
        ::

            dbt_test_task = DBTTest(name="test-task")

            @workflow
            def my_workflow() -> DBTTestOutput:
                # run all models
                dbt_test_task(
                    input=DBTTestInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                    )
                )

                # run singular test only
                dbt_test_task(
                    input=DBTTestInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                        select=["test_type:singular"],
                    )
                )

                # run both singular and generic test
                return dbt_test_task(
                    input=DBTTestInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                        select=["test_type:singular", "test_type:generic"],
                    )
                )


        Parameters
        ----------
        input : DBTTestInput
            DBT test input

        Returns
        -------
        DBTTestOutput
            DBT test output

        Raises
        ------
        DBTHandledError
            If the ``dbt test`` command returns ``exit code 1``.
        DBTUnhandledError
            If the ``dbt test`` command returns ``exit code 2``.
        """

        task_input: DBTTestInput = kwargs["input"]

        args = task_input.to_args()
        cmd = ["dbt", "--log-format", "json", "test"] + args
        full_command = " ".join(cmd)

        logger.info(f"Executing command: {full_command}")
        exit_code, logs = run_cli(cmd)
        logger.info(f"dbt exited with return code {exit_code}")

        if exit_code == HANDLED_ERROR_CODE and not task_input.ignore_handled_error:
            raise DBTHandledError(f"handled error while executing {full_command}", logs)

        if exit_code == UNHANDLED_ERROR_CODE:
            raise DBTUnhandledError(f"unhandled error while executing {full_command}", logs)

        output_dir = os.path.join(task_input.project_dir, task_input.output_path)
        run_result_path = os.path.join(output_dir, "run_results.json")
        with open(run_result_path) as file:
            run_result = file.read()

        # read manifest.json
        manifest_path = os.path.join(output_dir, "manifest.json")
        with open(manifest_path) as file:
            manifest = file.read()

        return DBTTestOutput(
            command=full_command,
            exit_code=exit_code,
            raw_run_result=run_result,
            raw_manifest=manifest,
        )


class DBTFreshness(PythonInstanceTask):
    """Execute DBT Freshness CLI command

    The task will execute ``dbt freshness`` CLI command in a subprocess.
    Input from :class:`flytekitplugins.dbt.schema.DBTFreshnessInput` will be converted into the corresponding CLI flags
    and stored in :class:`flytekitplugins.dbt.schema.DBTFreshnessOutput`'s command.

    Parameters
    ----------
    name : str
        Task name.
    """

    def __init__(
        self,
        name: str,
        **kwargs,
    ):
        super(DBTFreshness, self).__init__(
            task_type="dbt-freshness",
            name=name,
            task_config=None,
            interface=Interface(
                inputs={
                    "input": DBTFreshnessInput,
                },
                outputs={"output": DBTFreshnessOutput},
            ),
            **kwargs,
        )

    def execute(self, **kwargs) -> DBTFreshnessOutput:
        """
        This method will be invoked to execute the task.

        Example
        -------
        ::

            dbt_freshness_task = DBTFreshness(name="freshness-task")

            @workflow
            def my_workflow() -> DBTFreshnessOutput:
                # run all models
                dbt_freshness_task(
                    input=DBTFreshnessInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                    )
                )

                # run singular freshness only
                dbt_freshness_task(
                    input=DBTFreshnessInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                        select=["test_type:singular"],
                    )
                )

                # run both singular and generic freshness
                return dbt_freshness_task(
                    input=DBTFreshnessInput(
                        project_dir="tests/jaffle_shop",
                        profiles_dir="tests/jaffle_shop/profiles",
                        profile="jaffle_shop",
                        select=["test_type:singular", "test_type:generic"],
                    )
                )


        Parameters
        ----------
        input : DBTFreshnessInput
            DBT freshness input

        Returns
        -------
        DBTFreshnessOutput
            DBT freshness output

        Raises
        ------
        DBTHandledError
            If the ``dbt source freshness`` command returns ``exit code 1``.
        DBTUnhandledError
            If the ``dbt source freshness`` command returns ``exit code 2``.
        """

        task_input: DBTFreshnessInput = kwargs["input"]

        args = task_input.to_args()
        cmd = ["dbt", "--log-format", "json", "source", "freshness"] + args
        full_command = " ".join(cmd)

        logger.info(f"Executing command: {full_command}")
        exit_code, logs = run_cli(cmd)
        logger.info(f"dbt exited with return code {exit_code}")

        if exit_code == HANDLED_ERROR_CODE and not task_input.ignore_handled_error:
            raise DBTHandledError(f"handled error while executing {full_command}", logs)

        if exit_code == UNHANDLED_ERROR_CODE:
            raise DBTUnhandledError(f"unhandled error while executing {full_command}", logs)

        output_dir = os.path.join(task_input.project_dir, task_input.output_path)
        sources_path = os.path.join(output_dir, "sources.json")
        with open(sources_path) as file:
            sources = file.read()

        return DBTFreshnessOutput(command=full_command, exit_code=exit_code, raw_sources=sources)
