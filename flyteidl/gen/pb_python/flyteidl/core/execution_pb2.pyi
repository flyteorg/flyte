from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNDEFINED: _ClassVar[WorkflowExecution.Phase]
        QUEUED: _ClassVar[WorkflowExecution.Phase]
        RUNNING: _ClassVar[WorkflowExecution.Phase]
        SUCCEEDING: _ClassVar[WorkflowExecution.Phase]
        SUCCEEDED: _ClassVar[WorkflowExecution.Phase]
        FAILING: _ClassVar[WorkflowExecution.Phase]
        FAILED: _ClassVar[WorkflowExecution.Phase]
        ABORTED: _ClassVar[WorkflowExecution.Phase]
        TIMED_OUT: _ClassVar[WorkflowExecution.Phase]
        ABORTING: _ClassVar[WorkflowExecution.Phase]
    UNDEFINED: WorkflowExecution.Phase
    QUEUED: WorkflowExecution.Phase
    RUNNING: WorkflowExecution.Phase
    SUCCEEDING: WorkflowExecution.Phase
    SUCCEEDED: WorkflowExecution.Phase
    FAILING: WorkflowExecution.Phase
    FAILED: WorkflowExecution.Phase
    ABORTED: WorkflowExecution.Phase
    TIMED_OUT: WorkflowExecution.Phase
    ABORTING: WorkflowExecution.Phase
    def __init__(self) -> None: ...

class NodeExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNDEFINED: _ClassVar[NodeExecution.Phase]
        QUEUED: _ClassVar[NodeExecution.Phase]
        RUNNING: _ClassVar[NodeExecution.Phase]
        SUCCEEDED: _ClassVar[NodeExecution.Phase]
        FAILING: _ClassVar[NodeExecution.Phase]
        FAILED: _ClassVar[NodeExecution.Phase]
        ABORTED: _ClassVar[NodeExecution.Phase]
        SKIPPED: _ClassVar[NodeExecution.Phase]
        TIMED_OUT: _ClassVar[NodeExecution.Phase]
        DYNAMIC_RUNNING: _ClassVar[NodeExecution.Phase]
        RECOVERED: _ClassVar[NodeExecution.Phase]
    UNDEFINED: NodeExecution.Phase
    QUEUED: NodeExecution.Phase
    RUNNING: NodeExecution.Phase
    SUCCEEDED: NodeExecution.Phase
    FAILING: NodeExecution.Phase
    FAILED: NodeExecution.Phase
    ABORTED: NodeExecution.Phase
    SKIPPED: NodeExecution.Phase
    TIMED_OUT: NodeExecution.Phase
    DYNAMIC_RUNNING: NodeExecution.Phase
    RECOVERED: NodeExecution.Phase
    def __init__(self) -> None: ...

class TaskExecution(_message.Message):
    __slots__ = []
    class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNDEFINED: _ClassVar[TaskExecution.Phase]
        QUEUED: _ClassVar[TaskExecution.Phase]
        RUNNING: _ClassVar[TaskExecution.Phase]
        SUCCEEDED: _ClassVar[TaskExecution.Phase]
        ABORTED: _ClassVar[TaskExecution.Phase]
        FAILED: _ClassVar[TaskExecution.Phase]
        INITIALIZING: _ClassVar[TaskExecution.Phase]
        WAITING_FOR_RESOURCES: _ClassVar[TaskExecution.Phase]
        RETRYABLE_FAILED: _ClassVar[TaskExecution.Phase]
    UNDEFINED: TaskExecution.Phase
    QUEUED: TaskExecution.Phase
    RUNNING: TaskExecution.Phase
    SUCCEEDED: TaskExecution.Phase
    ABORTED: TaskExecution.Phase
    FAILED: TaskExecution.Phase
    INITIALIZING: TaskExecution.Phase
    WAITING_FOR_RESOURCES: TaskExecution.Phase
    RETRYABLE_FAILED: TaskExecution.Phase
    def __init__(self) -> None: ...

class ExecutionError(_message.Message):
    __slots__ = ["code", "message", "error_uri", "kind", "timestamp", "worker"]
    class ErrorKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[ExecutionError.ErrorKind]
        USER: _ClassVar[ExecutionError.ErrorKind]
        SYSTEM: _ClassVar[ExecutionError.ErrorKind]
    UNKNOWN: ExecutionError.ErrorKind
    USER: ExecutionError.ErrorKind
    SYSTEM: ExecutionError.ErrorKind
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_URI_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    WORKER_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    error_uri: str
    kind: ExecutionError.ErrorKind
    timestamp: _timestamp_pb2.Timestamp
    worker: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ..., error_uri: _Optional[str] = ..., kind: _Optional[_Union[ExecutionError.ErrorKind, str]] = ..., timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., worker: _Optional[str] = ...) -> None: ...

