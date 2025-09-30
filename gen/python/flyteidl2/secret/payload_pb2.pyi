from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.secret import definition_pb2 as _definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateSecretRequest(_message.Message):
    __slots__ = ["id", "secret_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SECRET_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _definition_pb2.SecretIdentifier
    secret_spec: _definition_pb2.SecretSpec
    def __init__(self, id: _Optional[_Union[_definition_pb2.SecretIdentifier, _Mapping]] = ..., secret_spec: _Optional[_Union[_definition_pb2.SecretSpec, _Mapping]] = ...) -> None: ...

class CreateSecretResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UpdateSecretRequest(_message.Message):
    __slots__ = ["id", "secret_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SECRET_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _definition_pb2.SecretIdentifier
    secret_spec: _definition_pb2.SecretSpec
    def __init__(self, id: _Optional[_Union[_definition_pb2.SecretIdentifier, _Mapping]] = ..., secret_spec: _Optional[_Union[_definition_pb2.SecretSpec, _Mapping]] = ...) -> None: ...

class UpdateSecretResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetSecretRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _definition_pb2.SecretIdentifier
    def __init__(self, id: _Optional[_Union[_definition_pb2.SecretIdentifier, _Mapping]] = ...) -> None: ...

class GetSecretResponse(_message.Message):
    __slots__ = ["secret"]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    secret: _definition_pb2.Secret
    def __init__(self, secret: _Optional[_Union[_definition_pb2.Secret, _Mapping]] = ...) -> None: ...

class DeleteSecretRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _definition_pb2.SecretIdentifier
    def __init__(self, id: _Optional[_Union[_definition_pb2.SecretIdentifier, _Mapping]] = ...) -> None: ...

class DeleteSecretResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListSecretsRequest(_message.Message):
    __slots__ = ["organization", "domain", "project", "limit", "token", "per_cluster_tokens"]
    class PerClusterTokensEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    PER_CLUSTER_TOKENS_FIELD_NUMBER: _ClassVar[int]
    organization: str
    domain: str
    project: str
    limit: int
    token: str
    per_cluster_tokens: _containers.ScalarMap[str, str]
    def __init__(self, organization: _Optional[str] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., per_cluster_tokens: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ListSecretsResponse(_message.Message):
    __slots__ = ["secrets", "token", "per_cluster_tokens"]
    class PerClusterTokensEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    PER_CLUSTER_TOKENS_FIELD_NUMBER: _ClassVar[int]
    secrets: _containers.RepeatedCompositeFieldContainer[_definition_pb2.Secret]
    token: str
    per_cluster_tokens: _containers.ScalarMap[str, str]
    def __init__(self, secrets: _Optional[_Iterable[_Union[_definition_pb2.Secret, _Mapping]]] = ..., token: _Optional[str] = ..., per_cluster_tokens: _Optional[_Mapping[str, str]] = ...) -> None: ...
