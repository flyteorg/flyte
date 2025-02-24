import os

from click.testing import CliRunner

from flytekit.clis.sdk_in_container import pyflyte
from flytekit.image_spec.image_spec import ImageBuildEngine

WORKFLOW_FILE = os.path.join(os.path.dirname(os.path.realpath(__file__)), "image_spec_wf.py")


def test_build(mock_image_spec_builder):
    ImageBuildEngine.register("test", mock_image_spec_builder)
    runner = CliRunner()
    result = runner.invoke(pyflyte.main, ["build", "--fast", WORKFLOW_FILE, "wf"])
    assert result.exit_code == 0

    result = runner.invoke(pyflyte.main, ["build", WORKFLOW_FILE, "wf"])
    assert result.exit_code == 0

    result = runner.invoke(pyflyte.main, ["build", WORKFLOW_FILE, "wf"])
    assert result.exit_code == 0

    result = runner.invoke(pyflyte.main, ["build", "--help"])
    assert result.exit_code == 0

    result = runner.invoke(pyflyte.main, ["build", "../", "wf"])
    assert result.exit_code == 1
