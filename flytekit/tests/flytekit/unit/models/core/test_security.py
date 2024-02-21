from unittest.mock import Mock

import flytekit.configuration.plugin
from flytekit.models.security import Secret


def test_secret():
    obj = Secret("grp", "key")
    obj2 = Secret.from_flyte_idl(obj.to_flyte_idl())
    assert obj2.key == "key"
    assert obj2.group_version is None

    obj = Secret("grp", group_version="v1")
    obj2 = Secret.from_flyte_idl(obj.to_flyte_idl())
    assert obj2.key is None
    assert obj2.group_version == "v1"


def test_secret_no_group(monkeypatch):
    plugin_mock = Mock()
    plugin_mock.secret_requires_group.return_value = False
    mock_global_plugin = {"plugin": plugin_mock}
    monkeypatch.setattr(flytekit.configuration.plugin, "_GLOBAL_CONFIG", mock_global_plugin)

    s = Secret(key="key")
    assert s.group is None
