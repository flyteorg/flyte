from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import run_pb2 as _run_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2_1
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from flyteidl2.workflow import run_service_pb2 as _run_service_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.rpc import status_pb2 as _status_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateLocalRunRequest(_message.Message):
    __slots__ = ["run_id", "project_id", "task_id", "task_spec", "offloaded_input_data", "run_spec", "labels", "run_start_time"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    OFFLOADED_INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    RUN_START_TIME_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    task_spec: _task_definition_pb2.TaskSpec
    offloaded_input_data: _run_pb2.OffloadedInputData
    run_spec: _run_pb2_1.RunSpec
    labels: _containers.ScalarMap[str, str]
    run_start_time: _timestamp_pb2.Timestamp
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., offloaded_input_data: _Optional[_Union[_run_pb2.OffloadedInputData, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2_1.RunSpec, _Mapping]] = ..., labels: _Optional[_Mapping[str, str]] = ..., run_start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class LocalActionUpdate(_message.Message):
    __slots__ = ["event", "parent_name", "group", "task", "trace", "status"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    PARENT_NAME_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    TASK_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    event: _run_definition_pb2.ActionEvent
    parent_name: str
    group: str
    task: _run_definition_pb2.TaskAction
    trace: _run_definition_pb2.TraceAction
    status: _run_definition_pb2.ActionStatus
    def __init__(self, event: _Optional[_Union[_run_definition_pb2.ActionEvent, _Mapping]] = ..., parent_name: _Optional[str] = ..., group: _Optional[str] = ..., task: _Optional[_Union[_run_definition_pb2.TaskAction, _Mapping]] = ..., trace: _Optional[_Union[_run_definition_pb2.TraceAction, _Mapping]] = ..., status: _Optional[_Union[_run_definition_pb2.ActionStatus, _Mapping]] = ...) -> None: ...

class ReportLocalActionsRequest(_message.Message):
    __slots__ = ["run_id", "updates"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATES_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    updates: _containers.RepeatedCompositeFieldContainer[LocalActionUpdate]
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., updates: _Optional[_Iterable[_Union[LocalActionUpdate, _Mapping]]] = ...) -> None: ...

class ReportLocalActionsResponse(_message.Message):
    __slots__ = ["statuses"]
    STATUSES_FIELD_NUMBER: _ClassVar[int]
    statuses: _containers.RepeatedCompositeFieldContainer[_status_pb2.Status]
    def __init__(self, statuses: _Optional[_Iterable[_Union[_status_pb2.Status, _Mapping]]] = ...) -> None: ...
