from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.event import event_pb2 as _event_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TaskExecutionGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionListRequest(_message.Message):
    __slots__ = ["node_execution_id", "limit", "token", "filters", "sort_by"]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    limit: int
    token: str
    filters: str
    sort_by: _common_pb2.Sort
    def __init__(self, node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...

class TaskExecution(_message.Message):
    __slots__ = ["id", "input_uri", "closure", "is_parent"]
    ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    IS_PARENT_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    input_uri: str
    closure: TaskExecutionClosure
    is_parent: bool
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., input_uri: _Optional[str] = ..., closure: _Optional[_Union[TaskExecutionClosure, _Mapping]] = ..., is_parent: bool = ...) -> None: ...

class TaskExecutionList(_message.Message):
    __slots__ = ["task_executions", "token"]
    TASK_EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    task_executions: _containers.RepeatedCompositeFieldContainer[TaskExecution]
    token: str
    def __init__(self, task_executions: _Optional[_Iterable[_Union[TaskExecution, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class TaskExecutionClosure(_message.Message):
    __slots__ = ["output_uri", "error", "output_data", "phase", "logs", "started_at", "duration", "created_at", "updated_at", "custom_info", "reason", "task_type", "metadata", "event_version", "reasons"]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    REASONS_FIELD_NUMBER: _ClassVar[int]
    output_uri: str
    error: _execution_pb2.ExecutionError
    output_data: _literals_pb2.LiteralMap
    phase: _execution_pb2.TaskExecution.Phase
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    started_at: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    custom_info: _struct_pb2.Struct
    reason: str
    task_type: str
    metadata: _event_pb2.TaskExecutionMetadata
    event_version: int
    reasons: _containers.RepeatedCompositeFieldContainer[Reason]
    def __init__(self, output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., reason: _Optional[str] = ..., task_type: _Optional[str] = ..., metadata: _Optional[_Union[_event_pb2.TaskExecutionMetadata, _Mapping]] = ..., event_version: _Optional[int] = ..., reasons: _Optional[_Iterable[_Union[Reason, _Mapping]]] = ...) -> None: ...

class Reason(_message.Message):
    __slots__ = ["occurred_at", "message"]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    occurred_at: _timestamp_pb2.Timestamp
    message: str
    def __init__(self, occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., message: _Optional[str] = ...) -> None: ...

class TaskExecutionGetDataRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionGetDataResponse(_message.Message):
    __slots__ = ["inputs", "outputs", "full_inputs", "full_outputs", "flyte_urls"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_INPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    FLYTE_URLS_FIELD_NUMBER: _ClassVar[int]
    inputs: _common_pb2.UrlBlob
    outputs: _common_pb2.UrlBlob
    full_inputs: _literals_pb2.LiteralMap
    full_outputs: _literals_pb2.LiteralMap
    flyte_urls: _common_pb2.FlyteURLs
    def __init__(self, inputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., full_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., full_outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., flyte_urls: _Optional[_Union[_common_pb2.FlyteURLs, _Mapping]] = ...) -> None: ...
