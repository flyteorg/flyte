from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import run_pb2 as _run_pb2
from flyteidl2.logs.dataplane import payload_pb2 as _payload_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2_1
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from flyteidl2.workflow import run_service_pb2 as _run_service_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.rpc import status_pb2 as _status_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateTrackedRunRequest(_message.Message):
    __slots__ = ["run_id", "project_id", "task_id", "task_spec", "offloaded_input_data", "run_spec", "labels", "run_start_time"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    OFFLOADED_INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    RUN_START_TIME_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    task_spec: _task_definition_pb2.TaskSpec
    offloaded_input_data: _run_pb2.OffloadedInputData
    run_spec: _run_pb2_1.RunSpec
    labels: _containers.ScalarMap[str, str]
    run_start_time: _timestamp_pb2.Timestamp
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., offloaded_input_data: _Optional[_Union[_run_pb2.OffloadedInputData, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2_1.RunSpec, _Mapping]] = ..., labels: _Optional[_Mapping[str, str]] = ..., run_start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class TrackedActionUpdate(_message.Message):
    __slots__ = ["event", "parent_name", "group", "task", "trace", "status", "log_tail"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    PARENT_NAME_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    LOG_TAIL_FIELD_NUMBER: _ClassVar[int]
    event: _run_definition_pb2.ActionEvent
    parent_name: str
    group: str
    task: _run_definition_pb2.TaskAction
    trace: _run_definition_pb2.TraceAction
    status: _run_definition_pb2.ActionStatus
    log_tail: LogTail
    def __init__(self, event: _Optional[_Union[_run_definition_pb2.ActionEvent, _Mapping]] = ..., parent_name: _Optional[str] = ..., group: _Optional[str] = ..., task: _Optional[_Union[_run_definition_pb2.TaskAction, _Mapping]] = ..., trace: _Optional[_Union[_run_definition_pb2.TraceAction, _Mapping]] = ..., status: _Optional[_Union[_run_definition_pb2.ActionStatus, _Mapping]] = ..., log_tail: _Optional[_Union[LogTail, _Mapping]] = ...) -> None: ...

class LogTail(_message.Message):
    __slots__ = ["lines", "truncated"]
    LINES_FIELD_NUMBER: _ClassVar[int]
    TRUNCATED_FIELD_NUMBER: _ClassVar[int]
    lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
    truncated: bool
    def __init__(self, lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ..., truncated: bool = ...) -> None: ...

class StreamLogsRequest(_message.Message):
    __slots__ = ["register", "batch", "error"]
    class Register(_message.Message):
        __slots__ = ["run_id"]
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        run_id: _identifier_pb2.RunIdentifier
        def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...
    class LogBatch(_message.Message):
        __slots__ = ["request_id", "lines", "eof"]
        REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
        LINES_FIELD_NUMBER: _ClassVar[int]
        EOF_FIELD_NUMBER: _ClassVar[int]
        request_id: str
        lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
        eof: bool
        def __init__(self, request_id: _Optional[str] = ..., lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ..., eof: bool = ...) -> None: ...
    class LogError(_message.Message):
        __slots__ = ["request_id", "error"]
        REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        request_id: str
        error: _status_pb2.Status
        def __init__(self, request_id: _Optional[str] = ..., error: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...
    REGISTER_FIELD_NUMBER: _ClassVar[int]
    BATCH_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    register: StreamLogsRequest.Register
    batch: StreamLogsRequest.LogBatch
    error: StreamLogsRequest.LogError
    def __init__(self, register: _Optional[_Union[StreamLogsRequest.Register, _Mapping]] = ..., batch: _Optional[_Union[StreamLogsRequest.LogBatch, _Mapping]] = ..., error: _Optional[_Union[StreamLogsRequest.LogError, _Mapping]] = ...) -> None: ...

class StreamLogsResponse(_message.Message):
    __slots__ = ["registered", "serve", "cancel"]
    class Registered(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...
    class ServeLogs(_message.Message):
        __slots__ = ["request_id", "action_attempt_id", "from_timestamp", "follow"]
        REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
        ACTION_ATTEMPT_ID_FIELD_NUMBER: _ClassVar[int]
        FROM_TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
        FOLLOW_FIELD_NUMBER: _ClassVar[int]
        request_id: str
        action_attempt_id: _identifier_pb2.ActionAttemptIdentifier
        from_timestamp: _timestamp_pb2.Timestamp
        follow: bool
        def __init__(self, request_id: _Optional[str] = ..., action_attempt_id: _Optional[_Union[_identifier_pb2.ActionAttemptIdentifier, _Mapping]] = ..., from_timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., follow: bool = ...) -> None: ...
    class CancelLogs(_message.Message):
        __slots__ = ["request_id"]
        REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
        request_id: str
        def __init__(self, request_id: _Optional[str] = ...) -> None: ...
    REGISTERED_FIELD_NUMBER: _ClassVar[int]
    SERVE_FIELD_NUMBER: _ClassVar[int]
    CANCEL_FIELD_NUMBER: _ClassVar[int]
    registered: StreamLogsResponse.Registered
    serve: StreamLogsResponse.ServeLogs
    cancel: StreamLogsResponse.CancelLogs
    def __init__(self, registered: _Optional[_Union[StreamLogsResponse.Registered, _Mapping]] = ..., serve: _Optional[_Union[StreamLogsResponse.ServeLogs, _Mapping]] = ..., cancel: _Optional[_Union[StreamLogsResponse.CancelLogs, _Mapping]] = ...) -> None: ...

class TailTrackedLogsRequest(_message.Message):
    __slots__ = ["action_attempt_id", "follow"]
    ACTION_ATTEMPT_ID_FIELD_NUMBER: _ClassVar[int]
    FOLLOW_FIELD_NUMBER: _ClassVar[int]
    action_attempt_id: _identifier_pb2.ActionAttemptIdentifier
    follow: bool
    def __init__(self, action_attempt_id: _Optional[_Union[_identifier_pb2.ActionAttemptIdentifier, _Mapping]] = ..., follow: bool = ...) -> None: ...

class TailTrackedLogsResponse(_message.Message):
    __slots__ = ["lines", "source", "truncated"]
    class Source(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        SOURCE_UNSPECIFIED: _ClassVar[TailTrackedLogsResponse.Source]
        SOURCE_LIVE: _ClassVar[TailTrackedLogsResponse.Source]
        SOURCE_PERSISTED: _ClassVar[TailTrackedLogsResponse.Source]
    SOURCE_UNSPECIFIED: TailTrackedLogsResponse.Source
    SOURCE_LIVE: TailTrackedLogsResponse.Source
    SOURCE_PERSISTED: TailTrackedLogsResponse.Source
    LINES_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    TRUNCATED_FIELD_NUMBER: _ClassVar[int]
    lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
    source: TailTrackedLogsResponse.Source
    truncated: bool
    def __init__(self, lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ..., source: _Optional[_Union[TailTrackedLogsResponse.Source, str]] = ..., truncated: bool = ...) -> None: ...

class ReportTrackedActionsRequest(_message.Message):
    __slots__ = ["run_id", "updates"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATES_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    updates: _containers.RepeatedCompositeFieldContainer[TrackedActionUpdate]
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., updates: _Optional[_Iterable[_Union[TrackedActionUpdate, _Mapping]]] = ...) -> None: ...

class ReportTrackedActionsResponse(_message.Message):
    __slots__ = ["statuses"]
    STATUSES_FIELD_NUMBER: _ClassVar[int]
    statuses: _containers.RepeatedCompositeFieldContainer[_status_pb2.Status]
    def __init__(self, statuses: _Optional[_Iterable[_Union[_status_pb2.Status, _Mapping]]] = ...) -> None: ...
