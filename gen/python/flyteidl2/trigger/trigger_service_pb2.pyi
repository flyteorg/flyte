from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from flyteidl2.trigger import trigger_definition_pb2 as _trigger_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeployTriggerRequest(_message.Message):
    __slots__ = ["id", "spec", "automation_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    AUTOMATION_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TriggerIdentifier
    spec: _trigger_definition_pb2.TriggerSpec
    automation_spec: _common_pb2.TriggerAutomationSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ..., spec: _Optional[_Union[_trigger_definition_pb2.TriggerSpec, _Mapping]] = ..., automation_spec: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ...) -> None: ...

class DeployTriggerResponse(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_definition_pb2.TriggerDetails
    def __init__(self, trigger: _Optional[_Union[_trigger_definition_pb2.TriggerDetails, _Mapping]] = ...) -> None: ...

class GetTriggerDetailsRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: _identifier_pb2.TriggerName
    def __init__(self, name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ...) -> None: ...

class GetTriggerDetailsResponse(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_definition_pb2.TriggerDetails
    def __init__(self, trigger: _Optional[_Union[_trigger_definition_pb2.TriggerDetails, _Mapping]] = ...) -> None: ...

class GetTriggerRevisionDetailsRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TriggerIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ...) -> None: ...

class GetTriggerRevisionDetailsResponse(_message.Message):
    __slots__ = ["trigger"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    trigger: _trigger_definition_pb2.TriggerDetails
    def __init__(self, trigger: _Optional[_Union[_trigger_definition_pb2.TriggerDetails, _Mapping]] = ...) -> None: ...

class ListTriggersRequest(_message.Message):
    __slots__ = ["request", "org", "project_id", "task_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class ListTriggersResponse(_message.Message):
    __slots__ = ["triggers", "token"]
    TRIGGERS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    triggers: _containers.RepeatedCompositeFieldContainer[_trigger_definition_pb2.Trigger]
    token: str
    def __init__(self, triggers: _Optional[_Iterable[_Union[_trigger_definition_pb2.Trigger, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class GetTriggerRevisionHistoryRequest(_message.Message):
    __slots__ = ["request", "name"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    name: _identifier_pb2.TriggerName
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ...) -> None: ...

class GetTriggerRevisionHistoryResponse(_message.Message):
    __slots__ = ["triggers", "token"]
    TRIGGERS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    triggers: _containers.RepeatedCompositeFieldContainer[_trigger_definition_pb2.TriggerRevision]
    token: str
    def __init__(self, triggers: _Optional[_Iterable[_Union[_trigger_definition_pb2.TriggerRevision, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class UpdateTriggersRequest(_message.Message):
    __slots__ = ["names", "active"]
    NAMES_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_FIELD_NUMBER: _ClassVar[int]
    names: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.TriggerName]
    active: bool
    def __init__(self, names: _Optional[_Iterable[_Union[_identifier_pb2.TriggerName, _Mapping]]] = ..., active: bool = ...) -> None: ...

class UpdateTriggersResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteTriggersRequest(_message.Message):
    __slots__ = ["names"]
    NAMES_FIELD_NUMBER: _ClassVar[int]
    names: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.TriggerName]
    def __init__(self, names: _Optional[_Iterable[_Union[_identifier_pb2.TriggerName, _Mapping]]] = ...) -> None: ...

class DeleteTriggersResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
