from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SecretType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SECRET_TYPE_GENERIC: _ClassVar[SecretType]
    SECRET_TYPE_IMAGE_PULL_SECRET: _ClassVar[SecretType]

class OverallStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNSPECIFIED: _ClassVar[OverallStatus]
    PARTIALLY_PRESENT: _ClassVar[OverallStatus]
    FULLY_PRESENT: _ClassVar[OverallStatus]
    UNKNOWN_STATUS: _ClassVar[OverallStatus]

class SecretPresenceStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNKNOWN: _ClassVar[SecretPresenceStatus]
    MISSING: _ClassVar[SecretPresenceStatus]
    PRESENT: _ClassVar[SecretPresenceStatus]
SECRET_TYPE_GENERIC: SecretType
SECRET_TYPE_IMAGE_PULL_SECRET: SecretType
UNSPECIFIED: OverallStatus
PARTIALLY_PRESENT: OverallStatus
FULLY_PRESENT: OverallStatus
UNKNOWN_STATUS: OverallStatus
UNKNOWN: SecretPresenceStatus
MISSING: SecretPresenceStatus
PRESENT: SecretPresenceStatus

class SecretSpec(_message.Message):
    __slots__ = ["string_value", "binary_value", "type"]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    BINARY_VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    string_value: str
    binary_value: bytes
    type: SecretType
    def __init__(self, string_value: _Optional[str] = ..., binary_value: _Optional[bytes] = ..., type: _Optional[_Union[SecretType, str]] = ...) -> None: ...

class SecretIdentifier(_message.Message):
    __slots__ = ["name", "organization", "domain", "project"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    name: str
    organization: str
    domain: str
    project: str
    def __init__(self, name: _Optional[str] = ..., organization: _Optional[str] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ...) -> None: ...

class SecretMetadata(_message.Message):
    __slots__ = ["created_time", "secret_status", "type"]
    CREATED_TIME_FIELD_NUMBER: _ClassVar[int]
    SECRET_STATUS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    created_time: _timestamp_pb2.Timestamp
    secret_status: SecretStatus
    type: SecretType
    def __init__(self, created_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., secret_status: _Optional[_Union[SecretStatus, _Mapping]] = ..., type: _Optional[_Union[SecretType, str]] = ...) -> None: ...

class SecretStatus(_message.Message):
    __slots__ = ["overall_status", "cluster_status"]
    OVERALL_STATUS_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_STATUS_FIELD_NUMBER: _ClassVar[int]
    overall_status: OverallStatus
    cluster_status: _containers.RepeatedCompositeFieldContainer[ClusterSecretStatus]
    def __init__(self, overall_status: _Optional[_Union[OverallStatus, str]] = ..., cluster_status: _Optional[_Iterable[_Union[ClusterSecretStatus, _Mapping]]] = ...) -> None: ...

class ClusterSecretStatus(_message.Message):
    __slots__ = ["cluster", "presence_status"]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    PRESENCE_STATUS_FIELD_NUMBER: _ClassVar[int]
    cluster: _identifier_pb2.ClusterIdentifier
    presence_status: SecretPresenceStatus
    def __init__(self, cluster: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ..., presence_status: _Optional[_Union[SecretPresenceStatus, str]] = ...) -> None: ...

class Secret(_message.Message):
    __slots__ = ["id", "secret_metadata"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SECRET_METADATA_FIELD_NUMBER: _ClassVar[int]
    id: SecretIdentifier
    secret_metadata: SecretMetadata
    def __init__(self, id: _Optional[_Union[SecretIdentifier, _Mapping]] = ..., secret_metadata: _Optional[_Union[SecretMetadata, _Mapping]] = ...) -> None: ...