class TaskLog(_message.Message):
    __slots__ = ["uri", "name", "message_format", "ttl", "ShowWhilePending", "HideOnceFinished"]
    class MessageFormat(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[TaskLog.MessageFormat]
        CSV: _ClassVar[TaskLog.MessageFormat]
        JSON: _ClassVar[TaskLog.MessageFormat]
    UNKNOWN: TaskLog.MessageFormat
    CSV: TaskLog.MessageFormat
    JSON: TaskLog.MessageFormat
    URI_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FORMAT_FIELD_NUMBER: _ClassVar[int]
    TTL_FIELD_NUMBER: _ClassVar[int]
    SHOWWHILEPENDING_FIELD_NUMBER: _ClassVar[int]
    HIDEONCEFINISHED_FIELD_NUMBER: _ClassVar[int]
    uri: str
    name: str
    message_format: TaskLog.MessageFormat
    ttl: _duration_pb2.Duration
    ShowWhilePending: bool
    HideOnceFinished: bool
    def __init__(self, uri: _Optional[str] = ..., name: _Optional[str] = ..., message_format: _Optional[_Union[TaskLog.MessageFormat, str]] = ..., ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., ShowWhilePending: bool = ..., HideOnceFinished: bool = ...) -> None: ...

class LogContext(_message.Message):
    __slots__ = ["pods", "primary_pod_name"]
    PODS_FIELD_NUMBER: _ClassVar[int]
    PRIMARY_POD_NAME_FIELD_NUMBER: _ClassVar[int]
    pods: _containers.RepeatedCompositeFieldContainer[PodLogContext]
    primary_pod_name: str
    def __init__(self, pods: _Optional[_Iterable[_Union[PodLogContext, _Mapping]]] = ..., primary_pod_name: _Optional[str] = ...) -> None: ...

class PodLogContext(_message.Message):
    __slots__ = ["namespace", "pod_name", "containers", "primary_container_name", "init_containers"]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    POD_NAME_FIELD_NUMBER: _ClassVar[int]
    CONTAINERS_FIELD_NUMBER: _ClassVar[int]
    PRIMARY_CONTAINER_NAME_FIELD_NUMBER: _ClassVar[int]
    INIT_CONTAINERS_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    pod_name: str
    containers: _containers.RepeatedCompositeFieldContainer[ContainerContext]
    primary_container_name: str
    init_containers: _containers.RepeatedCompositeFieldContainer[ContainerContext]
    def __init__(self, namespace: _Optional[str] = ..., pod_name: _Optional[str] = ..., containers: _Optional[_Iterable[_Union[ContainerContext, _Mapping]]] = ..., primary_container_name: _Optional[str] = ..., init_containers: _Optional[_Iterable[_Union[ContainerContext, _Mapping]]] = ...) -> None: ...

class ContainerContext(_message.Message):
    __slots__ = ["container_name", "process"]
    class ProcessContext(_message.Message):
        __slots__ = ["container_start_time", "container_end_time"]
        CONTAINER_START_TIME_FIELD_NUMBER: _ClassVar[int]
        CONTAINER_END_TIME_FIELD_NUMBER: _ClassVar[int]
        container_start_time: _timestamp_pb2.Timestamp
        container_end_time: _timestamp_pb2.Timestamp
        def __init__(self, container_start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., container_end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    CONTAINER_NAME_FIELD_NUMBER: _ClassVar[int]
    PROCESS_FIELD_NUMBER: _ClassVar[int]
    container_name: str
    process: ContainerContext.ProcessContext
    def __init__(self, container_name: _Optional[str] = ..., process: _Optional[_Union[ContainerContext.ProcessContext, _Mapping]] = ...) -> None: ...

class QualityOfServiceSpec(_message.Message):
    __slots__ = ["queueing_budget"]
    QUEUEING_BUDGET_FIELD_NUMBER: _ClassVar[int]
    queueing_budget: _duration_pb2.Duration
    def __init__(self, queueing_budget: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class QualityOfService(_message.Message):
    __slots__ = ["tier", "spec"]
    class Tier(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNDEFINED: _ClassVar[QualityOfService.Tier]
        HIGH: _ClassVar[QualityOfService.Tier]
        MEDIUM: _ClassVar[QualityOfService.Tier]
        LOW: _ClassVar[QualityOfService.Tier]
    UNDEFINED: QualityOfService.Tier
    HIGH: QualityOfService.Tier
    MEDIUM: QualityOfService.Tier
    LOW: QualityOfService.Tier
    TIER_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    tier: QualityOfService.Tier
    spec: QualityOfServiceSpec
    def __init__(self, tier: _Optional[_Union[QualityOfService.Tier, str]] = ..., spec: _Optional[_Union[QualityOfServiceSpec, _Mapping]] = ...) -> None: ...
