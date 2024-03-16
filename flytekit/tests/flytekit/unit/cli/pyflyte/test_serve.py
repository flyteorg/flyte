from click.testing import CliRunner

from flytekit.clis.sdk_in_container import pyflyte


def test_pyflyte_serve():
    runner = CliRunner()
    result = runner.invoke(pyflyte.main, ["serve", "agent", "--port", "0", "--timeout", "1"], catch_exceptions=False)
    assert result.exit_code == 0
