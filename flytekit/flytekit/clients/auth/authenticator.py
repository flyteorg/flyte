import logging
import subprocess
import typing
from abc import abstractmethod
from dataclasses import dataclass

import click
import requests

from . import token_client
from .auth_client import AuthorizationClient
from .exceptions import AccessTokenNotFoundError, AuthenticationError
from .keyring import Credentials, KeyringStore


@dataclass
class ClientConfig:
    """
    Client Configuration that is needed by the authenticator
    """

    token_endpoint: str
    authorization_endpoint: str
    redirect_uri: str
    client_id: str
    device_authorization_endpoint: typing.Optional[str] = None
    scopes: typing.List[str] = None
    header_key: str = "authorization"
    audience: typing.Optional[str] = None


class ClientConfigStore(object):
    """
    Client Config store retrieve client config. this can be done in multiple ways
    """

    @abstractmethod
    def get_client_config(self) -> ClientConfig:
        ...


class StaticClientConfigStore(ClientConfigStore):
    def __init__(self, cfg: ClientConfig):
        self._cfg = cfg

    def get_client_config(self) -> ClientConfig:
        return self._cfg


class Authenticator(object):
    """
    Base authenticator for all authentication flows
    """

    def __init__(
        self,
        endpoint: str,
        header_key: str,
        credentials: Credentials = None,
        http_proxy_url: typing.Optional[str] = None,
        verify: typing.Optional[typing.Union[bool, str]] = None,
    ):
        self._endpoint = endpoint
        self._creds = credentials
        self._header_key = header_key if header_key else "authorization"
        self._http_proxy_url = http_proxy_url
        self._verify = verify

    def get_credentials(self) -> Credentials:
        return self._creds

    def _set_credentials(self, creds):
        self._creds = creds

    def _set_header_key(self, h: str):
        self._header_key = h

    def fetch_grpc_call_auth_metadata(self) -> typing.Optional[typing.Tuple[str, str]]:
        if self._creds:
            return self._header_key, f"Bearer {self._creds.access_token}"
        return None

    @abstractmethod
    def refresh_credentials(self):
        ...


class PKCEAuthenticator(Authenticator):
    """
    This Authenticator encapsulates the entire PKCE flow and automatically opens a browser window for login

    For Auth0 - you will need to manually configure your config.yaml to include a scopes list of the syntax:
    admin.scopes: ["offline_access", "offline", "all", "openid"] and/or similar scopes in order to get the refresh token +
    caching. Otherwise, it will just receive the access token alone. Your FlyteCTL Helm config however should only
    contain ["offline", "all"] - as OIDC scopes are ungrantable in Auth0 customer APIs. They are simply requested
    for in the POST request during the token caching process.
    """

    def __init__(
        self,
        endpoint: str,
        cfg_store: ClientConfigStore,
        scopes: typing.Optional[typing.List[str]] = None,
        header_key: typing.Optional[str] = None,
        verify: typing.Optional[typing.Union[bool, str]] = None,
        session: typing.Optional[requests.Session] = None,
    ):
        """
        Initialize with default creds from KeyStore using the endpoint name
        """
        super().__init__(endpoint, header_key, KeyringStore.retrieve(endpoint), verify=verify)
        self._cfg_store = cfg_store
        self._auth_client = None
        self._scopes = scopes
        self._session = session or requests.Session()

    def _initialize_auth_client(self):
        if not self._auth_client:
            from .auth_client import _create_code_challenge, _generate_code_verifier

            code_verifier = _generate_code_verifier()
            code_challenge = _create_code_challenge(code_verifier)

            cfg = self._cfg_store.get_client_config()
            self._set_header_key(cfg.header_key)
            self._auth_client = AuthorizationClient(
                endpoint=self._endpoint,
                redirect_uri=cfg.redirect_uri,
                client_id=cfg.client_id,
                # Audience only needed for Auth0 - Taken from client config
                audience=cfg.audience,
                scopes=self._scopes or cfg.scopes,
                # self._scopes refers to flytekit.configuration.PlatformConfig (config.yaml)
                # cfg.scopes refers to PublicClientConfig scopes (can be defined in Helm deployments)
                auth_endpoint=cfg.authorization_endpoint,
                token_endpoint=cfg.token_endpoint,
                verify=self._verify,
                session=self._session,
                request_auth_code_params={
                    "code_challenge": code_challenge,
                    "code_challenge_method": "S256",
                },
                request_access_token_params={
                    "code_verifier": code_verifier,
                },
                refresh_access_token_params={},
                add_request_auth_code_params_to_request_access_token_params=True,
            )

    def refresh_credentials(self):
        """ """
        self._initialize_auth_client()
        if self._creds:
            """We have an access token so lets try to refresh it"""
            try:
                self._creds = self._auth_client.refresh_access_token(self._creds)
                if self._creds:
                    KeyringStore.store(self._creds)
                return
            except AccessTokenNotFoundError:
                logging.warning("Failed to refresh token. Kicking off a full authorization flow.")
                KeyringStore.delete(self._endpoint)

        self._creds = self._auth_client.get_creds_from_remote()
        KeyringStore.store(self._creds)


