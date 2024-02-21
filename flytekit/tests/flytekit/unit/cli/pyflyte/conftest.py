import os
import sys

import mock as _mock
import pytest
from click.testing import CliRunner

from flytekit import configuration as _config
from flytekit.clis.sdk_in_container import constants as _constants
from flytekit.clis.sdk_in_container import pyflyte as _pyflyte
from flytekit.tools import module_loader as _module_loader


def _fake_module_load(names):
    assert names == ("common.workflows",)
    from common.workflows import simple

    yield simple


@pytest.fixture(
    scope="function",
    params=[
        os.path.join(
            os.path.dirname(os.path.realpath(__file__)),
            "../../../common/configs/local.config",
        ),
        "/foo/bar",
        None,
    ],
)
def mock_ctx(request):
    with _config.TemporaryConfiguration(request.param):
        sys.path.append(os.path.join(os.path.dirname(os.path.realpath(__file__)), "../../.."))
        try:
            with _mock.patch("flytekit.tools.module_loader.iterate_modules") as mock_module_load:
                mock_module_load.side_effect = _fake_module_load
                ctx = _mock.MagicMock()
                ctx.obj = {
                    _constants.CTX_PACKAGES: ("common.workflows",),
                    _constants.CTX_PROJECT: "tests",
                    _constants.CTX_DOMAIN: "unit",
                    _constants.CTX_VERSION: "version",
                }
                yield ctx
        finally:
            sys.path.pop()


@pytest.fixture
def mock_clirunner(monkeypatch):
    def f(*args, **kwargs):
        runner = CliRunner()
        base_args = [
            "--pkgs",
            "common.workflows",
        ]

        result = runner.invoke(_pyflyte.main, base_args + list(args), **kwargs)

        if result.exception:
            raise result.exception

        return result

    tests_dir_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), "../../..")
    config_path = os.path.join(tests_dir_path, "common/configs/local.config")
    with _config.TemporaryConfiguration(config_path):
        monkeypatch.syspath_prepend(tests_dir_path)
        monkeypatch.setattr(_module_loader, "iterate_modules", _fake_module_load)
        yield f
