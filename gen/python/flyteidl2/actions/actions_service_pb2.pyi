from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.core import literals_pb2 as _literals_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from flyteidl2.workflow import state_service_pb2 as _state_service_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Action(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "input_uri", "run_output_base", "group", "subject", "task", "trace", "condition"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    RUN_OUTPUT_BASE_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    input_uri: str
    run_output_base: str
    group: str
    subject: str
    task: _run_definition_pb2.TaskAction
    trace: _run_definition_pb2.TraceAction
    condition: _run_definition_pb2.ConditionAction
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., input_uri: _Optional[str] = ..., run_output_base: _Optional[str] = ..., group: _Optional[str] = ..., subject: _Optional[str] = ..., task: _Optional[_Union[_run_definition_pb2.TaskAction, _Mapping]] = ..., trace: _Optional[_Union[_run_definition_pb2.TraceAction, _Mapping]] = ..., condition: _Optional[_Union[_run_definition_pb2.ConditionAction, _Mapping]] = ...) -> None: ...

class EnqueueRequest(_message.Message):
    __slots__ = ["action", "run_spec"]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    action: Action
    run_spec: _run_pb2.RunSpec
    def __init__(self, action: _Optional[_Union[Action, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ...) -> None: ...

class EnqueueResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["action_id", "attempt", "status", "state"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    status: _run_definition_pb2.ActionStatus
    state: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., status: _Optional[_Union[_run_definition_pb2.ActionStatus, _Mapping]] = ..., state: _Optional[str] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetLatestStateRequest(_message.Message):
    __slots__ = ["action_id", "attempt"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class GetLatestStateResponse(_message.Message):
    __slots__ = ["state"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    state: str
    def __init__(self, state: _Optional[str] = ...) -> None: ...

class WatchForUpdatesRequest(_message.Message):
    __slots__ = ["parent_action_id"]
    PARENT_ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    parent_action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, parent_action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class WatchForUpdatesResponse(_message.Message):
    __slots__ = ["action_update", "control_message"]
    ACTION_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONTROL_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    action_update: _state_service_pb2.ActionUpdate
    control_message: _state_service_pb2.ControlMessage
    def __init__(self, action_update: _Optional[_Union[_state_service_pb2.ActionUpdate, _Mapping]] = ..., control_message: _Optional[_Union[_state_service_pb2.ControlMessage, _Mapping]] = ...) -> None: ...

class AbortRequest(_message.Message):
    __slots__ = ["action_id", "reason"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    reason: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class SignalRequest(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "value"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    value: _literals_pb2.Literal
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...

class SignalResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
