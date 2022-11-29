from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ExecutionError(_message.Message):
    __slots__ = ["code", "error_uri", "kind", "message"]
    class ErrorKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    CODE_FIELD_NUMBER: _ClassVar[int]
    ERROR_URI_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    SYSTEM: ExecutionError.ErrorKind
    UNKNOWN: ExecutionError.ErrorKind
    USER: ExecutionError.ErrorKind
    code: str
    error_uri: str
    kind: ExecutionError.ErrorKind
    message: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ..., error_uri: _Optional[str] = ..., kind: _Optional[_Union[ExecutionError.ErrorKind, str]] = ...) -> None: ...

class NodeExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ABORTED: NodeExecution.Phase
    DYNAMIC_RUNNING: NodeExecution.Phase
    FAILED: NodeExecution.Phase
    FAILING: NodeExecution.Phase
    QUEUED: NodeExecution.Phase
    RECOVERED: NodeExecution.Phase
    RUNNING: NodeExecution.Phase
    SKIPPED: NodeExecution.Phase
    SUCCEEDED: NodeExecution.Phase
    TIMED_OUT: NodeExecution.Phase
    UNDEFINED: NodeExecution.Phase
    def __init__(self) -> None: ...

class QualityOfService(_message.Message):
    __slots__ = ["spec", "tier"]
    class Tier(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    HIGH: QualityOfService.Tier
    LOW: QualityOfService.Tier
    MEDIUM: QualityOfService.Tier
    SPEC_FIELD_NUMBER: _ClassVar[int]
    TIER_FIELD_NUMBER: _ClassVar[int]
    UNDEFINED: QualityOfService.Tier
    spec: QualityOfServiceSpec
    tier: QualityOfService.Tier
    def __init__(self, tier: _Optional[_Union[QualityOfService.Tier, str]] = ..., spec: _Optional[_Union[QualityOfServiceSpec, _Mapping]] = ...) -> None: ...

class QualityOfServiceSpec(_message.Message):
    __slots__ = ["queueing_budget"]
    QUEUEING_BUDGET_FIELD_NUMBER: _ClassVar[int]
    queueing_budget: _duration_pb2.Duration
    def __init__(self, queueing_budget: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class TaskExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ABORTED: TaskExecution.Phase
    FAILED: TaskExecution.Phase
    INITIALIZING: TaskExecution.Phase
    QUEUED: TaskExecution.Phase
    RUNNING: TaskExecution.Phase
    SUCCEEDED: TaskExecution.Phase
    UNDEFINED: TaskExecution.Phase
    WAITING_FOR_RESOURCES: TaskExecution.Phase
    def __init__(self) -> None: ...

class TaskLog(_message.Message):
    __slots__ = ["message_format", "name", "ttl", "uri"]
    class MessageFormat(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    CSV: TaskLog.MessageFormat
    JSON: TaskLog.MessageFormat
    MESSAGE_FORMAT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TTL_FIELD_NUMBER: _ClassVar[int]
    UNKNOWN: TaskLog.MessageFormat
    URI_FIELD_NUMBER: _ClassVar[int]
    message_format: TaskLog.MessageFormat
    name: str
    ttl: _duration_pb2.Duration
    uri: str
    def __init__(self, uri: _Optional[str] = ..., name: _Optional[str] = ..., message_format: _Optional[_Union[TaskLog.MessageFormat, str]] = ..., ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class WorkflowExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ABORTED: WorkflowExecution.Phase
    ABORTING: WorkflowExecution.Phase
    FAILED: WorkflowExecution.Phase
    FAILING: WorkflowExecution.Phase
    QUEUED: WorkflowExecution.Phase
    RUNNING: WorkflowExecution.Phase
    SUCCEEDED: WorkflowExecution.Phase
    SUCCEEDING: WorkflowExecution.Phase
    TIMED_OUT: WorkflowExecution.Phase
    UNDEFINED: WorkflowExecution.Phase
    def __init__(self) -> None: ...
