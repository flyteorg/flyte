from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import authorization_pb2 as _authorization_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Policy(_message.Message):
    __slots__ = ["id", "bindings", "description"]
    ID_FIELD_NUMBER: _ClassVar[int]
    BINDINGS_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.PolicyIdentifier
    bindings: _containers.RepeatedCompositeFieldContainer[PolicyBinding]
    description: str
    def __init__(self, id: _Optional[_Union[_identifier_pb2.PolicyIdentifier, _Mapping]] = ..., bindings: _Optional[_Iterable[_Union[PolicyBinding, _Mapping]]] = ..., description: _Optional[str] = ...) -> None: ...

class PolicyBinding(_message.Message):
    __slots__ = ["role_id", "resource"]
    ROLE_ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    role_id: _identifier_pb2.RoleIdentifier
    resource: _authorization_pb2.Resource
    def __init__(self, role_id: _Optional[_Union[_identifier_pb2.RoleIdentifier, _Mapping]] = ..., resource: _Optional[_Union[_authorization_pb2.Resource, _Mapping]] = ...) -> None: ...
