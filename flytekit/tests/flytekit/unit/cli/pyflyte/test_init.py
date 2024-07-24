import tempfile

import pytest
from click.testing import CliRunner

from flytekit.clis.sdk_in_container import pyflyte


@pytest.mark.parametrize(
    "command",
    [
        ["example"],
        ["example", "--template", "basic-template-imagespec"],
        ["example", "--template", "bayesian-optimization"],
    ],
)
def test_pyflyte_init(command, monkeypatch: pytest.MonkeyPatch):
    tmp_dir = tempfile.mkdtemp()
    monkeypatch.chdir(tmp_dir)
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.init,
        command,
        catch_exceptions=True,
    )
    assert result.exit_code == 0
