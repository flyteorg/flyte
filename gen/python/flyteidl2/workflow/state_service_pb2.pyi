from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.core import execution_pb2 as _execution_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.rpc import status_pb2 as _status_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PutRequest(_message.Message):
    __slots__ = ["action_id", "parent_action_name", "state"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_ACTION_NAME_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    parent_action_name: str
    state: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., parent_action_name: _Optional[str] = ..., state: _Optional[str] = ...) -> None: ...

class PutResponse(_message.Message):
    __slots__ = ["action_id", "status"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _status_pb2.Status
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["action_id", "status", "state"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    status: _status_pb2.Status
    state: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., status: _Optional[_Union[_status_pb2.Status, _Mapping]] = ..., state: _Optional[str] = ...) -> None: ...

class WatchRequest(_message.Message):
    __slots__ = ["parent_action_id"]
    PARENT_ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    parent_action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, parent_action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class WatchResponse(_message.Message):
    __slots__ = ["action_update", "control_message"]
    ACTION_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONTROL_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    action_update: ActionUpdate
    control_message: ControlMessage
    def __init__(self, action_update: _Optional[_Union[ActionUpdate, _Mapping]] = ..., control_message: _Optional[_Union[ControlMessage, _Mapping]] = ...) -> None: ...

class ControlMessage(_message.Message):
    __slots__ = ["sentinel"]
    SENTINEL_FIELD_NUMBER: _ClassVar[int]
    sentinel: bool
    def __init__(self, sentinel: bool = ...) -> None: ...

class ActionUpdate(_message.Message):
    __slots__ = ["action_id", "phase", "error", "output_uri"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    phase: _run_definition_pb2.Phase
    error: _execution_pb2.ExecutionError
    output_uri: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., phase: _Optional[_Union[_run_definition_pb2.Phase, str]] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_uri: _Optional[str] = ...) -> None: ...
