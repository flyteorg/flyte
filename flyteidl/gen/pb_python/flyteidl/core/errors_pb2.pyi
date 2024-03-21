from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

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

class CacheEvictionError(_message.Message):
    __slots__ = ["code", "message", "node_execution_id", "task_execution_id", "workflow_execution_id"]
    class Code(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        INTERNAL: _ClassVar[CacheEvictionError.Code]
        RESERVATION_NOT_ACQUIRED: _ClassVar[CacheEvictionError.Code]
        DATABASE_UPDATE_FAILED: _ClassVar[CacheEvictionError.Code]
        ARTIFACT_DELETE_FAILED: _ClassVar[CacheEvictionError.Code]
        RESERVATION_NOT_RELEASED: _ClassVar[CacheEvictionError.Code]
    INTERNAL: CacheEvictionError.Code
    RESERVATION_NOT_ACQUIRED: CacheEvictionError.Code
    DATABASE_UPDATE_FAILED: CacheEvictionError.Code
    ARTIFACT_DELETE_FAILED: CacheEvictionError.Code
    RESERVATION_NOT_RELEASED: CacheEvictionError.Code
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    code: CacheEvictionError.Code
    message: str
    node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    task_execution_id: _identifier_pb2.TaskExecutionIdentifier
    workflow_execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, code: _Optional[_Union[CacheEvictionError.Code, str]] = ..., message: _Optional[str] = ..., node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., workflow_execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class CacheEvictionErrorList(_message.Message):
    __slots__ = ["errors"]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    errors: _containers.RepeatedCompositeFieldContainer[CacheEvictionError]
    def __init__(self, errors: _Optional[_Iterable[_Union[CacheEvictionError, _Mapping]]] = ...) -> None: ...
