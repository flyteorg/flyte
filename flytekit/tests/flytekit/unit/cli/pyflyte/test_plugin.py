import re
from unittest.mock import Mock, patch

import click
import pytest

from flytekit.configuration.plugin import FlytekitPlugin, FlytekitPluginProtocol, _get_plugin_from_entrypoint


@patch("flytekit.configuration.plugin.entry_points")
def test_get_plugin_default(entry_points):
    entry_points.side_effect = lambda *args, **kwargs: []

    default_plugin = _get_plugin_from_entrypoint()
    assert default_plugin is FlytekitPlugin


@patch("flytekit.configuration.plugin.entry_points")
def test_get_plugin_errors_with_multiple_plugins(entry_points, caplog):
    entry_1 = Mock()
    entry_1.name = "entry_1"

    entry_2 = Mock()
    entry_2.name = "entry_2"
    entry_points.side_effect = lambda *args, **kwargs: [entry_1, entry_2]

    msg = re.escape(
        "Multiple plugins installed: ['entry_1', 'entry_2']. flytekit only supports one installed plugin at a time."
    )
    with pytest.raises(ValueError, match=msg):
        _get_plugin_from_entrypoint()


class CustomPlugin(FlytekitPlugin):
    @staticmethod
    def configure_pyflyte_cli(main):
        """Make config hidden in main CLI."""
        for p in main.params:
            if p.name == "config":
                p.hidden = True
        return main


@click.command
@click.option(
    "-c",
    "--config",
    required=False,
    type=str,
    help="Path to config file for use within container",
)
def click_main(config):
    pass


@patch("flytekit.configuration.plugin.entry_points")
def test_get_plugin_custom(entry_points):
    entry_1 = Mock()
    entry_1.name.side_effect = "custom_plugin"
    entry_1.load.side_effect = lambda: CustomPlugin

    entry_points.side_effect = lambda *args, **kwargs: [entry_1]

    plugin = _get_plugin_from_entrypoint()
    assert plugin is CustomPlugin

    assert not click_main.params[0].hidden

    plugin.configure_pyflyte_cli(click_main)
    assert click_main.params[0].hidden


def test_plugin_follow_protocol():
    assert issubclass(FlytekitPlugin, FlytekitPluginProtocol)
