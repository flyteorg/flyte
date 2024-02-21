import logging
import os
import typing

import click
import jwt
from google.api_core.exceptions import NotFound
from google.auth import default
from google.auth.transport.requests import Request
from google.cloud import secretmanager
from google.oauth2 import id_token

from flytekit.clients.auth.auth_client import AuthorizationClient
from flytekit.clients.auth.authenticator import Authenticator
from flytekit.clients.auth.exceptions import AccessTokenNotFoundError
from flytekit.clients.auth.keyring import Credentials, KeyringStore

WEBAPP_CLIENT_ID_HELP = (
    "Webapp type OAuth 2.0 client ID used by the IAP. "
    "Typically in the form of `<xyz>.apps.googleusercontent.com`. "
    "Created when activating IAP for the Flyte deployment. "
    "https://cloud.google.com/iap/docs/enabling-kubernetes-howto#oauth-credentials"
)


class GCPIdentityAwareProxyAuthenticator(Authenticator):
    """
    This Authenticator encapsulates the entire OAauth 2.0 flow with GCP Identity Aware Proxy.

    The auth flow is described in https://cloud.google.com/iap/docs/authentication-howto#signing_in_to_the_application

    Automatically opens a browser window for login.
    """

    def __init__(
        self,
        audience: str,
        client_id: str,
        client_secret: str,
        verify: typing.Optional[typing.Union[bool, str]] = None,
    ):
        """
        Initialize with default creds from KeyStore using the audience name.
        """
        super().__init__(audience, "proxy-authorization", KeyringStore.retrieve(audience), verify=verify)
        self._auth_client = None

        self.audience = audience
        self.client_id = client_id
        self.client_secret = client_secret
        self.redirect_uri = "http://localhost:4444"

    def _initialize_auth_client(self):
        if not self._auth_client:
            self._auth_client = AuthorizationClient(
                endpoint=self.audience,
                # See step 3 in https://cloud.google.com/iap/docs/authentication-howto#signing_in_to_the_application
                auth_endpoint="https://accounts.google.com/o/oauth2/v2/auth",
                token_endpoint="https://oauth2.googleapis.com/token",
                # See step 3 in https://cloud.google.com/iap/docs/authentication-howto#signing_in_to_the_application
                scopes=["openid", "email"],
                client_id=self.client_id,
                redirect_uri=self.redirect_uri,
                verify=self._verify,
                # See step 3 in https://cloud.google.com/iap/docs/authentication-howto#signing_in_to_the_application
                request_auth_code_params={
                    "cred_ref": "true",
                    "access_type": "offline",
                },
                # See step 4 in https://cloud.google.com/iap/docs/authentication-howto#signing_in_to_the_application
                request_access_token_params={
                    "client_id": self.client_id,
                    "client_secret": self.client_secret,
                    "audience": self.audience,
                    "redirect_uri": self.redirect_uri,
                },
                # See https://cloud.google.com/iap/docs/authentication-howto#refresh_token
                refresh_access_token_params={
                    "client_secret": self.client_secret,
                    "audience": self.audience,
                },
            )

    def refresh_credentials(self):
        """Refresh the IAP credentials. If no credentials are found, it will kick off a full OAuth 2.0 authorization flow."""
        self._initialize_auth_client()
        if self._creds:
            """We have an id token so lets try to refresh it"""
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


def get_gcp_secret_manager_secret(project_id: str, secret_id: str, version: typing.Optional[str] = "latest"):
    """Retrieve secret from GCP secret manager."""
    client = secretmanager.SecretManagerServiceClient()
    name = f"projects/{project_id}/secrets/{secret_id}/versions/{version}"
    try:
        response = client.access_secret_version(name=name)
    except NotFound as e:
        raise click.BadParameter(e.message)
    payload = response.payload.data.decode("UTF-8")
    return payload


@click.group()
def cli():
    """Generate ID tokens for GCP Identity Aware Proxy (IAP)."""
    pass


