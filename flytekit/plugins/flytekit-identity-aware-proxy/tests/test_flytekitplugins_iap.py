import uuid
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

import click
import jwt
import pytest
from click.testing import CliRunner
from flytekitplugins.identity_aware_proxy.cli import cli, get_gcp_secret_manager_secret, get_service_account_id_token
from google.api_core.exceptions import NotFound


def test_help() -> None:
    """Smoke test external command IAP ID token generator cli by printing help message."""
    runner = CliRunner()
    result = runner.invoke(cli, "--help")
    assert "Generate ID tokens" in result.output
    assert result.exit_code == 0

    result = runner.invoke(cli, ["generate-user-id-token", "--help"])
    assert "Generate a user account ID token" in result.output
    assert result.exit_code == 0

    result = runner.invoke(cli, ["generate-service-account-id-token", "--help"])
    assert "Generate a service account ID token" in result.output
    assert result.exit_code == 0


def test_get_gcp_secret_manager_secret():
    """Test retrieval of GCP secret manager secret."""
    project_id = "test_project"
    secret_id = "test_secret"
    version = "latest"
    expected_payload = "test_payload"

    mock_client = MagicMock()
    mock_client.access_secret_version.return_value.payload.data.decode.return_value = expected_payload
    with patch("google.cloud.secretmanager.SecretManagerServiceClient", return_value=mock_client):
        payload = get_gcp_secret_manager_secret(project_id, secret_id, version)
        assert payload == expected_payload

        name = f"projects/{project_id}/secrets/{secret_id}/versions/{version}"
        mock_client.access_secret_version.assert_called_once_with(name=name)


def test_get_gcp_secret_manager_secret_not_found():
    """Test retrieving non-existing secret from GCP secret manager."""
    project_id = "test_project"
    secret_id = "test_secret"
    version = "latest"

    mock_client = MagicMock()
    mock_client.access_secret_version.side_effect = NotFound("Secret not found")
    with patch("google.cloud.secretmanager.SecretManagerServiceClient", return_value=mock_client):
        with pytest.raises(click.BadParameter):
            get_gcp_secret_manager_secret(project_id, secret_id, version)


def create_mock_token(aud: str, expires_in: timedelta = None):
    """Create a mock JWT token with a certain audience, expiration time, and random JTI."""
    exp = datetime.utcnow() + expires_in
    jti = "test_token" + str(uuid.uuid4())
    payload = {"exp": exp, "aud": aud, "jti": jti}

    secret = "your-secret-key"
    algorithm = "HS256"

    return jwt.encode(payload, secret, algorithm=algorithm)


@patch("flytekitplugins.identity_aware_proxy.cli.id_token.fetch_id_token")
@patch("keyring.get_password")
@patch("keyring.set_password")
def test_sa_id_token_no_token_in_keyring(kr_set_password, kr_get_password, mock_fetch_id_token):
    """Test retrieval and caching of service account ID token when no token is stored in keyring yet."""
    test_audience = "test_audience"
    service_account_email = "default"

    # Start with a clean KeyringStore
    tmp_test_keyring_store = {}
    kr_get_password.side_effect = lambda service, user: tmp_test_keyring_store.get(service, {}).get(user, None)
    kr_set_password.side_effect = lambda service, user, pwd: tmp_test_keyring_store.update({service: {user: pwd}})

    mock_fetch_id_token.side_effect = lambda _, aud: create_mock_token(aud, expires_in=timedelta(hours=1))

    token = get_service_account_id_token(test_audience, service_account_email)

    assert jwt.decode(token.encode("utf-8"), options={"verify_signature": False})["aud"] == test_audience
    assert jwt.decode(token.encode("utf-8"), options={"verify_signature": False})["jti"].startswith("test_token")

    # Check that the token is cached in the KeyringStore
    second_token = get_service_account_id_token(test_audience, service_account_email)

    assert token == second_token


@patch("flytekitplugins.identity_aware_proxy.cli.id_token.fetch_id_token")
@patch("keyring.get_password")
@patch("keyring.set_password")
def test_sa_id_token_expired_token_in_keyring(kr_set_password, kr_get_password, mock_fetch_id_token):
    """Test that expired service account ID token in keyring is replaced with a new one."""
    test_audience = "test_audience"
    service_account_email = "default"

    # Start with an expired token in the KeyringStore
    expired_id_token = create_mock_token(test_audience, expires_in=timedelta(hours=-1))
    tmp_test_keyring_store = {test_audience + "-" + service_account_email: {"id_token": expired_id_token}}
    kr_get_password.side_effect = lambda service, user: tmp_test_keyring_store.get(service, {}).get(user, None)
    kr_set_password.side_effect = lambda service, user, pwd: tmp_test_keyring_store.update({service: {user: pwd}})

    mock_fetch_id_token.side_effect = lambda _, aud: create_mock_token(aud, expires_in=timedelta(hours=1))

    token = get_service_account_id_token(test_audience, service_account_email)

    assert token != expired_id_token
    assert jwt.decode(token.encode("utf-8"), options={"verify_signature": False})["aud"] == test_audience
    assert jwt.decode(token.encode("utf-8"), options={"verify_signature": False})["jti"].startswith("test_token")


@patch("flytekitplugins.identity_aware_proxy.cli.id_token.fetch_id_token")
@patch("keyring.get_password")
@patch("keyring.set_password")
def test_sa_id_token_switch_accounts(kr_set_password, kr_get_password, mock_fetch_id_token):
    """Test that caching works when switching service accounts."""
    test_audience = "test_audience"
    service_account_email = "default"
    service_account_other_email = "other"

    # Start with a clean KeyringStore
    tmp_test_keyring_store = {}
    kr_get_password.side_effect = lambda service, user: tmp_test_keyring_store.get(service, {}).get(user, None)
    kr_set_password.side_effect = lambda service, user, pwd: tmp_test_keyring_store.update({service: {user: pwd}})

    mock_fetch_id_token.side_effect = lambda _, aud: create_mock_token(aud, expires_in=timedelta(hours=1))

    default_token = get_service_account_id_token(test_audience, service_account_email)
    other_token = get_service_account_id_token(test_audience, service_account_other_email)

    assert default_token != other_token

    # Check that the tokens are cached in the KeyringStore
    new_default_token = get_service_account_id_token(test_audience, service_account_email)
    new_other_token = get_service_account_id_token(test_audience, service_account_other_email)

    assert default_token == new_default_token
    assert other_token == new_other_token