class CommandAuthenticator(Authenticator):
    """
    This Authenticator retrieves access_token using the provided command
    """

    def __init__(self, command: typing.List[str], header_key: str = None):
        self._cmd = command
        if not self._cmd:
            raise AuthenticationError("Command cannot be empty for command authenticator")
        super().__init__(None, header_key)

    def refresh_credentials(self):
        """
        This function is used when the configuration value for AUTH_MODE is set to 'external_process'.
        It reads an id token generated by an external process started by running the 'command'.
        """
        logging.debug("Starting external process to generate id token. Command {}".format(self._cmd))
        try:
            output = subprocess.run(self._cmd, capture_output=True, text=True, check=True)
        except subprocess.CalledProcessError as e:
            logging.error("Failed to generate token from command {}".format(self._cmd))
            raise AuthenticationError("Problems refreshing token with command: " + str(e))
        self._creds = Credentials(output.stdout.strip())


class ClientCredentialsAuthenticator(Authenticator):
    """
    This Authenticator uses ClientId and ClientSecret to authenticate
    """

    def __init__(
        self,
        endpoint: str,
        client_id: str,
        client_secret: str,
        cfg_store: ClientConfigStore,
        header_key: typing.Optional[str] = None,
        scopes: typing.Optional[typing.List[str]] = None,
        http_proxy_url: typing.Optional[str] = None,
        verify: typing.Optional[typing.Union[bool, str]] = None,
        audience: typing.Optional[str] = None,
        session: typing.Optional[requests.Session] = None,
    ):
        if not client_id or not client_secret:
            raise ValueError("Client ID and Client SECRET both are required.")
        cfg = cfg_store.get_client_config()
        self._token_endpoint = cfg.token_endpoint
        # Use scopes from `flytekit.configuration.PlatformConfig` if passed
        self._scopes = scopes or cfg.scopes
        self._client_id = client_id
        self._client_secret = client_secret
        self._audience = audience or cfg.audience
        self._session = session or requests.Session()
        super().__init__(endpoint, cfg.header_key or header_key, http_proxy_url=http_proxy_url, verify=verify)

    def refresh_credentials(self):
        """
        This function is used by the _handle_rpc_error() decorator, depending on the AUTH_MODE config object. This handler
        is meant for SDK use-cases of auth (like pyflyte, or when users call SDK functions that require access to Admin,
        like when waiting for another workflow to complete from within a task). This function uses basic auth, which means
        the credentials for basic auth must be present from wherever this code is running.

        """
        token_endpoint = self._token_endpoint
        scopes = self._scopes
        audience = self._audience

        # Note that unlike the Pkce flow, the client ID does not come from Admin.
        logging.debug(f"Basic authorization flow with client id {self._client_id} scope {scopes}")
        authorization_header = token_client.get_basic_authorization_header(self._client_id, self._client_secret)

        token, expires_in = token_client.get_token(
            token_endpoint=token_endpoint,
            authorization_header=authorization_header,
            http_proxy_url=self._http_proxy_url,
            verify=self._verify,
            scopes=scopes,
            audience=audience,
            session=self._session,
        )

        logging.info("Retrieved new token, expires in {}".format(expires_in))
        self._creds = Credentials(token)


class DeviceCodeAuthenticator(Authenticator):
    """
    This Authenticator implements the Device Code authorization flow useful for headless user authentication.

    Examples described
    - https://developer.okta.com/docs/guides/device-authorization-grant/main/
    - https://auth0.com/docs/get-started/authentication-and-authorization-flow/device-authorization-flow#device-flow
    """

    def __init__(
        self,
        endpoint: str,
        cfg_store: ClientConfigStore,
        header_key: typing.Optional[str] = None,
        audience: typing.Optional[str] = None,
        scopes: typing.Optional[typing.List[str]] = None,
        http_proxy_url: typing.Optional[str] = None,
        verify: typing.Optional[typing.Union[bool, str]] = None,
        session: typing.Optional[requests.Session] = None,
    ):
        cfg = cfg_store.get_client_config()
        self._audience = audience or cfg.audience
        self._client_id = cfg.client_id
        self._device_auth_endpoint = cfg.device_authorization_endpoint
        # Input param: scopes refers to flytekit.configuration.PlatformConfig (config.yaml)
        # cfg.scopes refers to PublicClientConfig scopes (can be defined in Helm deployments)
        # Use "scope" from object instantiation if value is not None - otherwise, default to cfg.scopes
        self._scopes = scopes or cfg.scopes
        self._token_endpoint = cfg.token_endpoint
        if self._device_auth_endpoint is None:
            raise AuthenticationError(
                "Device Authentication is not available on the Flyte backend / authentication server"
            )
        self._session = session or requests.Session()
        super().__init__(
            endpoint=endpoint,
            header_key=header_key or cfg.header_key,
            credentials=KeyringStore.retrieve(endpoint),
            http_proxy_url=http_proxy_url,
            verify=verify,
        )

    def refresh_credentials(self):
        resp = token_client.get_device_code(
            self._device_auth_endpoint,
            self._client_id,
            self._audience,
            self._scopes,
            self._http_proxy_url,
            self._verify,
            self._session,
        )
        text = f"To Authenticate, navigate in a browser to the following URL: {click.style(resp.verification_uri, fg='blue', underline=True)} and enter code: {click.style(resp.user_code, fg='blue')}"
        click.secho(text)
        try:
            # Currently the refresh token is not retrieved. We may want to add support for refreshTokens so that
            # access tokens can be refreshed for once authenticated machines
            token, expires_in = token_client.poll_token_endpoint(
                resp,
                self._token_endpoint,
                client_id=self._client_id,
                audience=self._audience,
                scopes=self._scopes,
                http_proxy_url=self._http_proxy_url,
                verify=self._verify,
            )
            self._creds = Credentials(access_token=token, expires_in=expires_in, for_endpoint=self._endpoint)
            KeyringStore.store(self._creds)
        except Exception:
            KeyringStore.delete(self._endpoint)
            raise
