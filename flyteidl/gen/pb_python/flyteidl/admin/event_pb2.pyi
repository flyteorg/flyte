from flyteidl.event import event_pb2 as _event_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventErrorAlreadyInTerminalState(_message.Message):
    __slots__ = ["current_phase"]
    CURRENT_PHASE_FIELD_NUMBER: _ClassVar[int]
    current_phase: str
    def __init__(self, current_phase: _Optional[str] = ...) -> None: ...

class EventErrorIncompatibleCluster(_message.Message):
    __slots__ = ["cluster"]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    cluster: str
    def __init__(self, cluster: _Optional[str] = ...) -> None: ...

class EventFailureReason(_message.Message):
    __slots__ = ["already_in_terminal_state", "incompatible_cluster"]
    ALREADY_IN_TERMINAL_STATE_FIELD_NUMBER: _ClassVar[int]
    INCOMPATIBLE_CLUSTER_FIELD_NUMBER: _ClassVar[int]
    already_in_terminal_state: EventErrorAlreadyInTerminalState
    incompatible_cluster: EventErrorIncompatibleCluster
    def __init__(self, already_in_terminal_state: _Optional[_Union[EventErrorAlreadyInTerminalState, _Mapping]] = ..., incompatible_cluster: _Optional[_Union[EventErrorIncompatibleCluster, _Mapping]] = ...) -> None: ...

class WorkflowExecutionEventRequest(_message.Message):
    __slots__ = ["request_id", "event"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    event: _event_pb2.WorkflowExecutionEvent
    def __init__(self, request_id: _Optional[str] = ..., event: _Optional[_Union[_event_pb2.WorkflowExecutionEvent, _Mapping]] = ...) -> None: ...

class WorkflowExecutionEventResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class NodeExecutionEventRequest(_message.Message):
    __slots__ = ["request_id", "event"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    event: _event_pb2.NodeExecutionEvent
    def __init__(self, request_id: _Optional[str] = ..., event: _Optional[_Union[_event_pb2.NodeExecutionEvent, _Mapping]] = ...) -> None: ...

class NodeExecutionEventResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class TaskExecutionEventRequest(_message.Message):
    __slots__ = ["request_id", "event"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    request_id: str
    event: _event_pb2.TaskExecutionEvent
    def __init__(self, request_id: _Optional[str] = ..., event: _Optional[_Union[_event_pb2.TaskExecutionEvent, _Mapping]] = ...) -> None: ...

class TaskExecutionEventResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
