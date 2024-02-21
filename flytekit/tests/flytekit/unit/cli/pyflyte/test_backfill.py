from datetime import datetime, timedelta

import click
import pytest
from click.testing import CliRunner
from mock import mock

from flytekit.clis.sdk_in_container import pyflyte
from flytekit.clis.sdk_in_container.backfill import resolve_backfill_window
from flytekit.remote import FlyteRemote


def test_resolve_backfill_window():
    dt = datetime(2022, 12, 1, 8)
    window = timedelta(days=10)
    assert resolve_backfill_window(None, dt + window, window) == (dt, dt + window)
    assert resolve_backfill_window(dt, None, window) == (dt, dt + window)
    assert resolve_backfill_window(dt, dt + window) == (dt, dt + window)
    with pytest.raises(click.BadParameter):
        resolve_backfill_window()


@mock.patch("flytekit.configuration.plugin.FlyteRemote", spec=FlyteRemote)
def test_pyflyte_backfill(mock_remote):
    mock_remote.generate_console_url.return_value = "ex"
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(
            pyflyte.main,
            [
                "backfill",
                "--parallel",
                "-p",
                "flytesnacks",
                "-d",
                "development",
                "--from-date",
                "now",
                "--backfill-window",
                "5 day",
                "daily",
            ],
        )
        assert result.exit_code == 0
        assert "Execution launched" in result.output
