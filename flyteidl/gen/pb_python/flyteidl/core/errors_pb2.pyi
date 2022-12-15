from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CacheEvictionError(_message.Message):
    __slots__ = ["code", "message", "node_execution_id", "task_execution_id", "workflow_execution_id"]
    class Code(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ARTIFACT_DELETE_FAILED: CacheEvictionError.Code
    CODE_FIELD_NUMBER: _ClassVar[int]
    DATABASE_UPDATE_FAILED: CacheEvictionError.Code
    INTERNAL: CacheEvictionError.Code
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_NOT_ACQUIRED: CacheEvictionError.Code
    RESERVATION_NOT_RELEASED: CacheEvictionError.Code
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

class ContainerError(_message.Message):
    __slots__ = ["code", "kind", "message", "origin"]
    class Kind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    CODE_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    NON_RECOVERABLE: ContainerError.Kind
    ORIGIN_FIELD_NUMBER: _ClassVar[int]
    RECOVERABLE: ContainerError.Kind
    code: str
    kind: ContainerError.Kind
    message: str
    origin: _execution_pb2.ExecutionError.ErrorKind
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ..., kind: _Optional[_Union[ContainerError.Kind, str]] = ..., origin: _Optional[_Union[_execution_pb2.ExecutionError.ErrorKind, str]] = ...) -> None: ...

class ErrorDocument(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: ContainerError
    def __init__(self, error: _Optional[_Union[ContainerError, _Mapping]] = ...) -> None: ...
