from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.core import interface_pb2 as _interface_pb2
from flyteidl2.core import tasks_pb2 as _tasks_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import environment_pb2 as _environment_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TaskName(_message.Message):
    __slots__ = ["org", "project", "domain", "name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class TaskIdentifier(_message.Message):
    __slots__ = ["org", "project", "domain", "name", "version"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    version: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class TaskTriggersSummary(_message.Message):
    __slots__ = ["automation_spec", "stats"]
    class TriggerStats(_message.Message):
        __slots__ = ["total", "active"]
        TOTAL_FIELD_NUMBER: _ClassVar[int]
        ACTIVE_FIELD_NUMBER: _ClassVar[int]
        total: int
        active: int
        def __init__(self, total: _Optional[int] = ..., active: _Optional[int] = ...) -> None: ...
    AUTOMATION_SPEC_FIELD_NUMBER: _ClassVar[int]
    STATS_FIELD_NUMBER: _ClassVar[int]
    automation_spec: _common_pb2.TriggerAutomationSpec
    stats: TaskTriggersSummary.TriggerStats
    def __init__(self, automation_spec: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ..., stats: _Optional[_Union[TaskTriggersSummary.TriggerStats, _Mapping]] = ...) -> None: ...

class TaskMetadata(_message.Message):
    __slots__ = ["deployed_by", "short_name", "deployed_at", "environment_name", "triggers_summary"]
    DEPLOYED_BY_FIELD_NUMBER: _ClassVar[int]
    SHORT_NAME_FIELD_NUMBER: _ClassVar[int]
    DEPLOYED_AT_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_NAME_FIELD_NUMBER: _ClassVar[int]
    TRIGGERS_SUMMARY_FIELD_NUMBER: _ClassVar[int]
    deployed_by: _identity_pb2.EnrichedIdentity
    short_name: str
    deployed_at: _timestamp_pb2.Timestamp
    environment_name: str
    triggers_summary: TaskTriggersSummary
    def __init__(self, deployed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., short_name: _Optional[str] = ..., deployed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., environment_name: _Optional[str] = ..., triggers_summary: _Optional[_Union[TaskTriggersSummary, _Mapping]] = ...) -> None: ...

class Task(_message.Message):
    __slots__ = ["task_id", "metadata"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    task_id: TaskIdentifier
    metadata: TaskMetadata
    def __init__(self, task_id: _Optional[_Union[TaskIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[TaskMetadata, _Mapping]] = ...) -> None: ...

class TaskSpec(_message.Message):
    __slots__ = ["task_template", "default_inputs", "short_name", "environment"]
    TASK_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_INPUTS_FIELD_NUMBER: _ClassVar[int]
    SHORT_NAME_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    task_template: _tasks_pb2.TaskTemplate
    default_inputs: _containers.RepeatedCompositeFieldContainer[_common_pb2.NamedParameter]
    short_name: str
    environment: _environment_pb2.Environment
    def __init__(self, task_template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., default_inputs: _Optional[_Iterable[_Union[_common_pb2.NamedParameter, _Mapping]]] = ..., short_name: _Optional[str] = ..., environment: _Optional[_Union[_environment_pb2.Environment, _Mapping]] = ...) -> None: ...

class TraceSpec(_message.Message):
    __slots__ = ["interface"]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    interface: _interface_pb2.TypedInterface
    def __init__(self, interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ...) -> None: ...

class TaskDetails(_message.Message):
    __slots__ = ["task_id", "metadata", "spec"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    task_id: TaskIdentifier
    metadata: TaskMetadata
    spec: TaskSpec
    def __init__(self, task_id: _Optional[_Union[TaskIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[TaskMetadata, _Mapping]] = ..., spec: _Optional[_Union[TaskSpec, _Mapping]] = ...) -> None: ...

class TaskTrigger(_message.Message):
    __slots__ = ["name", "spec", "automation_spec"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    AUTOMATION_SPEC_FIELD_NUMBER: _ClassVar[int]
    name: str
    spec: TaskTriggerSpec
    automation_spec: _common_pb2.TriggerAutomationSpec
    def __init__(self, name: _Optional[str] = ..., spec: _Optional[_Union[TaskTriggerSpec, _Mapping]] = ..., automation_spec: _Optional[_Union[_common_pb2.TriggerAutomationSpec, _Mapping]] = ...) -> None: ...

class TaskTriggerSpec(_message.Message):
    __slots__ = ["active", "inputs", "run_spec"]
    ACTIVE_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    active: bool
    inputs: _common_pb2.Inputs
    run_spec: _run_pb2.RunSpec
    def __init__(self, active: bool = ..., inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ...) -> None: ...
