from common import identifier_pb2 as _identifier_pb2
from common import identity_pb2 as _identity_pb2
from core import catalog_pb2 as _catalog_pb2
from core import execution_pb2 as _execution_pb2
from core import literals_pb2 as _literals_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from validate import validate_pb2 as _validate_pb2
from workflow import task_definition_pb2 as _task_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Phase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    PHASE_UNSPECIFIED: _ClassVar[Phase]
    PHASE_QUEUED: _ClassVar[Phase]
    PHASE_WAITING_FOR_RESOURCES: _ClassVar[Phase]
    PHASE_INITIALIZING: _ClassVar[Phase]
    PHASE_RUNNING: _ClassVar[Phase]
    PHASE_SUCCEEDED: _ClassVar[Phase]
    PHASE_FAILED: _ClassVar[Phase]
    PHASE_ABORTED: _ClassVar[Phase]
    PHASE_TIMED_OUT: _ClassVar[Phase]
PHASE_UNSPECIFIED: Phase
PHASE_QUEUED: Phase
PHASE_WAITING_FOR_RESOURCES: Phase
PHASE_INITIALIZING: Phase
PHASE_RUNNING: Phase
PHASE_SUCCEEDED: Phase
PHASE_FAILED: Phase
PHASE_ABORTED: Phase
PHASE_TIMED_OUT: Phase

