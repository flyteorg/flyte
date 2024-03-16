import logging
import typing
from dataclasses import dataclass

import keyring as _keyring
from keyring.errors import NoKeyringError, PasswordDeleteError


@dataclass
class Credentials(object):
    """
    Stores the credentials together
    """

    access_token: str
    refresh_token: str = "na"
    for_endpoint: str = "flyte-default"
    expires_in: typing.Optional[int] = None
    id_token: typing.Optional[str] = None


class KeyringStore:
    """
    Methods to access Keyring Store.
    """

    _access_token_key = "access_token"
    _refresh_token_key = "refresh_token"
    _id_token_key = "id_token"

    @staticmethod
    def store(credentials: Credentials) -> Credentials:
        try:
            if credentials.refresh_token:
                _keyring.set_password(
                    credentials.for_endpoint,
                    KeyringStore._refresh_token_key,
                    credentials.refresh_token,
                )
            _keyring.set_password(
                credentials.for_endpoint,
                KeyringStore._access_token_key,
                credentials.access_token,
            )
            if credentials.id_token:
                _keyring.set_password(
                    credentials.for_endpoint,
                    KeyringStore._id_token_key,
                    credentials.id_token,
                )
        except NoKeyringError as e:
            logging.debug(f"KeyRing not available, tokens will not be cached. Error: {e}")
        return credentials

    @staticmethod
    def retrieve(for_endpoint: str) -> typing.Optional[Credentials]:
        try:
            refresh_token = _keyring.get_password(for_endpoint, KeyringStore._refresh_token_key)
            access_token = _keyring.get_password(for_endpoint, KeyringStore._access_token_key)
            id_token = _keyring.get_password(for_endpoint, KeyringStore._id_token_key)
        except NoKeyringError as e:
            logging.debug(f"KeyRing not available, tokens will not be cached. Error: {e}")
            return None

        if not access_token and not id_token:
            return None
        return Credentials(access_token, refresh_token, for_endpoint, id_token=id_token)

    @staticmethod
    def delete(for_endpoint: str):
        def _delete_key(key):
            try:
                _keyring.delete_password(for_endpoint, key)
            except PasswordDeleteError as e:
                logging.debug(f"Key {key} not found in key store, Ignoring. Error: {e}")
            except NoKeyringError as e:
                logging.debug(f"KeyRing not available, Key {key} deletion failed. Error: {e}")

        _delete_key(KeyringStore._access_token_key)
        _delete_key(KeyringStore._refresh_token_key)
        _delete_key(KeyringStore._id_token_key)
