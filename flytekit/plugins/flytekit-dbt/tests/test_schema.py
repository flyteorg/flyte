import shlex

import pytest
from flytekitplugins.dbt.schema import BaseDBTInput, DBTFreshnessInput, DBTRunInput, DBTTestInput

project_dir = "."
profiles_dir = "profiles"
profile_name = "development"


class TestBaseDBTInput:
    @pytest.mark.parametrize(
        "task_input,expected",
        [
            (
                BaseDBTInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name}",
            ),
            (
                BaseDBTInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    target="production",
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --target production",
            ),
            (
                BaseDBTInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    flags={"vars": {"var1": "val1", "var2": 2}},
                ),
                f"""--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --vars '{{"var1": "val1", "var2": 2}}'""",
            ),
            (
                BaseDBTInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    flags={"bool-flag": True},
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --bool-flag",
            ),
            (
                BaseDBTInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    flags={"list-flag": ["a", "b", "c"]},
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --list-flag a b c",
            ),
        ],
    )
    def test_to_args(self, task_input, expected):
        assert task_input.to_args() == shlex.split(expected)


class TestDBRunTInput:
    @pytest.mark.parametrize(
        "task_input,expected",
        [
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name}",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select model_a model_b",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select tag:nightly my_model finance.base.*",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select path:marts/finance,tag:nightly,config.materialized:table",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude model_a model_b",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude tag:nightly my_model finance.base.*",
            ),
            (
                DBTRunInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude path:marts/finance,tag:nightly,config.materialized:table",
            ),
        ],
    )
    def test_to_args(self, task_input, expected):
        assert task_input.to_args() == shlex.split(expected)


class TestDBTestInput:
    @pytest.mark.parametrize(
        "task_input,expected",
        [
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name}",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select test_type:singular",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select model_a model_b",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select tag:nightly my_model finance.base.*",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["tag:nightly", "my_model", "finance.base.*", "test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select tag:nightly my_model finance.base.* test_type:singular",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["path:marts/finance,tag:nightly,config.materialized:table,test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select path:marts/finance,tag:nightly,config.materialized:table,test_type:singular",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude model_a model_b",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude tag:nightly my_model finance.base.*",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude path:marts/finance,tag:nightly,config.materialized:table",
            ),
            (
                DBTTestInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["test_type:singular"],
                    exclude=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select test_type:singular --exclude path:marts/finance,tag:nightly,config.materialized:table",
            ),
        ],
    )
    def test_to_args(self, task_input, expected):
        assert task_input.to_args() == shlex.split(expected)


class TestDBFreshnessInput:
    @pytest.mark.parametrize(
        "task_input,expected",
        [
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name}",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select test_type:singular",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select model_a model_b",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select tag:nightly my_model finance.base.*",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["tag:nightly", "my_model", "finance.base.*", "test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select tag:nightly my_model finance.base.* test_type:singular",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["path:marts/finance,tag:nightly,config.materialized:table,test_type:singular"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select path:marts/finance,tag:nightly,config.materialized:table,test_type:singular",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["model_a", "model_b"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude model_a model_b",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["tag:nightly", "my_model", "finance.base.*"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude tag:nightly my_model finance.base.*",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    exclude=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --exclude path:marts/finance,tag:nightly,config.materialized:table",
            ),
            (
                DBTFreshnessInput(
                    project_dir=project_dir,
                    profiles_dir=profiles_dir,
                    profile=profile_name,
                    select=["test_type:singular"],
                    exclude=["path:marts/finance,tag:nightly,config.materialized:table"],
                ),
                f"--project-dir {project_dir} --profiles-dir {profiles_dir} --profile {profile_name} --select test_type:singular --exclude path:marts/finance,tag:nightly,config.materialized:table",
            ),
        ],
    )
    def test_to_args(self, task_input, expected):
        assert task_input.to_args() == shlex.split(expected)
