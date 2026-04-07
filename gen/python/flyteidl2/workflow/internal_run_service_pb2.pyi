from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.rpc import status_pb2 as _status_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecordActionRequest(_message.Message):
    __slots__ = ["action_id", "parent", "group", "subject", "input_uri", "task", "trace", "condition"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent: str
    group: str
    subject: str
    input_uri: str
    task: _run_definition_pb2.TaskAction
    trace: _run_definition_pb2.TraceAction
    condition: _run_definition_pb2.ConditionAction
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent: _Optional[str] = ..., group: _Optional[str] = ..., subject: _Optional[str] = ..., input_uri: _Optional[str] = ..., task: _Optional[_Union[_run_definition_pb2.TaskAction, _Mapping]] = ..., trace: _Optional[_Union[_run_definition_pb2.TraceAction, _Mapping]] = ..., condition: _Optional[_Union[_run_definition_pb2.ConditionAction, _Mapping]] = ...) -> None: ...

class RecordActionResponse(_message.Message):
    __slots__ = ["action_id", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _status_pb2.Status
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class RecordActionStreamRequest(_message.Message):
    __slots__ = ["request"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    request: RecordActionRequest
    def __init__(self, request: _Optional[_Union[RecordActionRequest, _Mapping]] = ...) -> None: ...

class RecordActionStreamResponse(_message.Message):
    __slots__ = ["response"]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    response: RecordActionResponse
    def __init__(self, response: _Optional[_Union[RecordActionResponse, _Mapping]] = ...) -> None: ...

class UpdateActionStatusRequest(_message.Message):
    __slots__ = ["action_id", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _run_definition_pb2.ActionStatus
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_run_definition_pb2.ActionStatus, _Mapping]] = ...) -> None: ...

class UpdateActionStatusResponse(_message.Message):
    __slots__ = ["action_id", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _status_pb2.Status
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class UpdateActionStatusStreamRequest(_message.Message):
    __slots__ = ["request", "nonce"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    request: UpdateActionStatusRequest
    nonce: int
    def __init__(self, request: _Optional[_Union[UpdateActionStatusRequest, _Mapping]] = ..., nonce: _Optional[int] = ...) -> None: ...

class UpdateActionStatusStreamResponse(_message.Message):
    __slots__ = ["response", "nonce"]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    response: UpdateActionStatusResponse
    nonce: int
    def __init__(self, response: _Optional[_Union[UpdateActionStatusResponse, _Mapping]] = ..., nonce: _Optional[int] = ...) -> None: ...

class RecordActionEventRequest(_message.Message):
    __slots__ = ["event"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _run_definition_pb2.ActionEvent
    def __init__(self, event: _Optional[_Union[_run_definition_pb2.ActionEvent, _Mapping]] = ...) -> None: ...

class RecordActionEventResponse(_message.Message):
    __slots__ = ["action_id", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _status_pb2.Status
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class RecordActionEventsRequest(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.ActionEvent]
    def __init__(self, events: _Optional[_Iterable[_Union[_run_definition_pb2.ActionEvent, _Mapping]]] = ...) -> None: ...

class RecordActionEventsResponse(_message.Message):
    __slots__ = ["status"]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: _status_pb2.Status
    def __init__(self, status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class RecordActionEventStreamRequest(_message.Message):
    __slots__ = ["request", "nonce"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    request: RecordActionEventRequest
    nonce: int
    def __init__(self, request: _Optional[_Union[RecordActionEventRequest, _Mapping]] = ..., nonce: _Optional[int] = ...) -> None: ...

class RecordActionEventStreamResponse(_message.Message):
    __slots__ = ["response", "nonce"]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    response: RecordActionEventResponse
    nonce: int
    def __init__(self, response: _Optional[_Union[RecordActionEventResponse, _Mapping]] = ..., nonce: _Optional[int] = ...) -> None: ...
