import json
import subprocess
from unittest.mock import MagicMock, patch

import pytest

from flytekit.clients.auth.authenticator import (
    ClientConfig,
    ClientCredentialsAuthenticator,
    CommandAuthenticator,
    DeviceCodeAuthenticator,
    PKCEAuthenticator,
    StaticClientConfigStore,
)
from flytekit.clients.auth.exceptions import AuthenticationError
from flytekit.clients.auth.token_client import DeviceCodeResponse

ENDPOINT = "example.com"

client_config = ClientConfig(
    token_endpoint="token_endpoint",
    authorization_endpoint="auth_endpoint",
    redirect_uri="redirect_uri",
    client_id="client",
)

static_cfg_store = StaticClientConfigStore(client_config)


@patch("flytekit.clients.auth.authenticator.KeyringStore")
@patch("flytekit.clients.auth.auth_client.AuthorizationClient.get_creds_from_remote")
@patch("flytekit.clients.auth.auth_client.AuthorizationClient.refresh_access_token")
def test_pkce_authenticator(mock_refresh: MagicMock, mock_get_creds: MagicMock, mock_keyring: MagicMock):
    mock_keyring.retrieve.return_value = None
    authn = PKCEAuthenticator(ENDPOINT, static_cfg_store)
    assert authn._verify is None

    authn = PKCEAuthenticator(ENDPOINT, static_cfg_store, verify=False)
    assert authn._verify is False

    assert authn._creds is None
    assert authn._auth_client is None
    authn.refresh_credentials()
    assert authn._auth_client
    mock_get_creds.assert_called()
    mock_refresh.assert_not_called()
    mock_keyring.store.assert_called()

    authn.refresh_credentials()
    mock_refresh.assert_called()


@patch("subprocess.run")
def test_command_authenticator(mock_subprocess: MagicMock):
    with pytest.raises(AuthenticationError):
        authn = CommandAuthenticator(None)  # noqa

    authn = CommandAuthenticator(["echo"])

    authn.refresh_credentials()
    assert authn._creds
    mock_subprocess.assert_called()

    mock_subprocess.side_effect = subprocess.CalledProcessError(-1, ["x"])

    with pytest.raises(AuthenticationError):
        authn.refresh_credentials()


@patch("flytekit.clients.auth.token_client.requests.Session")
def test_client_creds_authenticator(mock_session):
    session = MagicMock()
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = json.loads("""{"access_token": "abc", "expires_in": 60}""")
    session.post.return_value = response
    mock_session.return_value = session

    authn = ClientCredentialsAuthenticator(
        ENDPOINT,
        client_id="client",
        client_secret="secret",
        cfg_store=static_cfg_store,
        http_proxy_url="https://my-proxy:31111",
    )

    authn.refresh_credentials()
    expected_scopes = static_cfg_store.get_client_config().scopes

    assert authn._creds
    assert authn._creds.access_token == "abc"
    assert authn._scopes == expected_scopes


@patch("flytekit.clients.auth.authenticator.KeyringStore")
@patch("flytekit.clients.auth.token_client.get_device_code")
@patch("flytekit.clients.auth.token_client.poll_token_endpoint")
def test_device_flow_authenticator(poll_mock: MagicMock, device_mock: MagicMock, mock_keyring: MagicMock):
    with pytest.raises(AuthenticationError):
        DeviceCodeAuthenticator(ENDPOINT, static_cfg_store, audience="x", verify=True)

    cfg_store = StaticClientConfigStore(
        ClientConfig(
            token_endpoint="token_endpoint",
            authorization_endpoint="auth_endpoint",
            redirect_uri="redirect_uri",
            client_id="client",
            device_authorization_endpoint="dev",
        )
    )
    authn = DeviceCodeAuthenticator(
        ENDPOINT, cfg_store, audience="x", http_proxy_url="http://my-proxy:9000", verify=False
    )

    device_mock.return_value = DeviceCodeResponse("x", "y", "s", 1000, 0)
    poll_mock.return_value = ("access", 100)
    authn.refresh_credentials()
    assert authn._creds


@patch("flytekit.clients.auth.token_client.requests.Session")
def test_client_creds_authenticator_with_custom_scopes(mock_session):
    expected_scopes = ["foo", "baz"]

    session = MagicMock()
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = json.loads("""{"access_token": "abc", "expires_in": 60}""")
    session.post.return_value = response
    mock_session.return_value = session

    authn = ClientCredentialsAuthenticator(
        ENDPOINT,
        client_id="client",
        client_secret="secret",
        cfg_store=static_cfg_store,
        scopes=expected_scopes,
        verify=True,
    )

    authn.refresh_credentials()

    assert authn._creds
    assert authn._creds.access_token == "abc"
    assert authn._scopes == expected_scopes
