from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import authorization_pb2 as _authorization_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RoleType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ROLE_TYPE_NONE: _ClassVar[RoleType]
    ROLE_TYPE_ADMIN: _ClassVar[RoleType]
    ROLE_TYPE_CONTRIBUTOR: _ClassVar[RoleType]
    ROLE_TYPE_VIEWER: _ClassVar[RoleType]
    ROLE_TYPE_CUSTOM: _ClassVar[RoleType]
    ROLE_TYPE_CLUSTER_MANAGER: _ClassVar[RoleType]
    ROLE_TYPE_FLYTE_PROJECT_ADMIN: _ClassVar[RoleType]
    ROLE_TYPE_SERVERLESS_VIEWER: _ClassVar[RoleType]
    ROLE_TYPE_SERVERLESS_CONTRIBUTOR: _ClassVar[RoleType]
    ROLE_TYPE_SUPPORT: _ClassVar[RoleType]
ROLE_TYPE_NONE: RoleType
ROLE_TYPE_ADMIN: RoleType
ROLE_TYPE_CONTRIBUTOR: RoleType
ROLE_TYPE_VIEWER: RoleType
ROLE_TYPE_CUSTOM: RoleType
ROLE_TYPE_CLUSTER_MANAGER: RoleType
ROLE_TYPE_FLYTE_PROJECT_ADMIN: RoleType
ROLE_TYPE_SERVERLESS_VIEWER: RoleType
ROLE_TYPE_SERVERLESS_CONTRIBUTOR: RoleType
ROLE_TYPE_SUPPORT: RoleType

class Role(_message.Message):
    __slots__ = ["id", "permissions", "role_spec", "role_type", "actions"]
    ID_FIELD_NUMBER: _ClassVar[int]
    PERMISSIONS_FIELD_NUMBER: _ClassVar[int]
    ROLE_SPEC_FIELD_NUMBER: _ClassVar[int]
    ROLE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.RoleIdentifier
    permissions: _containers.RepeatedCompositeFieldContainer[_authorization_pb2.Permission]
    role_spec: RoleSpec
    role_type: RoleType
    actions: _containers.RepeatedScalarFieldContainer[_authorization_pb2.Action]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.RoleIdentifier, _Mapping]] = ..., permissions: _Optional[_Iterable[_Union[_authorization_pb2.Permission, _Mapping]]] = ..., role_spec: _Optional[_Union[RoleSpec, _Mapping]] = ..., role_type: _Optional[_Union[RoleType, str]] = ..., actions: _Optional[_Iterable[_Union[_authorization_pb2.Action, str]]] = ...) -> None: ...

class RoleSpec(_message.Message):
    __slots__ = ["description"]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    description: str
    def __init__(self, description: _Optional[str] = ...) -> None: ...
