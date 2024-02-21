import json
from unittest.mock import MagicMock, patch

import pytest

from flytekit.clients.auth.exceptions import AuthenticationError
from flytekit.clients.auth.token_client import (
    DeviceCodeResponse,
    error_auth_pending,
    get_basic_authorization_header,
    get_device_code,
    get_token,
    poll_token_endpoint,
)


def test_get_basic_authorization_header():
    header = get_basic_authorization_header("client_id", "abc")
    assert header == "Basic Y2xpZW50X2lkOmFiYw=="

    header = get_basic_authorization_header("client_id", "abc%%$?\\/\\/")
    assert header == "Basic Y2xpZW50X2lkOmFiYyUyNSUyNSUyNCUzRiU1QyUyRiU1QyUyRg=="


@patch("flytekit.clients.auth.token_client.requests.Session")
def test_get_token(mock_session):
    session = MagicMock()
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = json.loads("""{"access_token": "abc", "expires_in": 60}""")
    session.post.return_value = response
    mock_session.return_value = session
    access, expiration = get_token(
        "https://corp.idp.net", client_id="abc123", scopes=["my_scope"], http_proxy_url="http://proxy:3000", verify=True
    )
    assert access == "abc"
    assert expiration == 60


@patch("flytekit.clients.auth.token_client.requests.Session")
def test_get_device_code(mock_session):
    session = MagicMock()
    response = MagicMock()
    response.ok = False
    session.post.return_value = response
    mock_session.return_value = session
    with pytest.raises(AuthenticationError):
        get_device_code("test.com", "test", http_proxy_url="http://proxy:3000")

    response.ok = True
    response.json.return_value = {
        "device_code": "code",
        "user_code": "BNDJJFXL",
        "verification_uri": "url",
        "expires_in": 600,
        "interval": 5,
    }
    session.post.return_value = response
    c = get_device_code("test.com", "test", http_proxy_url="http://proxy:3000")
    assert c
    assert c.device_code == "code"


@patch("flytekit.clients.auth.token_client.requests.Session")
def test_poll_token_endpoint(mock_session):
    session = MagicMock()
    response = MagicMock()
    response.ok = False
    response.json.return_value = {"error": error_auth_pending}

    session.post.return_value = response
    mock_session.return_value = session

    r = DeviceCodeResponse(device_code="x", user_code="y", verification_uri="v", expires_in=1, interval=1)
    with pytest.raises(AuthenticationError):
        poll_token_endpoint(r, "test.com", "test", http_proxy_url="http://proxy:3000", verify=True)

    response = MagicMock()
    response.ok = True
    response.json.return_value = {"access_token": "abc", "expires_in": 60}
    session.post.return_value = response
    r = DeviceCodeResponse(device_code="x", user_code="y", verification_uri="v", expires_in=1, interval=0)
    t, e = poll_token_endpoint(r, "test.com", "test", http_proxy_url="http://proxy:3000", verify=True)

    assert t == "abc"
    assert e == 60
