from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class OAuth2MetadataRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class OAuth2MetadataResponse(_message.Message):
    __slots__ = ["issuer", "authorization_endpoint", "token_endpoint", "response_types_supported", "scopes_supported", "token_endpoint_auth_methods_supported", "jwks_uri", "code_challenge_methods_supported", "grant_types_supported", "device_authorization_endpoint"]
    ISSUER_FIELD_NUMBER: _ClassVar[int]
    AUTHORIZATION_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_TYPES_SUPPORTED_FIELD_NUMBER: _ClassVar[int]
    SCOPES_SUPPORTED_FIELD_NUMBER: _ClassVar[int]
    TOKEN_ENDPOINT_AUTH_METHODS_SUPPORTED_FIELD_NUMBER: _ClassVar[int]
    JWKS_URI_FIELD_NUMBER: _ClassVar[int]
    CODE_CHALLENGE_METHODS_SUPPORTED_FIELD_NUMBER: _ClassVar[int]
    GRANT_TYPES_SUPPORTED_FIELD_NUMBER: _ClassVar[int]
    DEVICE_AUTHORIZATION_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    issuer: str
    authorization_endpoint: str
    token_endpoint: str
    response_types_supported: _containers.RepeatedScalarFieldContainer[str]
    scopes_supported: _containers.RepeatedScalarFieldContainer[str]
    token_endpoint_auth_methods_supported: _containers.RepeatedScalarFieldContainer[str]
    jwks_uri: str
    code_challenge_methods_supported: _containers.RepeatedScalarFieldContainer[str]
    grant_types_supported: _containers.RepeatedScalarFieldContainer[str]
    device_authorization_endpoint: str
    def __init__(self, issuer: _Optional[str] = ..., authorization_endpoint: _Optional[str] = ..., token_endpoint: _Optional[str] = ..., response_types_supported: _Optional[_Iterable[str]] = ..., scopes_supported: _Optional[_Iterable[str]] = ..., token_endpoint_auth_methods_supported: _Optional[_Iterable[str]] = ..., jwks_uri: _Optional[str] = ..., code_challenge_methods_supported: _Optional[_Iterable[str]] = ..., grant_types_supported: _Optional[_Iterable[str]] = ..., device_authorization_endpoint: _Optional[str] = ...) -> None: ...

class PublicClientAuthConfigRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class PublicClientAuthConfigResponse(_message.Message):
    __slots__ = ["client_id", "redirect_uri", "scopes", "authorization_metadata_key", "service_http_endpoint", "audience"]
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    REDIRECT_URI_FIELD_NUMBER: _ClassVar[int]
    SCOPES_FIELD_NUMBER: _ClassVar[int]
    AUTHORIZATION_METADATA_KEY_FIELD_NUMBER: _ClassVar[int]
    SERVICE_HTTP_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    AUDIENCE_FIELD_NUMBER: _ClassVar[int]
    client_id: str
    redirect_uri: str
    scopes: _containers.RepeatedScalarFieldContainer[str]
    authorization_metadata_key: str
    service_http_endpoint: str
    audience: str
    def __init__(self, client_id: _Optional[str] = ..., redirect_uri: _Optional[str] = ..., scopes: _Optional[_Iterable[str]] = ..., authorization_metadata_key: _Optional[str] = ..., service_http_endpoint: _Optional[str] = ..., audience: _Optional[str] = ...) -> None: ...
