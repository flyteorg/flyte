import configparser
import datetime
import os

import mock
import pytest
from pytimeparse.timeparse import timeparse

from flytekit.configuration import ConfigEntry, get_config_file, set_if_exists
from flytekit.configuration.file import LegacyConfigEntry, _exists
from flytekit.configuration.internal import Platform


def test_set_if_exists():
    d = {}
    d = set_if_exists(d, "k", None)
    assert len(d) == 0
    d = set_if_exists(d, "k", [])
    assert len(d) == 0
    d = set_if_exists(d, "k", "x")
    assert len(d) == 1
    assert d["k"] == "x"


@pytest.mark.parametrize(
    "data, expected",
    [
        [1, True],
        [1.0, True],
        ["foo", True],
        [True, True],
        [False, True],
        [[1], True],
        [{"k": "v"}, True],
        [None, False],
        [[], False],
        [{}, False],
    ],
)
def test_exists(data, expected):
    assert _exists(data) is expected


def test_get_config_file():
    c = get_config_file(None)
    assert c is None
    c = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/good.config"))
    assert c is not None
    assert c.legacy_config is not None

    with pytest.raises(configparser.Error):
        get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/bad.config"))


def test_config_entry_envvar():
    # Pytest feature
    c = ConfigEntry(LegacyConfigEntry("test", "op1", str))
    assert c.read() is None

    old_environ = dict(os.environ)
    os.environ["FLYTE_TEST_OP1"] = "xyz"
    assert c.read() == "xyz"
    os.environ = old_environ


def test_config_entry_file():
    # Pytest feature
    c = ConfigEntry(LegacyConfigEntry("platform", "url", str))
    assert c.read() is None

    cfg = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/good.config"))
    assert c.read(cfg) == "fakeflyte.com"

    c = ConfigEntry(LegacyConfigEntry("platform", "url2", str))  # Does not exist
    assert c.read(cfg) is None


def test_config_entry_precedence():
    # Pytest feature
    c = ConfigEntry(LegacyConfigEntry("platform", "url", str))
    assert c.read() is None

    old_environ = dict(os.environ)
    os.environ["FLYTE_PLATFORM_URL"] = "xyz"
    cfg = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/good.config"))
    assert c.read(cfg) == "xyz"
    # reset
    os.environ = old_environ


def test_config_entry_types():
    cfg = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/good.config"))

    l = ConfigEntry(LegacyConfigEntry("sdk", "workflow_packages", list))
    assert l.read(cfg) == ["this.module", "that.module"]

    s = ConfigEntry(LegacyConfigEntry("madeup", "string_value"))
    assert s.read(cfg) == "abc"

    i = ConfigEntry(LegacyConfigEntry("madeup", "int_value", int))
    assert i.read(cfg) == 3

    b = ConfigEntry(LegacyConfigEntry("madeup", "bool_value", bool))
    assert b.read(cfg) is False

    t = ConfigEntry(
        LegacyConfigEntry("madeup", "timedelta_value", datetime.timedelta),
        transform=lambda x: datetime.timedelta(seconds=timeparse(x)),
    )
    assert t.read(cfg) == datetime.timedelta(hours=20)


@mock.patch("flytekit.configuration.file.LegacyConfigEntry.read_from_file")
def test_env_var_bool_transformer(mock_file_read):
    mock_file_read.return_value = None
    test_env_var = "FLYTE_MADEUP_TEST_VAR_ABC123"
    b = ConfigEntry(LegacyConfigEntry("madeup", "test_var_abc123", bool))

    os.environ[test_env_var] = "FALSE"
    assert b.read() is False

    os.environ[test_env_var] = ""
    assert b.read() is False

    os.environ[test_env_var] = "1"
    assert b.read()

    os.environ[test_env_var] = "truee"
    assert b.read()
    # The above reads shouldn't have triggered the file read since the env var was set
    assert mock_file_read.call_count == 0

    del os.environ[test_env_var]

    cfg_file_mock = mock.MagicMock()
    cfg_file_mock.legacy_config.return_value = True
    assert b.read(cfg=cfg_file_mock) is None

    # The last read should've triggered the file read since now the env var is no longer set.
    assert mock_file_read.call_count == 1


def test_use_ssl():
    config_file = get_config_file(os.path.join(os.path.dirname(os.path.realpath(__file__)), "configs/good.config"))
    res = Platform.INSECURE.read(config_file)
    assert res is False
