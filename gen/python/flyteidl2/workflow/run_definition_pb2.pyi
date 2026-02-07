from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.common import phase_pb2 as _phase_pb2
from flyteidl2.core import catalog_pb2 as _catalog_pb2
from flyteidl2.core import execution_pb2 as _execution_pb2
from flyteidl2.core import types_pb2 as _types_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ActionType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ACTION_TYPE_UNSPECIFIED: _ClassVar[ActionType]
    ACTION_TYPE_TASK: _ClassVar[ActionType]
    ACTION_TYPE_TRACE: _ClassVar[ActionType]
    ACTION_TYPE_CONDITION: _ClassVar[ActionType]

class RunSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    RUN_SOURCE_UNSPECIFIED: _ClassVar[RunSource]
    RUN_SOURCE_WEB: _ClassVar[RunSource]
    RUN_SOURCE_CLI: _ClassVar[RunSource]
    RUN_SOURCE_SCHEDULE_TRIGGER: _ClassVar[RunSource]
ACTION_TYPE_UNSPECIFIED: ActionType
ACTION_TYPE_TASK: ActionType
ACTION_TYPE_TRACE: ActionType
ACTION_TYPE_CONDITION: ActionType
RUN_SOURCE_UNSPECIFIED: RunSource
RUN_SOURCE_WEB: RunSource
RUN_SOURCE_CLI: RunSource
RUN_SOURCE_SCHEDULE_TRIGGER: RunSource

class Run(_message.Message):
    __slots__ = ["action"]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    action: Action
    def __init__(self, action: _Optional[_Union[Action, _Mapping]] = ...) -> None: ...

class RunDetails(_message.Message):
    __slots__ = ["run_spec", "action"]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    run_spec: _run_pb2.RunSpec
    action: ActionDetails
    def __init__(self, run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ..., action: _Optional[_Union[ActionDetails, _Mapping]] = ...) -> None: ...

class TaskAction(_message.Message):
    __slots__ = ["id", "spec", "cache_key", "cluster"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    CACHE_KEY_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    id: _task_definition_pb2.TaskIdentifier
    spec: _task_definition_pb2.TaskSpec
    cache_key: _wrappers_pb2.StringValue
    cluster: str
    def __init__(self, id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., cache_key: _Optional[_Union[_wrappers_pb2.StringValue, _Mapping]] = ..., cluster: _Optional[str] = ...) -> None: ...

class TraceAction(_message.Message):
    __slots__ = ["name", "phase", "start_time", "end_time", "outputs", "spec"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    name: str
    phase: _phase_pb2.ActionPhase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    outputs: _common_pb2.OutputReferences
    spec: _task_definition_pb2.TraceSpec
    def __init__(self, name: _Optional[str] = ..., phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.OutputReferences, _Mapping]] = ..., spec: _Optional[_Union[_task_definition_pb2.TraceSpec, _Mapping]] = ...) -> None: ...

