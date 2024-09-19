import json
from dataclasses import dataclass
from typing import List, Optional

from dataclasses_json import DataClassJsonMixin


@dataclass
class BaseDBTInput(DataClassJsonMixin):
    """
    Base class for DBT Task Input.

    Attributes
    ----------
    project_dir : str
        Path to directory containing the DBT ``dbt_project.yml``.
    profiles_dir : str
        Path to directory containing the DBT ``profiles.yml``.
    profile : str
        Profile name to be used for the DBT task. It will override value in ``dbt_project.yml``.
    target : str
        Target to load for the given profile (default=None).
    output_path : str
        Path to directory where compiled files (e.g. models) will be written when running the task (default=target).
    ignore_handled_error : bool
        Ignore handled error (exit code = 1) returned by DBT, see https://docs.getdbt.com/reference/exit-codes (default=False).
    flags : dict
        Dictionary containing CLI flags to be added to the ``dbt run`` command (default=False).
    """

    project_dir: str
    profiles_dir: str
    profile: str
    target: str = None
    output_path: str = "target"
    ignore_handled_error: bool = False
    flags: dict = None

    def to_args(self) -> List[str]:
        """
        Convert the instance of BaseDBTInput into list of arguments.

        Returns
        -------
        List[str]
            List of arguments.
        """

        args = []
        args += ["--project-dir", self.project_dir]
        args += ["--profiles-dir", self.profiles_dir]
        args += ["--profile", self.profile]
        if self.target is not None:
            args += ["--target", self.target]

        if self.flags is not None:
            for flag, value in self.flags.items():
                if not value:
                    continue

                args.append(f"--{flag}")
                if isinstance(value, bool):
                    continue

                if isinstance(value, list):
                    args += value
                    continue

                if isinstance(value, dict):
                    args.append(json.dumps(value))
                    continue

                args.append(str(value))

        return args


@dataclass
class BaseDBTOutput(DataClassJsonMixin):
    """
    Base class for output of DBT task.

    Attributes
    ----------
    command : str
        Complete CLI command and flags that was executed by DBT Task.
    exit_code : int
        Exit code returned by DBT CLI.
    """

    command: str
    exit_code: int


@dataclass
class DBTRunInput(BaseDBTInput):
    """
    Input to DBT Run task.

    Attributes
    ----------
    select : List[str]
        List of model to be executed (default=None).
    exclude : List[str]
        List of model to be excluded (default=None).
    """

    select: Optional[List[str]] = None
    exclude: Optional[List[str]] = None

    def to_args(self) -> List[str]:
        """
        Convert the instance of BaseDBTInput into list of arguments.

        Returns
        -------
        List[str]
            List of arguments.
        """

        args = BaseDBTInput.to_args(self)
        if self.select is not None:
            args += ["--select"] + self.select

        if self.exclude is not None:
            args += ["--exclude"] + self.exclude

        return args


@dataclass
class DBTRunOutput(BaseDBTOutput):
    """
    Output of DBT run task.

    Attributes
    ----------
    raw_run_result : str
        Raw value of DBT's ``run_result.json``.
    raw_manifest : str
        Raw value of DBT's ``manifest.json``.
    """

    raw_run_result: str
    raw_manifest: str


@dataclass
class DBTTestInput(BaseDBTInput):
    """
    Input to DBT Test task.

    Attributes
    ----------
    select : List[str]
        List of model to be executed (default : None).
    exclude : List[str]
        List of model to be excluded (default : None).
    """

    select: Optional[List[str]] = None
    exclude: Optional[List[str]] = None

    def to_args(self) -> List[str]:
        """
        Convert the instance of DBTTestInput into list of arguments.

        Returns
        -------
        List[str]
            List of arguments.
        """

        args = BaseDBTInput.to_args(self)

        if self.select is not None:
            args += ["--select"] + self.select

        if self.exclude is not None:
            args += ["--exclude"] + self.exclude

        return args


@dataclass
class DBTTestOutput(BaseDBTOutput):
    """
    Output of DBT test task.

    Attributes
    ----------
    raw_run_result : str
        Raw value of DBT's ``run_result.json``.
    raw_manifest : str
        Raw value of DBT's ``manifest.json``.
    """

    raw_run_result: str
    raw_manifest: str


@dataclass
class DBTFreshnessInput(BaseDBTInput):
    """
    Input to DBT Freshness task.

    Attributes
    ----------
    select : List[str]
        List of model to be executed (default : None).
    exclude : List[str]
        List of model to be excluded (default : None).
    """

    select: Optional[List[str]] = None
    exclude: Optional[List[str]] = None

    def to_args(self) -> List[str]:
        """
        Convert the instance of DBTFreshnessInput into list of arguments.

        Returns
        -------
        List[str]
            List of arguments.
        """

        args = BaseDBTInput.to_args(self)

        if self.select is not None:
            args += ["--select"] + self.select

        if self.exclude is not None:
            args += ["--exclude"] + self.exclude

        return args


@dataclass
class DBTFreshnessOutput(BaseDBTOutput):
    """
    Output of DBT Freshness task.

    Attributes
    ----------
    raw_sources : str
        Raw value of DBT's ``sources.json``.
    """

    raw_sources: str
