from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RelationType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    RELATION_TYPE_UNSPECIFIED: _ClassVar[RelationType]
    RELATION_TYPE_RERUN: _ClassVar[RelationType]
    RELATION_TYPE_RECOVER: _ClassVar[RelationType]
    RELATION_TYPE_SPAWN: _ClassVar[RelationType]
RELATION_TYPE_UNSPECIFIED: RelationType
RELATION_TYPE_RERUN: RelationType
RELATION_TYPE_RECOVER: RelationType
RELATION_TYPE_SPAWN: RelationType

class OffloadedInputData(_message.Message):
    __slots__ = ["uri", "inputs_hash"]
    URI_FIELD_NUMBER: _ClassVar[int]
    INPUTS_HASH_FIELD_NUMBER: _ClassVar[int]
    uri: str
    inputs_hash: str
    def __init__(self, uri: _Optional[str] = ..., inputs_hash: _Optional[str] = ...) -> None: ...

class Relation(_message.Message):
    __slots__ = ["related_to", "relation_type"]
    RELATED_TO_FIELD_NUMBER: _ClassVar[int]
    RELATION_TYPE_FIELD_NUMBER: _ClassVar[int]
    related_to: _identifier_pb2.RunIdentifier
    relation_type: RelationType
    def __init__(self, related_to: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., relation_type: _Optional[_Union[RelationType, str]] = ...) -> None: ...
