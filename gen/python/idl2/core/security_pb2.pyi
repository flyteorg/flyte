from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Secret(_message.Message):
    __slots__ = ["group", "group_version", "key", "mount_requirement", "env_var"]
    class MountType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        ANY: _ClassVar[Secret.MountType]
        ENV_VAR: _ClassVar[Secret.MountType]
        FILE: _ClassVar[Secret.MountType]
    ANY: Secret.MountType
    ENV_VAR: Secret.MountType
    FILE: Secret.MountType
    GROUP_FIELD_NUMBER: _ClassVar[int]
    GROUP_VERSION_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    MOUNT_REQUIREMENT_FIELD_NUMBER: _ClassVar[int]
    ENV_VAR_FIELD_NUMBER: _ClassVar[int]
    group: str
    group_version: str
    key: str
    mount_requirement: Secret.MountType
    env_var: str
    def __init__(self, group: _Optional[str] = ..., group_version: _Optional[str] = ..., key: _Optional[str] = ..., mount_requirement: _Optional[_Union[Secret.MountType, str]] = ..., env_var: _Optional[str] = ...) -> None: ...

class OAuth2Client(_message.Message):
    __slots__ = ["client_id", "client_secret"]
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    CLIENT_SECRET_FIELD_NUMBER: _ClassVar[int]
    client_id: str
    client_secret: Secret
    def __init__(self, client_id: _Optional[str] = ..., client_secret: _Optional[_Union[Secret, _Mapping]] = ...) -> None: ...

class Identity(_message.Message):
    __slots__ = ["iam_role", "k8s_service_account", "oauth2_client", "execution_identity"]
    IAM_ROLE_FIELD_NUMBER: _ClassVar[int]
    K8S_SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    OAUTH2_CLIENT_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_IDENTITY_FIELD_NUMBER: _ClassVar[int]
    iam_role: str
    k8s_service_account: str
    oauth2_client: OAuth2Client
    execution_identity: str
    def __init__(self, iam_role: _Optional[str] = ..., k8s_service_account: _Optional[str] = ..., oauth2_client: _Optional[_Union[OAuth2Client, _Mapping]] = ..., execution_identity: _Optional[str] = ...) -> None: ...

class OAuth2TokenRequest(_message.Message):
    __slots__ = ["name", "type", "client", "idp_discovery_endpoint", "token_endpoint"]
    class Type(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        CLIENT_CREDENTIALS: _ClassVar[OAuth2TokenRequest.Type]
    CLIENT_CREDENTIALS: OAuth2TokenRequest.Type
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    CLIENT_FIELD_NUMBER: _ClassVar[int]
    IDP_DISCOVERY_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    name: str
    type: OAuth2TokenRequest.Type
    client: OAuth2Client
    idp_discovery_endpoint: str
    token_endpoint: str
    def __init__(self, name: _Optional[str] = ..., type: _Optional[_Union[OAuth2TokenRequest.Type, str]] = ..., client: _Optional[_Union[OAuth2Client, _Mapping]] = ..., idp_discovery_endpoint: _Optional[str] = ..., token_endpoint: _Optional[str] = ...) -> None: ...

class SecurityContext(_message.Message):
    __slots__ = ["run_as", "secrets", "tokens"]
    RUN_AS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    TOKENS_FIELD_NUMBER: _ClassVar[int]
    run_as: Identity
    secrets: _containers.RepeatedCompositeFieldContainer[Secret]
    tokens: _containers.RepeatedCompositeFieldContainer[OAuth2TokenRequest]
    def __init__(self, run_as: _Optional[_Union[Identity, _Mapping]] = ..., secrets: _Optional[_Iterable[_Union[Secret, _Mapping]]] = ..., tokens: _Optional[_Iterable[_Union[OAuth2TokenRequest, _Mapping]]] = ...) -> None: ...
