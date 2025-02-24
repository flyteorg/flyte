import base64
import enum
import logging
import time
import typing
import urllib.parse
from dataclasses import dataclass
from datetime import datetime, timedelta

import requests

from flytekit import logger
from flytekit.clients.auth.exceptions import AuthenticationError, AuthenticationPending

utf_8 = "utf-8"

# Errors that Token endpoint will return
error_slow_down = "slow_down"
error_auth_pending = "authorization_pending"


# Grant Types
class GrantType(str, enum.Enum):
    CLIENT_CREDS = "client_credentials"
    DEVICE_CODE = "urn:ietf:params:oauth:grant-type:device_code"


@dataclass
class DeviceCodeResponse:
    """
    Response from device auth flow endpoint
    {'device_code': 'code',
         'user_code': 'BNDJJFXL',
         'verification_uri': 'url',
         'expires_in': 600,
         'interval': 5}
    """

    device_code: str
    user_code: str
    verification_uri: str
    expires_in: int
    interval: int

    @classmethod
    def from_json_response(cls, j: typing.Dict) -> "DeviceCodeResponse":
        return cls(
            device_code=j["device_code"],
            user_code=j["user_code"],
            verification_uri=j["verification_uri"],
            expires_in=j["expires_in"],
            interval=j["interval"],
        )


def get_basic_authorization_header(client_id: str, client_secret: str) -> str:
    """
    This function transforms the client id and the client secret into a header that conforms with http basic auth.
    It joins the id and the secret with a : then base64 encodes it, then adds the appropriate text. Secrets are
    first URL encoded to escape illegal characters.

    :param client_id: str
    :param client_secret: str
    :rtype: str
    """
    encoded = urllib.parse.quote_plus(client_secret)
    concatenated = "{}:{}".format(client_id, encoded)
    return "Basic {}".format(base64.b64encode(concatenated.encode(utf_8)).decode(utf_8))


def get_token(
    token_endpoint: str,
    scopes: typing.Optional[typing.List[str]] = None,
    authorization_header: typing.Optional[str] = None,
    client_id: typing.Optional[str] = None,
    device_code: typing.Optional[str] = None,
    audience: typing.Optional[str] = None,
    grant_type: GrantType = GrantType.CLIENT_CREDS,
    http_proxy_url: typing.Optional[str] = None,
    verify: typing.Optional[typing.Union[bool, str]] = None,
    session: typing.Optional[requests.Session] = None,
) -> typing.Tuple[str, int]:
    """
    :rtype: (Text,Int) The first element is the access token retrieved from the IDP, the second is the expiration
            in seconds
    """
    headers = {
        "Cache-Control": "no-cache",
        "Accept": "application/json",
        "Content-Type": "application/x-www-form-urlencoded",
    }
    if authorization_header:
        headers["Authorization"] = authorization_header
    body = {
        "grant_type": grant_type.value,
    }
    if client_id:
        body["client_id"] = client_id
    if device_code:
        body["device_code"] = device_code
    if scopes is not None:
        body["scope"] = " ".join(s.strip("' ") for s in scopes).strip("[]'")
    if audience:
        body["audience"] = audience

    proxies = {"https": http_proxy_url, "http": http_proxy_url} if http_proxy_url else None

    if not session:
        session = requests.Session()
    response = session.post(token_endpoint, data=body, headers=headers, proxies=proxies, verify=verify)

    if not response.ok:
        j = response.json()
        if "error" in j:
            err = j["error"]
            if err == error_auth_pending or err == error_slow_down:
                raise AuthenticationPending(f"Token not yet available, try again in some time {err}")
        logging.error("Status Code ({}) received from IDP: {}".format(response.status_code, response.text))
        raise AuthenticationError("Status Code ({}) received from IDP: {}".format(response.status_code, response.text))

    j = response.json()
    return j["access_token"], j["expires_in"]


def get_device_code(
    device_auth_endpoint: str,
    client_id: str,
    audience: typing.Optional[str] = None,
    scope: typing.Optional[typing.List[str]] = None,
    http_proxy_url: typing.Optional[str] = None,
    verify: typing.Optional[typing.Union[bool, str]] = None,
    session: typing.Optional[requests.Session] = None,
) -> DeviceCodeResponse:
    """
    Retrieves the device Authentication code that can be done to authenticate the request using a browser on a
    separate device
    """
    _scope = " ".join(s.strip("' ") for s in scope).strip("[]'") if scope is not None else ""
    payload = {"client_id": client_id, "scope": _scope, "audience": audience}
    proxies = {"https": http_proxy_url, "http": http_proxy_url} if http_proxy_url else None
    if not session:
        session = requests.Session()
    resp = session.post(device_auth_endpoint, payload, proxies=proxies, verify=verify)
    if not resp.ok:
        raise AuthenticationError(f"Unable to retrieve Device Authentication Code for {payload}, Reason {resp.reason}")
    return DeviceCodeResponse.from_json_response(resp.json())


def poll_token_endpoint(
    resp: DeviceCodeResponse,
    token_endpoint: str,
    client_id: str,
    audience: typing.Optional[str] = None,
    scopes: typing.Optional[str] = None,
    http_proxy_url: typing.Optional[str] = None,
    verify: typing.Optional[typing.Union[bool, str]] = None,
) -> typing.Tuple[str, int]:
    tick = datetime.now()
    interval = timedelta(seconds=resp.interval)
    end_time = tick + timedelta(seconds=resp.expires_in)
    while tick < end_time:
        try:
            access_token, expires_in = get_token(
                token_endpoint,
                grant_type=GrantType.DEVICE_CODE,
                client_id=client_id,
                audience=audience,
                scopes=scopes,
                device_code=resp.device_code,
                http_proxy_url=http_proxy_url,
                verify=verify,
            )
            print("Authentication successful!")
            return access_token, expires_in
        except AuthenticationPending:
            ...
        except Exception as e:
            logger.error("Authentication attempt failed: ", e)
            raise e
        print("Authentication Pending...")
        time.sleep(interval.total_seconds())
        tick = tick + interval
    raise AuthenticationError("Authentication failed!")
