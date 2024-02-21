import mock

from flytekit.clis.flyte_cli.main import _get_client
from flytekit.configuration import AuthType, PlatformConfig


@mock.patch("flytekit.clients.friendly.SynchronousFlyteClient")
@mock.patch("click.get_current_context")
def test_get_client(click_current_ctx, mock_flyte_client):
    # This class helps in the process of overriding __getitem__ in an object.
    class FlexiMock(mock.MagicMock):
        def __init__(self, *args, **kwargs):
            super(FlexiMock, self).__init__(*args, **kwargs)
            self.__getitem__ = lambda obj, item: getattr(obj, item)  # set the mock for `[...]`

        def get(self, x, default=None):
            return getattr(self, x, default)

    click_current_ctx = mock.MagicMock
    obj_mock = FlexiMock(
        config=PlatformConfig(auth_mode=AuthType.EXTERNAL_PROCESS),
        cacert=None,
    )
    click_current_ctx.obj = obj_mock

    _ = _get_client(host="some-host:12345", insecure=False)

    expected_platform_config = PlatformConfig(
        endpoint="some-host:12345",
        insecure=False,
        auth_mode=AuthType.EXTERNAL_PROCESS,
    )

    mock_flyte_client.assert_called_with(expected_platform_config)