class Labels(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Annotations(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Envs(_message.Message):
    __slots__ = ["values"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    def __init__(self, values: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ...) -> None: ...

class RunSpec(_message.Message):
    __slots__ = ["labels", "annotations", "envs", "interruptible", "overwrite_cache", "cluster"]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    labels: Labels
    annotations: Annotations
    envs: Envs
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    cluster: str
    def __init__(self, labels: _Optional[_Union[Labels, _Mapping]] = ..., annotations: _Optional[_Union[Annotations, _Mapping]] = ..., envs: _Optional[_Union[Envs, _Mapping]] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., cluster: _Optional[str] = ...) -> None: ...

class Run(_message.Message):
    __slots__ = ["action"]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    action: Action
    def __init__(self, action: _Optional[_Union[Action, _Mapping]] = ...) -> None: ...

class RunDetails(_message.Message):
    __slots__ = ["run_spec", "action"]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    run_spec: RunSpec
    action: ActionDetails
    def __init__(self, run_spec: _Optional[_Union[RunSpec, _Mapping]] = ..., action: _Optional[_Union[ActionDetails, _Mapping]] = ...) -> None: ...

class TaskActionMetadata(_message.Message):
    __slots__ = ["id", "task_type", "short_name"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    SHORT_NAME_FIELD_NUMBER: _ClassVar[int]
    id: _task_definition_pb2.TaskIdentifier
    task_type: str
    short_name: str
    def __init__(self, id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_type: _Optional[str] = ..., short_name: _Optional[str] = ...) -> None: ...

class TraceActionMetadata(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class ConditionActionMetadata(_message.Message):
    __slots__ = ["name", "run_id", "action_id"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    GLOBAL_FIELD_NUMBER: _ClassVar[int]
    name: str
    run_id: str
    action_id: str
    def __init__(self, name: _Optional[str] = ..., run_id: _Optional[str] = ..., action_id: _Optional[str] = ..., **kwargs) -> None: ...

class ActionMetadata(_message.Message):
    __slots__ = ["parent", "group", "executed_by", "task", "trace", "condition"]
    PARENT_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    EXECUTED_BY_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    parent: str
    group: str
    executed_by: _identity_pb2.EnrichedIdentity
    task: TaskActionMetadata
    trace: TraceActionMetadata
    condition: ConditionActionMetadata
    def __init__(self, parent: _Optional[str] = ..., group: _Optional[str] = ..., executed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., task: _Optional[_Union[TaskActionMetadata, _Mapping]] = ..., trace: _Optional[_Union[TraceActionMetadata, _Mapping]] = ..., condition: _Optional[_Union[ConditionActionMetadata, _Mapping]] = ...) -> None: ...

class ActionStatus(_message.Message):
    __slots__ = ["phase", "start_time", "end_time", "attempts", "cache_status"]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    phase: Phase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    attempts: int
    cache_status: _catalog_pb2.CatalogCacheStatus
    def __init__(self, phase: _Optional[_Union[Phase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., attempts: _Optional[int] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ...) -> None: ...

class Action(_message.Message):
    __slots__ = ["id", "metadata", "status"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    metadata: ActionMetadata
    status: ActionStatus
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[ActionMetadata, _Mapping]] = ..., status: _Optional[_Union[ActionStatus, _Mapping]] = ...) -> None: ...

class EnrichedAction(_message.Message):
    __slots__ = ["action", "meets_filter", "children_phase_counts"]
    class ChildrenPhaseCountsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: int
        value: int
        def __init__(self, key: _Optional[int] = ..., value: _Optional[int] = ...) -> None: ...
    ACTION_FIELD_NUMBER: _ClassVar[int]
    MEETS_FILTER_FIELD_NUMBER: _ClassVar[int]
    CHILDREN_PHASE_COUNTS_FIELD_NUMBER: _ClassVar[int]
    action: Action
    meets_filter: bool
    children_phase_counts: _containers.ScalarMap[int, int]
    def __init__(self, action: _Optional[_Union[Action, _Mapping]] = ..., meets_filter: bool = ..., children_phase_counts: _Optional[_Mapping[int, int]] = ...) -> None: ...

class ErrorInfo(_message.Message):
    __slots__ = ["message", "kind"]
    class Kind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        KIND_UNSPECIFIED: _ClassVar[ErrorInfo.Kind]
        KIND_USER: _ClassVar[ErrorInfo.Kind]
        KIND_SYSTEM: _ClassVar[ErrorInfo.Kind]
    KIND_UNSPECIFIED: ErrorInfo.Kind
    KIND_USER: ErrorInfo.Kind
    KIND_SYSTEM: ErrorInfo.Kind
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    message: str
    kind: ErrorInfo.Kind
    def __init__(self, message: _Optional[str] = ..., kind: _Optional[_Union[ErrorInfo.Kind, str]] = ...) -> None: ...

class AbortInfo(_message.Message):
    __slots__ = ["reason", "aborted_by"]
    REASON_FIELD_NUMBER: _ClassVar[int]
    ABORTED_BY_FIELD_NUMBER: _ClassVar[int]
    reason: str
    aborted_by: _identity_pb2.EnrichedIdentity
    def __init__(self, reason: _Optional[str] = ..., aborted_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...

class ActionDetails(_message.Message):
    __slots__ = ["id", "metadata", "status", "error_info", "abort_info", "resolved_task_spec", "attempts"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_INFO_FIELD_NUMBER: _ClassVar[int]
    ABORT_INFO_FIELD_NUMBER: _ClassVar[int]
    RESOLVED_TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    metadata: ActionMetadata
    status: ActionStatus
    error_info: ErrorInfo
    abort_info: AbortInfo
    resolved_task_spec: _task_definition_pb2.TaskSpec
    attempts: _containers.RepeatedCompositeFieldContainer[ActionAttempt]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[ActionMetadata, _Mapping]] = ..., status: _Optional[_Union[ActionStatus, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., abort_info: _Optional[_Union[AbortInfo, _Mapping]] = ..., resolved_task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., attempts: _Optional[_Iterable[_Union[ActionAttempt, _Mapping]]] = ...) -> None: ...

class ActionAttempt(_message.Message):
    __slots__ = ["phase", "start_time", "end_time", "error_info", "attempt", "log_info", "outputs", "logs_available", "cache_status", "cluster_events", "phase_transitions", "cluster", "log_context"]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    ERROR_INFO_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    LOG_INFO_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    LOGS_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_EVENTS_FIELD_NUMBER: _ClassVar[int]
    PHASE_TRANSITIONS_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    LOG_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    phase: Phase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    error_info: ErrorInfo
    attempt: int
    log_info: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    outputs: OutputReferences
    logs_available: bool
    cache_status: _catalog_pb2.CatalogCacheStatus
    cluster_events: _containers.RepeatedCompositeFieldContainer[ClusterEvent]
    phase_transitions: _containers.RepeatedCompositeFieldContainer[PhaseTransition]
    cluster: str
    log_context: _execution_pb2.LogContext
    def __init__(self, phase: _Optional[_Union[Phase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., attempt: _Optional[int] = ..., log_info: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., outputs: _Optional[_Union[OutputReferences, _Mapping]] = ..., logs_available: bool = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., cluster_events: _Optional[_Iterable[_Union[ClusterEvent, _Mapping]]] = ..., phase_transitions: _Optional[_Iterable[_Union[PhaseTransition, _Mapping]]] = ..., cluster: _Optional[str] = ..., log_context: _Optional[_Union[_execution_pb2.LogContext, _Mapping]] = ...) -> None: ...

class ClusterEvent(_message.Message):
    __slots__ = ["occurred_at", "message"]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    occurred_at: _timestamp_pb2.Timestamp
    message: str
    def __init__(self, occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., message: _Optional[str] = ...) -> None: ...

class PhaseTransition(_message.Message):
    __slots__ = ["phase", "start_time", "end_time"]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    phase: Phase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    def __init__(self, phase: _Optional[_Union[Phase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ActionEvent(_message.Message):
    __slots__ = ["id", "attempt", "phase", "version", "start_time", "updated_time", "end_time", "error_info", "log_info", "log_context", "cluster", "outputs", "cache_status", "cluster_events", "reported_time"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    UPDATED_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    ERROR_INFO_FIELD_NUMBER: _ClassVar[int]
    LOG_INFO_FIELD_NUMBER: _ClassVar[int]
    LOG_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_EVENTS_FIELD_NUMBER: _ClassVar[int]
    REPORTED_TIME_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    attempt: int
    phase: Phase
    version: int
    start_time: _timestamp_pb2.Timestamp
    updated_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    error_info: ErrorInfo
    log_info: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    log_context: _execution_pb2.LogContext
    cluster: str
    outputs: OutputReferences
    cache_status: _catalog_pb2.CatalogCacheStatus
    cluster_events: _containers.RepeatedCompositeFieldContainer[ClusterEvent]
    reported_time: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., phase: _Optional[_Union[Phase, str]] = ..., version: _Optional[int] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., log_info: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., log_context: _Optional[_Union[_execution_pb2.LogContext, _Mapping]] = ..., cluster: _Optional[str] = ..., outputs: _Optional[_Union[OutputReferences, _Mapping]] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., cluster_events: _Optional[_Iterable[_Union[ClusterEvent, _Mapping]]] = ..., reported_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class NamedLiteral(_message.Message):
    __slots__ = ["name", "value"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    value: _literals_pb2.Literal
    def __init__(self, name: _Optional[str] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...

class OutputReferences(_message.Message):
    __slots__ = ["output_uri", "report_uri"]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    REPORT_URI_FIELD_NUMBER: _ClassVar[int]
    output_uri: str
    report_uri: str
    def __init__(self, output_uri: _Optional[str] = ..., report_uri: _Optional[str] = ...) -> None: ...

class Inputs(_message.Message):
    __slots__ = ["literals", "context"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[NamedLiteral]
    context: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    def __init__(self, literals: _Optional[_Iterable[_Union[NamedLiteral, _Mapping]]] = ..., context: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ...) -> None: ...

class Outputs(_message.Message):
    __slots__ = ["literals"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[NamedLiteral]
    def __init__(self, literals: _Optional[_Iterable[_Union[NamedLiteral, _Mapping]]] = ...) -> None: ...
