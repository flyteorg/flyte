from unittest.mock import Mock

import flytekit
from flytekit.clients.auth.default_html import get_default_success_html


def test_default_html():
    assert (
        get_default_success_html("flyte.org")
        == """
<html>
    <head>
            <title>OAuth2 Authentication Success</title>
    </head>
    <body>
            <h1>Successfully logged into flyte.org</h1>
            <img height="100" src="https://artwork.lfaidata.foundation/projects/flyte/horizontal/color/flyte-horizontal-color.svg" alt="Flyte login"></img>
    </body>
</html>
"""
    )  # noqa


def test_default_html_plugin(monkeypatch):
    def get_auth_success_html(endpoint):
        return f"<html><head><title>Successful Auth into {endpoint}!</title></head></html>"

    plugin_mock = Mock()
    plugin_mock.get_auth_success_html.side_effect = get_auth_success_html
    mock_global_plugin = {"plugin": plugin_mock}
    monkeypatch.setattr(flytekit.configuration.plugin, "_GLOBAL_CONFIG", mock_global_plugin)

    assert get_default_success_html("flyte.org") == get_auth_success_html("flyte.org")
