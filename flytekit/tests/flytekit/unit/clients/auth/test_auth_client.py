import http.server as _BaseHTTPServer
import re
from multiprocessing import Queue as _Queue

from flytekit.clients.auth.auth_client import (
    EndpointMetadata,
    OAuthHTTPServer,
    _create_code_challenge,
    _generate_code_verifier,
    _generate_state_parameter,
)


def test_generate_code_verifier():
    verifier = _generate_code_verifier()
    assert verifier is not None
    assert 43 < len(verifier) < 128
    assert not re.search(r"[^a-zA-Z0-9_\-.~]+", verifier)


def test_generate_state_parameter():
    param = _generate_state_parameter()
    assert not re.search(r"[^a-zA-Z0-9-_.,]+", param)


def test_create_code_challenge():
    test_code_verifier = "test_code_verifier"
    assert _create_code_challenge(test_code_verifier) == "Qq1fGD0HhxwbmeMrqaebgn1qhvKeguQPXqLdpmixaM4"


def test_oauth_http_server():
    queue = _Queue()
    server = OAuthHTTPServer(
        ("localhost", 9000),
        remote_metadata=EndpointMetadata(endpoint="example.com"),
        request_handler_class=_BaseHTTPServer.BaseHTTPRequestHandler,
        queue=queue,
    )
    test_auth_code = "auth_code"
    server.handle_authorization_code(test_auth_code)
    auth_code = queue.get()
    assert test_auth_code == auth_code
