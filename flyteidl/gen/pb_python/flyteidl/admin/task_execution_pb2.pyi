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

class TaskExecution(_message.Message):
    __slots__ = ["closure", "id", "input_uri", "is_parent"]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    IS_PARENT_FIELD_NUMBER: _ClassVar[int]
    closure: TaskExecutionClosure
    id: _identifier_pb2.TaskExecutionIdentifier
    input_uri: str
    is_parent: bool
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., input_uri: _Optional[str] = ..., closure: _Optional[_Union[TaskExecutionClosure, _Mapping]] = ..., is_parent: bool = ...) -> None: ...

class TaskExecutionClosure(_message.Message):
    __slots__ = ["created_at", "custom_info", "duration", "error", "event_version", "logs", "metadata", "output_data", "output_uri", "phase", "reason", "started_at", "task_type", "updated_at"]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    created_at: _timestamp_pb2.Timestamp
    custom_info: _struct_pb2.Struct
    duration: _duration_pb2.Duration
    error: _execution_pb2.ExecutionError
    event_version: int
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    metadata: _event_pb2.TaskExecutionMetadata
    output_data: _literals_pb2.LiteralMap
    output_uri: str
    phase: _execution_pb2.TaskExecution.Phase
    reason: str
    started_at: _timestamp_pb2.Timestamp
    task_type: str
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., reason: _Optional[str] = ..., task_type: _Optional[str] = ..., metadata: _Optional[_Union[_event_pb2.TaskExecutionMetadata, _Mapping]] = ..., event_version: _Optional[int] = ...) -> None: ...

class TaskExecutionGetDataRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionGetDataResponse(_message.Message):
    __slots__ = ["full_inputs", "full_outputs", "inputs", "outputs"]
    FULL_INPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    full_inputs: _literals_pb2.LiteralMap
    full_outputs: _literals_pb2.LiteralMap
    inputs: _common_pb2.UrlBlob
    outputs: _common_pb2.UrlBlob
    def __init__(self, inputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., full_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., full_outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class TaskExecutionGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionList(_message.Message):
    __slots__ = ["task_executions", "token"]
    TASK_EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    task_executions: _containers.RepeatedCompositeFieldContainer[TaskExecution]
    token: str
    def __init__(self, task_executions: _Optional[_Iterable[_Union[TaskExecution, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class TaskExecutionListRequest(_message.Message):
    __slots__ = ["filters", "limit", "node_execution_id", "sort_by", "token"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    filters: str
    limit: int
    node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    sort_by: _common_pb2.Sort
    token: str
    def __init__(self, node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...
