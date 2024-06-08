from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ExecutionEnvAssignment(_message.Message):
    __slots__ = ["node_ids", "task_type", "execution_env"]
    NODE_IDS_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ENV_FIELD_NUMBER: _ClassVar[int]
    node_ids: _containers.RepeatedScalarFieldContainer[str]
    task_type: str
    execution_env: ExecutionEnv
    def __init__(self, node_ids: _Optional[_Iterable[str]] = ..., task_type: _Optional[str] = ..., execution_env: _Optional[_Union[ExecutionEnv, _Mapping]] = ...) -> None: ...

class ExecutionEnv(_message.Message):
    __slots__ = ["id", "type", "extant", "spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    EXTANT_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    id: str
    type: str
    extant: _struct_pb2.Struct
    spec: _struct_pb2.Struct
    def __init__(self, id: _Optional[str] = ..., type: _Optional[str] = ..., extant: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., spec: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
