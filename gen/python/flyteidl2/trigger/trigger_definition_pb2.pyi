from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TriggerRevisionAction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    TRIGGER_REVISION_ACTION_UNSPECIFIED: _ClassVar[TriggerRevisionAction]
    TRIGGER_REVISION_ACTION_DEPLOY: _ClassVar[TriggerRevisionAction]
    TRIGGER_REVISION_ACTION_ACTIVATE: _ClassVar[TriggerRevisionAction]
    TRIGGER_REVISION_ACTION_DEACTIVATE: _ClassVar[TriggerRevisionAction]
    TRIGGER_REVISION_ACTION_DELETE: _ClassVar[TriggerRevisionAction]
TRIGGER_REVISION_ACTION_UNSPECIFIED: TriggerRevisionAction
TRIGGER_REVISION_ACTION_DEPLOY: TriggerRevisionAction
TRIGGER_REVISION_ACTION_ACTIVATE: TriggerRevisionAction
TRIGGER_REVISION_ACTION_DEACTIVATE: TriggerRevisionAction
TRIGGER_REVISION_ACTION_DELETE: TriggerRevisionAction

class TriggerMetadata(_message.Message):
    __slots__ = ["deployed_by", "updated_by"]
    DEPLOYED_BY_FIELD_NUMBER: _ClassVar[int]
    UPDATED_BY_FIELD_NUMBER: _ClassVar[int]
    deployed_by: _identity_pb2.EnrichedIdentity
    updated_by: _identity_pb2.EnrichedIdentity
    def __init__(self, deployed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., updated_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...

class TriggerSpec(_message.Message):
    __slots__ = ["task_id", "inputs", "run_spec", "active", "task_version"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_FIELD_NUMBER: _ClassVar[int]
    TASK_VERSION_FIELD_NUMBER: _ClassVar[int]
    task_id: _task_definition_pb2.TaskIdentifier
    inputs: _common_pb2.Inputs
    run_spec: _run_pb2.RunSpec
    active: bool
    task_version: str
    def __init__(self, task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ..., active: bool = ..., task_version: _Optional[str] = ...) -> None: ...

class TriggerStatus(_message.Message):
    __slots__ = ["deployed_at", "updated_at", "triggered_at", "deleted_at"]
    DEPLOYED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    TRIGGERED_AT_FIELD_NUMBER: _ClassVar[int]
    DELETED_AT_FIELD_NUMBER: _ClassVar[int]
    deployed_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    triggered_at: _timestamp_pb2.Timestamp
    deleted_at: _timestamp_pb2.Timestamp
    def __init__(self, deployed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., triggered_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., deleted_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class TriggerRevision(_message.Message):
    __slots__ = ["id", "metadata", "status", "action"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TriggerIdentifier
    metadata: TriggerMetadata
    status: TriggerStatus
    action: TriggerRevisionAction
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[TriggerMetadata, _Mapping]] = ..., status: _Optional[_Union[TriggerStatus, _Mapping]] = ..., action: _Optional[_Union[TriggerRevisionAction, str]] = ...) -> None: ...

class TriggerDetails(_message.Message):
    __slots__ = ["id", "metadata", "spec", "status", "automation_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    AUTOMATION_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TriggerIdentifier
    metadata: TriggerMetadata
    spec: TriggerSpec
    status: TriggerStatus
    automation_spec: _common_pb2.TriggerAutomationSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[TriggerMetadata, _Mapping]] = ..., spec: _Optional[_Union[TriggerSpec, _Mapping]] = ..., status: _Optional[_Union[TriggerStatus, _Mapping]] = ..., automation_spec: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ...) -> None: ...

class Trigger(_message.Message):
    __slots__ = ["id", "metadata", "status", "active", "automation_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_FIELD_NUMBER: _ClassVar[int]
    AUTOMATION_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TriggerIdentifier
    metadata: TriggerMetadata
    status: TriggerStatus
    active: bool
    automation_spec: _common_pb2.TriggerAutomationSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[TriggerMetadata, _Mapping]] = ..., status: _Optional[_Union[TriggerStatus, _Mapping]] = ..., active: bool = ..., automation_spec: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ...) -> None: ...
