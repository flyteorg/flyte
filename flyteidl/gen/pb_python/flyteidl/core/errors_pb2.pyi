from flyteidl.core import execution_pb2 as _execution_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ContainerError(_message.Message):
    __slots__ = ["code", "message", "kind", "origin"]
    class Kind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        NON_RECOVERABLE: _ClassVar[ContainerError.Kind]
        RECOVERABLE: _ClassVar[ContainerError.Kind]
    NON_RECOVERABLE: ContainerError.Kind
    RECOVERABLE: ContainerError.Kind
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    ORIGIN_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    kind: ContainerError.Kind
    origin: _execution_pb2.ExecutionError.ErrorKind
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ..., kind: _Optional[_Union[ContainerError.Kind, str]] = ..., origin: _Optional[_Union[_execution_pb2.ExecutionError.ErrorKind, str]] = ...) -> None: ...

class ErrorDocument(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: ContainerError
    def __init__(self, error: _Optional[_Union[ContainerError, _Mapping]] = ...) -> None: ...
