from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EnqueueActionRequest(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "run_spec", "input_uri", "run_output_base", "group", "subject", "task", "trace", "condition"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    RUN_OUTPUT_BASE_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    run_spec: _run_pb2.RunSpec
    input_uri: str
    run_output_base: str
    group: str
    subject: str
    task: _run_definition_pb2.TaskAction
    trace: _run_definition_pb2.TraceAction
    condition: _run_definition_pb2.ConditionAction
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ..., input_uri: _Optional[str] = ..., run_output_base: _Optional[str] = ..., group: _Optional[str] = ..., subject: _Optional[str] = ..., task: _Optional[_Union[_run_definition_pb2.TaskAction, _Mapping]] = ..., trace: _Optional[_Union[_run_definition_pb2.TraceAction, _Mapping]] = ..., condition: _Optional[_Union[_run_definition_pb2.ConditionAction, _Mapping]] = ...) -> None: ...

class EnqueueActionResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class AbortQueuedRunRequest(_message.Message):
    __slots__ = ["run_id", "reason"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    reason: str
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortQueuedRunResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class AbortQueuedActionRequest(_message.Message):
    __slots__ = ["action_id", "reason"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    reason: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortQueuedActionResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
