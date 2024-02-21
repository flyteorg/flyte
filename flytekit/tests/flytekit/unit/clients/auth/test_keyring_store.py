from unittest.mock import MagicMock, patch

from keyring.errors import NoKeyringError

from flytekit.clients.auth.keyring import Credentials, KeyringStore


@patch("keyring.get_password")
def test_keyring_store_get(kr_get_password: MagicMock):
    kr_get_password.return_value = "t"
    assert KeyringStore.retrieve("example1.com") is not None

    kr_get_password.side_effect = NoKeyringError()
    assert KeyringStore.retrieve("example2.com") is None


@patch("keyring.delete_password")
def test_keyring_store_delete(kr_del_password: MagicMock):
    kr_del_password.return_value = None
    assert KeyringStore.delete("example1.com") is None

    kr_del_password.side_effect = NoKeyringError()
    assert KeyringStore.delete("example2.com") is None


@patch("keyring.set_password")
def test_keyring_store_set(kr_set_password: MagicMock):
    kr_set_password.return_value = None
    assert KeyringStore.store(Credentials(access_token="a", refresh_token="r", for_endpoint="f"))

    kr_set_password.side_effect = NoKeyringError()
    assert KeyringStore.retrieve("example2.com") is None
