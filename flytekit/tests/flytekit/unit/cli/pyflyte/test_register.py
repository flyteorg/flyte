import os
import shutil
import subprocess

import mock
import pytest
from click.testing import CliRunner

from flytekit.clients.friendly import SynchronousFlyteClient
from flytekit.clis.sdk_in_container import pyflyte
from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context
from flytekit.configuration import Config
from flytekit.configuration.file import FLYTECTL_CONFIG_ENV_VAR
from flytekit.configuration.plugin import FlytekitPlugin
from flytekit.core import context_manager
from flytekit.core.context_manager import FlyteContextManager
from flytekit.remote.remote import FlyteRemote

sample_file_contents = """
from flytekit import task, workflow

@task(cache=True, cache_version="1", retries=3)
def sum(x: int, y: int) -> int:
    return x + y

@task(cache=True, cache_version="1", retries=3)
def square(z: int) -> int:
    return z*z

@workflow
def my_workflow(x: int, y: int) -> int:
    return sum(x=square(z=x), y=square(z=y))
"""

shell_task = """
from flytekit.extras.tasks.shell import ShellTask

t = ShellTask(
        name="test",
        script="echo 'Hello World'",
    )
"""


@pytest.fixture(scope="module")
def reset_flytectl_config_env_var() -> pytest.fixture():
    os.environ[FLYTECTL_CONFIG_ENV_VAR] = ""
    return os.environ[FLYTECTL_CONFIG_ENV_VAR]


@mock.patch("flytekit.configuration.plugin.FlyteRemote")
def test_get_remote(mock_remote, reset_flytectl_config_env_var):
    r = FlytekitPlugin.get_remote(None, "p", "d")
    assert r is not None
    mock_remote.assert_called_once_with(
        Config.for_sandbox(), default_project="p", default_domain="d", data_upload_location=None
    )


@mock.patch("flytekit.configuration.plugin.FlyteRemote")
def test_saving_remote(mock_remote):
    mock_context = mock.MagicMock
    mock_context.obj = {}
    get_and_save_remote_with_click_context(mock_context, "p", "d")
    assert mock_context.obj["flyte_remote"] is not None
    mock_remote.assert_called_once_with(
        Config.for_sandbox(), default_project="p", default_domain="d", data_upload_location=None
    )


def test_register_with_no_package_or_module_argument():
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(pyflyte.main, ["register"])
        assert result.exit_code == 1
        assert (
            "Missing argument 'PACKAGE_OR_MODULE...', at least one PACKAGE_OR_MODULE is required but multiple can be passed"
            in result.output
        )


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient", spec=SynchronousFlyteClient)
def test_register_with_no_output_dir_passed(mock_client, mock_remote):
    ctx = FlyteContextManager.current_context()
    mock_remote._client = mock_client
    mock_remote.return_value.context = ctx
    mock_remote.return_value._version_from_hash.return_value = "dummy_version_from_hash"
    mock_remote.return_value.fast_package.return_value = "dummy_md5_bytes", "dummy_native_url"
    runner = CliRunner()
    context_manager.FlyteEntities.entities.clear()
    with runner.isolated_filesystem():
        out = subprocess.run(["git", "init"], capture_output=True)
        assert out.returncode == 0
        os.makedirs("core1", exist_ok=True)
        with open(os.path.join("core1", "sample.py"), "w") as f:
            f.write(sample_file_contents)
            f.close()
        result = runner.invoke(pyflyte.main, ["register", "core1"])
        assert "Successfully registered 4 entities" in result.output
        shutil.rmtree("core1")


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient", spec=SynchronousFlyteClient)
def test_register_shell_task(mock_client, mock_remote):
    mock_remote._client = mock_client
    mock_remote.return_value._version_from_hash.return_value = "dummy_version_from_hash"
    mock_remote.return_value.fast_package.return_value = "dummy_md5_bytes", "dummy_native_url"
    runner = CliRunner()
    context_manager.FlyteEntities.entities.clear()
    with runner.isolated_filesystem():
        out = subprocess.run(["git", "init"], capture_output=True)
        assert out.returncode == 0
        os.makedirs("core2", exist_ok=True)
        with open(os.path.join("core2", "shell_task.py"), "w") as f:
            f.write(shell_task)
            f.close()
        result = runner.invoke(pyflyte.main, ["register", "core2"])
        assert "Successfully registered 2 entities" in result.output
        shutil.rmtree("core2")


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient", spec=SynchronousFlyteClient)
def test_non_fast_register(mock_client, mock_remote):
    ctx = FlyteContextManager.current_context()
    mock_remote.return_value.context = ctx
    mock_remote._client = mock_client
    runner = CliRunner()
    context_manager.FlyteEntities.entities.clear()
    with runner.isolated_filesystem():
        out = subprocess.run(["git", "init"], capture_output=True)
        assert out.returncode == 0
        os.makedirs("core2", exist_ok=True)
        with open(os.path.join("core2", "sample.py"), "w") as f:
            f.write(sample_file_contents)
            f.close()
        result = runner.invoke(pyflyte.main, ["register", "--non-fast", "--version", "a-version", "core2"])
        assert "Successfully registered 4 entities" in result.output
        shutil.rmtree("core2")


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient", spec=SynchronousFlyteClient)
def test_non_fast_register_require_version(mock_client, mock_remote):
    mock_remote._client = mock_client
    mock_remote.return_value._version_from_hash.return_value = "dummy_version_from_hash"
    mock_remote.return_value.upload_file.return_value = "dummy_md5_bytes", "dummy_native_url"
    runner = CliRunner()
    context_manager.FlyteEntities.entities.clear()
    with runner.isolated_filesystem():
        out = subprocess.run(["git", "init"], capture_output=True)
        assert out.returncode == 0
        os.makedirs("core3", exist_ok=True)
        with open(os.path.join("core3", "sample.py"), "w") as f:
            f.write(sample_file_contents)
            f.close()
        result = runner.invoke(pyflyte.main, ["register", "--non-fast", "core3"])
        assert result.exit_code == 1
        assert str(result.exception) == "Version is a required parameter in case --non-fast is specified."
        shutil.rmtree("core3")
