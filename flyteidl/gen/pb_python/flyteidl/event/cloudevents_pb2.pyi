from flyteidl.event import event_pb2 as _event_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.core import artifact_id_pb2 as _artifact_id_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CloudEventWorkflowExecution(_message.Message):
    __slots__ = ["raw_event", "output_interface", "artifact_ids", "reference_execution", "principal", "launch_plan_id", "labels"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RAW_EVENT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_INTERFACE_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_IDS_FIELD_NUMBER: _ClassVar[int]
    REFERENCE_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    LAUNCH_PLAN_ID_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    raw_event: _event_pb2.WorkflowExecutionEvent
    output_interface: _interface_pb2.TypedInterface
    artifact_ids: _containers.RepeatedCompositeFieldContainer[_artifact_id_pb2.ArtifactID]
    reference_execution: _identifier_pb2.WorkflowExecutionIdentifier
    principal: str
    launch_plan_id: _identifier_pb2.Identifier
    labels: _containers.ScalarMap[str, str]
    def __init__(self, raw_event: _Optional[_Union[_event_pb2.WorkflowExecutionEvent, _Mapping]] = ..., output_interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ..., artifact_ids: _Optional[_Iterable[_Union[_artifact_id_pb2.ArtifactID, _Mapping]]] = ..., reference_execution: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., principal: _Optional[str] = ..., launch_plan_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., labels: _Optional[_Mapping[str, str]] = ...) -> None: ...

class CloudEventNodeExecution(_message.Message):
    __slots__ = ["raw_event", "task_exec_id", "output_interface", "artifact_ids", "principal", "launch_plan_id", "labels"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RAW_EVENT_FIELD_NUMBER: _ClassVar[int]
    TASK_EXEC_ID_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_INTERFACE_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_IDS_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    LAUNCH_PLAN_ID_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    raw_event: _event_pb2.NodeExecutionEvent
    task_exec_id: _identifier_pb2.TaskExecutionIdentifier
    output_interface: _interface_pb2.TypedInterface
    artifact_ids: _containers.RepeatedCompositeFieldContainer[_artifact_id_pb2.ArtifactID]
    principal: str
    launch_plan_id: _identifier_pb2.Identifier
    labels: _containers.ScalarMap[str, str]
    def __init__(self, raw_event: _Optional[_Union[_event_pb2.NodeExecutionEvent, _Mapping]] = ..., task_exec_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., output_interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ..., artifact_ids: _Optional[_Iterable[_Union[_artifact_id_pb2.ArtifactID, _Mapping]]] = ..., principal: _Optional[str] = ..., launch_plan_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., labels: _Optional[_Mapping[str, str]] = ...) -> None: ...

class CloudEventTaskExecution(_message.Message):
    __slots__ = ["raw_event", "labels"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RAW_EVENT_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    raw_event: _event_pb2.TaskExecutionEvent
    labels: _containers.ScalarMap[str, str]
    def __init__(self, raw_event: _Optional[_Union[_event_pb2.TaskExecutionEvent, _Mapping]] = ..., labels: _Optional[_Mapping[str, str]] = ...) -> None: ...

class CloudEventExecutionStart(_message.Message):
    __slots__ = ["execution_id", "launch_plan_id", "workflow_id", "artifact_ids", "artifact_trackers", "principal"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    LAUNCH_PLAN_ID_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_IDS_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_TRACKERS_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    launch_plan_id: _identifier_pb2.Identifier
    workflow_id: _identifier_pb2.Identifier
    artifact_ids: _containers.RepeatedCompositeFieldContainer[_artifact_id_pb2.ArtifactID]
    artifact_trackers: _containers.RepeatedScalarFieldContainer[str]
    principal: str
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., launch_plan_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., workflow_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., artifact_ids: _Optional[_Iterable[_Union[_artifact_id_pb2.ArtifactID, _Mapping]]] = ..., artifact_trackers: _Optional[_Iterable[str]] = ..., principal: _Optional[str] = ...) -> None: ...