@cli.command()
@click.option(
    "--desktop_client_id",
    type=str,
    default=None,
    required=True,
    help=(
        "Desktop type OAuth 2.0 client ID. Typically in the form of `<xyz>.apps.googleusercontent.com`. "
        "Create by following https://cloud.google.com/iap/docs/authentication-howto#setting_up_the_client_id"
    ),
)
@click.option(
    "--desktop_client_secret_gcp_secret_name",
    type=str,
    default=None,
    required=True,
    help=(
        "Name of a GCP secret manager secret containing the desktop type OAuth 2.0 client secret "
        "obtained together with desktop type OAuth 2.0 client ID."
    ),
)
@click.option(
    "--webapp_client_id",
    type=str,
    default=None,
    required=True,
    help=WEBAPP_CLIENT_ID_HELP,
)
@click.option(
    "--project",
    type=str,
    default=None,
    required=True,
    help="GCP project ID (in which `desktop_client_secret_gcp_secret_name` is saved).",
)
def generate_user_id_token(
    desktop_client_id: str, desktop_client_secret_gcp_secret_name: str, webapp_client_id: str, project: str
):
    """Generate a user account ID token for proxy-authorization with GCP Identity Aware Proxy."""
    desktop_client_secret = get_gcp_secret_manager_secret(project, desktop_client_secret_gcp_secret_name)

    iap_authenticator = GCPIdentityAwareProxyAuthenticator(
        audience=webapp_client_id,
        client_id=desktop_client_id,
        client_secret=desktop_client_secret,
    )
    try:
        iap_authenticator.refresh_credentials()
    except Exception as e:
        raise click.ClickException(f"Failed to obtain credentials for GCP Identity Aware Proxy (IAP): {e}")

    click.echo(iap_authenticator.get_credentials().id_token)


def get_service_account_id_token(audience: str, service_account_email: str) -> str:
    """Fetch an ID Token for the service account used by the current environment.

    Uses flytekit's KeyringStore to cache the ID token.

    This function acquires ID token from the environment in the following order.
    See https://google.aip.dev/auth/4110.

    1. If the environment variable ``GOOGLE_APPLICATION_CREDENTIALS`` is set
       to the path of a valid service account JSON file, then ID token is
       acquired using this service account credentials.
    2. If the application is running in Compute Engine, App Engine or Cloud Run,
       then the ID token are obtained from the metadata server.

    Args:
        audience (str): The audience that this ID token is intended for.
        service_account_email (str): The email address of the service account.
    """
    # Flytekit's KeyringStore, by default, uses the endpoint as the key to store the credentials
    # We use the audience and the service account email as the key
    audience_and_account_key = audience + "-" + service_account_email
    creds = KeyringStore.retrieve(audience_and_account_key)
    if creds:
        is_expired = False
        try:
            exp_margin = -300  # Generate a new token if it expires in less than 5 minutes
            jwt.decode(
                creds.id_token.encode("utf-8"),
                options={"verify_signature": False, "verify_exp": True},
                leeway=exp_margin,
            )
        except jwt.ExpiredSignatureError:
            is_expired = True

        if not is_expired:
            return creds.id_token

    token = id_token.fetch_id_token(Request(), audience)

    KeyringStore.store(Credentials(for_endpoint=audience_and_account_key, access_token="", id_token=token))
    return token


@cli.command()
@click.option(
    "--webapp_client_id",
    type=str,
    default=None,
    required=True,
    help=WEBAPP_CLIENT_ID_HELP,
)
@click.option(
    "--service_account_key",
    type=click.Path(exists=True, dir_okay=False),
    default=None,
    required=False,
    help=(
        "Path to a service account key file. Alternatively set the environment variable "
        "`GOOGLE_APPLICATION_CREDENTIALS` to the path of the service account key file. "
        "If not provided and in Compute Engine, App Engine, or Cloud Run, will retrieve "
        "the ID token from the metadata server."
    ),
)
def generate_service_account_id_token(webapp_client_id: str, service_account_key: str):
    """Generate a service account ID token for proxy-authorization with GCP Identity Aware Proxy."""
    if service_account_key:
        os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = service_account_key

    application_default_credentials, _ = default()
    token = get_service_account_id_token(webapp_client_id, application_default_credentials.service_account_email)
    click.echo(token)


if __name__ == "__main__":
    cli()
