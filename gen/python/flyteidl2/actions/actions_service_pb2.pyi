from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
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
    __slots__ = ["action_id", "attempt", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    status: _run_definition_pb2.ActionStatus
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., status: _Optional[_Union[_run_definition_pb2.ActionStatus, _Mapping]] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

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
    __slots__ = ["action_id", "reason", "aborted_by"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    ABORTED_BY_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    reason: str
    aborted_by: _identity_pb2.EnrichedIdentity
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., reason: _Optional[str] = ..., aborted_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...

class AbortResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class SignalRequest(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "value", "signalled_by"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    SIGNALLED_BY_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    value: _literals_pb2.Literal
    signalled_by: _identity_pb2.EnrichedIdentity
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ..., signalled_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...

class SignalResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