class ConditionAction(_message.Message):
    __slots__ = ["name", "run_id", "action_id", "type", "prompt", "description"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    GLOBAL_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    PROMPT_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    name: str
    run_id: str
    action_id: str
    type: _types_pb2.LiteralType
    prompt: str
    description: str
    def __init__(self, name: _Optional[str] = ..., run_id: _Optional[str] = ..., action_id: _Optional[str] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., prompt: _Optional[str] = ..., description: _Optional[str] = ..., **kwargs) -> None: ...

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
    __slots__ = ["parent", "group", "executed_by", "task", "trace", "condition", "action_type", "trigger_id", "environment_name", "funtion_name", "trigger_name", "trigger_type", "source"]
    PARENT_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    EXECUTED_BY_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    ACTION_TYPE_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_NAME_FIELD_NUMBER: _ClassVar[int]
    FUNTION_NAME_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_TYPE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    parent: str
    group: str
    executed_by: _identity_pb2.EnrichedIdentity
    task: TaskActionMetadata
    trace: TraceActionMetadata
    condition: ConditionActionMetadata
    action_type: ActionType
    trigger_id: _identifier_pb2.TriggerIdentifier
    environment_name: str
    funtion_name: str
    trigger_name: str
    trigger_type: _common_pb2.TriggerAutomationSpec
    source: RunSource
    def __init__(self, parent: _Optional[str] = ..., group: _Optional[str] = ..., executed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., task: _Optional[_Union[TaskActionMetadata, _Mapping]] = ..., trace: _Optional[_Union[TraceActionMetadata, _Mapping]] = ..., condition: _Optional[_Union[ConditionActionMetadata, _Mapping]] = ..., action_type: _Optional[_Union[ActionType, str]] = ..., trigger_id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ..., environment_name: _Optional[str] = ..., funtion_name: _Optional[str] = ..., trigger_name: _Optional[str] = ..., trigger_type: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ..., source: _Optional[_Union[RunSource, str]] = ...) -> None: ...

class ActionStatus(_message.Message):
    __slots__ = ["phase", "start_time", "end_time", "attempts", "cache_status", "duration_ms"]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    DURATION_MS_FIELD_NUMBER: _ClassVar[int]
    phase: _phase_pb2.ActionPhase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    attempts: int
    cache_status: _catalog_pb2.CatalogCacheStatus
    duration_ms: int
    def __init__(self, phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., attempts: _Optional[int] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., duration_ms: _Optional[int] = ...) -> None: ...

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
    __slots__ = ["id", "metadata", "status", "error_info", "abort_info", "task", "trace", "attempts"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ERROR_INFO_FIELD_NUMBER: _ClassVar[int]
    ABORT_INFO_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    metadata: ActionMetadata
    status: ActionStatus
    error_info: ErrorInfo
    abort_info: AbortInfo
    task: _task_definition_pb2.TaskSpec
    trace: _task_definition_pb2.TraceSpec
    attempts: _containers.RepeatedCompositeFieldContainer[ActionAttempt]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[ActionMetadata, _Mapping]] = ..., status: _Optional[_Union[ActionStatus, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., abort_info: _Optional[_Union[AbortInfo, _Mapping]] = ..., task: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., trace: _Optional[_Union[_task_definition_pb2.TraceSpec, _Mapping]] = ..., attempts: _Optional[_Iterable[_Union[ActionAttempt, _Mapping]]] = ...) -> None: ...

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
    phase: _phase_pb2.ActionPhase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    error_info: ErrorInfo
    attempt: int
    log_info: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    outputs: _common_pb2.OutputReferences
    logs_available: bool
    cache_status: _catalog_pb2.CatalogCacheStatus
    cluster_events: _containers.RepeatedCompositeFieldContainer[ClusterEvent]
    phase_transitions: _containers.RepeatedCompositeFieldContainer[PhaseTransition]
    cluster: str
    log_context: _execution_pb2.LogContext
    def __init__(self, phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., attempt: _Optional[int] = ..., log_info: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., outputs: _Optional[_Union[_common_pb2.OutputReferences, _Mapping]] = ..., logs_available: bool = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., cluster_events: _Optional[_Iterable[_Union[ClusterEvent, _Mapping]]] = ..., phase_transitions: _Optional[_Iterable[_Union[PhaseTransition, _Mapping]]] = ..., cluster: _Optional[str] = ..., log_context: _Optional[_Union[_execution_pb2.LogContext, _Mapping]] = ...) -> None: ...

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
    phase: _phase_pb2.ActionPhase
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    def __init__(self, phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

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
    phase: _phase_pb2.ActionPhase
    version: int
    start_time: _timestamp_pb2.Timestamp
    updated_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    error_info: ErrorInfo
    log_info: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    log_context: _execution_pb2.LogContext
    cluster: str
    outputs: _common_pb2.OutputReferences
    cache_status: _catalog_pb2.CatalogCacheStatus
    cluster_events: _containers.RepeatedCompositeFieldContainer[ClusterEvent]
    reported_time: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., version: _Optional[int] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error_info: _Optional[_Union[ErrorInfo, _Mapping]] = ..., log_info: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., log_context: _Optional[_Union[_execution_pb2.LogContext, _Mapping]] = ..., cluster: _Optional[str] = ..., outputs: _Optional[_Union[_common_pb2.OutputReferences, _Mapping]] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., cluster_events: _Optional[_Iterable[_Union[ClusterEvent, _Mapping]]] = ..., reported_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ActionSpec(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "run_spec", "input_uri", "run_output_base", "task", "condition", "trace", "group"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    RUN_OUTPUT_BASE_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    run_spec: _run_pb2.RunSpec
    input_uri: str
    run_output_base: str
    task: TaskAction
    condition: ConditionAction
    trace: TraceAction
    group: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ..., input_uri: _Optional[str] = ..., run_output_base: _Optional[str] = ..., task: _Optional[_Union[TaskAction, _Mapping]] = ..., condition: _Optional[_Union[ConditionAction, _Mapping]] = ..., trace: _Optional[_Union[TraceAction, _Mapping]] = ..., group: _Optional[str] = ...) -> None: ...

class TaskGroup(_message.Message):
    __slots__ = ["task_name", "environment_name", "total_runs", "latest_run_time", "recent_statuses", "average_failure_rate", "average_duration", "latest_finished_time", "created_by", "should_delete", "short_name", "error_counts", "phase_counts", "average_time_to_running"]
    class RecentStatus(_message.Message):
        __slots__ = ["run_name", "phase"]
        RUN_NAME_FIELD_NUMBER: _ClassVar[int]
        PHASE_FIELD_NUMBER: _ClassVar[int]
        run_name: str
        phase: _phase_pb2.ActionPhase
        def __init__(self, run_name: _Optional[str] = ..., phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ...) -> None: ...
    class ErrorCounts(_message.Message):
        __slots__ = ["user_error", "system_error", "unspecified_error"]
        USER_ERROR_FIELD_NUMBER: _ClassVar[int]
        SYSTEM_ERROR_FIELD_NUMBER: _ClassVar[int]
        UNSPECIFIED_ERROR_FIELD_NUMBER: _ClassVar[int]
        user_error: int
        system_error: int
        unspecified_error: int
        def __init__(self, user_error: _Optional[int] = ..., system_error: _Optional[int] = ..., unspecified_error: _Optional[int] = ...) -> None: ...
    class PhaseCounts(_message.Message):
        __slots__ = ["phase", "count"]
        PHASE_FIELD_NUMBER: _ClassVar[int]
        COUNT_FIELD_NUMBER: _ClassVar[int]
        phase: _phase_pb2.ActionPhase
        count: int
        def __init__(self, phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., count: _Optional[int] = ...) -> None: ...
    TASK_NAME_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_NAME_FIELD_NUMBER: _ClassVar[int]
    TOTAL_RUNS_FIELD_NUMBER: _ClassVar[int]
    LATEST_RUN_TIME_FIELD_NUMBER: _ClassVar[int]
    RECENT_STATUSES_FIELD_NUMBER: _ClassVar[int]
    AVERAGE_FAILURE_RATE_FIELD_NUMBER: _ClassVar[int]
    AVERAGE_DURATION_FIELD_NUMBER: _ClassVar[int]
    LATEST_FINISHED_TIME_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_FIELD_NUMBER: _ClassVar[int]
    SHOULD_DELETE_FIELD_NUMBER: _ClassVar[int]
    SHORT_NAME_FIELD_NUMBER: _ClassVar[int]
    ERROR_COUNTS_FIELD_NUMBER: _ClassVar[int]
    PHASE_COUNTS_FIELD_NUMBER: _ClassVar[int]
    AVERAGE_TIME_TO_RUNNING_FIELD_NUMBER: _ClassVar[int]
    task_name: str
    environment_name: str
    total_runs: int
    latest_run_time: _timestamp_pb2.Timestamp
    recent_statuses: _containers.RepeatedCompositeFieldContainer[TaskGroup.RecentStatus]
    average_failure_rate: float
    average_duration: _duration_pb2.Duration
    latest_finished_time: _timestamp_pb2.Timestamp
    created_by: _containers.RepeatedCompositeFieldContainer[_identity_pb2.EnrichedIdentity]
    should_delete: bool
    short_name: str
    error_counts: TaskGroup.ErrorCounts
    phase_counts: _containers.RepeatedCompositeFieldContainer[TaskGroup.PhaseCounts]
    average_time_to_running: _duration_pb2.Duration
    def __init__(self, task_name: _Optional[str] = ..., environment_name: _Optional[str] = ..., total_runs: _Optional[int] = ..., latest_run_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., recent_statuses: _Optional[_Iterable[_Union[TaskGroup.RecentStatus, _Mapping]]] = ..., average_failure_rate: _Optional[float] = ..., average_duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., latest_finished_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., created_by: _Optional[_Iterable[_Union[_identity_pb2.EnrichedIdentity, _Mapping]]] = ..., should_delete: bool = ..., short_name: _Optional[str] = ..., error_counts: _Optional[_Union[TaskGroup.ErrorCounts, _Mapping]] = ..., phase_counts: _Optional[_Iterable[_Union[TaskGroup.PhaseCounts, _Mapping]]] = ..., average_time_to_running: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...
