import pytest
from click.testing import CliRunner
from mock import mock

from flytekit.clis.sdk_in_container import pyflyte
from flytekit.remote import FlyteRemote


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
@pytest.mark.parametrize(
    ("action", "expected_state"),
    [
        ("activate", "ACTIVE"),
        ("deactivate", "INACTIVE"),
    ],
)
def test_pyflyte_launchplan(mock_remote, action, expected_state):
    mock_remote.generate_console_url.return_value = "ex"
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(
            pyflyte.main,
            [
                "launchplan",
                f"--{action}",
                "-p",
                "flytesnacks",
                "-d",
                "development",
                "daily",
            ],
        )
        assert result.exit_code == 0
        assert f"Launchplan was set to {expected_state}: " in result.output
