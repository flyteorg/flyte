from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ExecutionEnvironmentAssignment(_message.Message):
    __slots__ = ["id", "node_ids", "type", "environment", "environment_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NODE_IDS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: str
    node_ids: _containers.RepeatedScalarFieldContainer[str]
    type: str
    environment: _struct_pb2.Struct
    environment_spec: _struct_pb2.Struct
    def __init__(self, id: _Optional[str] = ..., node_ids: _Optional[_Iterable[str]] = ..., type: _Optional[str] = ..., environment: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., environment_spec: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
