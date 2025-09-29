from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRunRequest(_message.Message):
    __slots__ = ["run_id", "project_id", "task_id", "task_spec", "trigger_name", "inputs", "run_spec"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    task_spec: _task_definition_pb2.TaskSpec
    trigger_name: _identifier_pb2.TriggerName
    inputs: _common_pb2.Inputs
    run_spec: _run_pb2.RunSpec
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., trigger_name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ..., inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ...) -> None: ...

class CreateRunResponse(_message.Message):
    __slots__ = ["run"]
    RUN_FIELD_NUMBER: _ClassVar[int]
    run: _run_definition_pb2.Run
    def __init__(self, run: _Optional[_Union[_run_definition_pb2.Run, _Mapping]] = ...) -> None: ...

class AbortRunRequest(_message.Message):
    __slots__ = ["run_id", "reason"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    reason: str
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortRunResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRunDetailsRequest(_message.Message):
    __slots__ = ["run_id"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class GetRunDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.RunDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.RunDetails, _Mapping]] = ...) -> None: ...

class WatchRunDetailsRequest(_message.Message):
    __slots__ = ["run_id"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class WatchRunDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.RunDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.RunDetails, _Mapping]] = ...) -> None: ...

class GetActionDetailsRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.ActionDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.ActionDetails, _Mapping]] = ...) -> None: ...

class WatchActionDetailsRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class WatchActionDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.ActionDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.ActionDetails, _Mapping]] = ...) -> None: ...

class GetActionDataRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDataResponse(_message.Message):
    __slots__ = ["inputs", "outputs"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    inputs: _common_pb2.Inputs
    outputs: _common_pb2.Outputs
    def __init__(self, inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.Outputs, _Mapping]] = ...) -> None: ...

class ListRunsRequest(_message.Message):
    __slots__ = ["request", "org", "project_id", "trigger_name"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    trigger_name: _identifier_pb2.TriggerName
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., trigger_name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ...) -> None: ...

class ListRunsResponse(_message.Message):
    __slots__ = ["runs", "token"]
    RUNS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    runs: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Run]
    token: str
    def __init__(self, runs: _Optional[_Iterable[_Union[_run_definition_pb2.Run, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WatchRunsRequest(_message.Message):
    __slots__ = ["org", "cluster_id", "project_id", "task_id"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    org: str
    cluster_id: _identifier_pb2.ClusterIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, org: _Optional[str] = ..., cluster_id: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class WatchRunsResponse(_message.Message):
    __slots__ = ["runs"]
    RUNS_FIELD_NUMBER: _ClassVar[int]
    runs: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Run]
    def __init__(self, runs: _Optional[_Iterable[_Union[_run_definition_pb2.Run, _Mapping]]] = ...) -> None: ...

class ListActionsRequest(_message.Message):
    __slots__ = ["request", "run_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class ListActionsResponse(_message.Message):
    __slots__ = ["actions", "token"]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    actions: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Action]
    token: str
    def __init__(self, actions: _Optional[_Iterable[_Union[_run_definition_pb2.Action, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WatchActionsRequest(_message.Message):
    __slots__ = ["run_id", "filter"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    filter: _containers.RepeatedCompositeFieldContainer[_list_pb2.Filter]
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., filter: _Optional[_Iterable[_Union[_list_pb2.Filter, _Mapping]]] = ...) -> None: ...

class WatchActionsResponse(_message.Message):
    __slots__ = ["enriched_actions"]
    ENRICHED_ACTIONS_FIELD_NUMBER: _ClassVar[int]
    enriched_actions: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.EnrichedAction]
    def __init__(self, enriched_actions: _Optional[_Iterable[_Union[_run_definition_pb2.EnrichedAction, _Mapping]]] = ...) -> None: ...

class WatchClusterEventsRequest(_message.Message):
    __slots__ = ["id", "attempt"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class WatchClusterEventsResponse(_message.Message):
    __slots__ = ["cluster_events"]
    CLUSTER_EVENTS_FIELD_NUMBER: _ClassVar[int]
    cluster_events: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.ClusterEvent]
    def __init__(self, cluster_events: _Optional[_Iterable[_Union[_run_definition_pb2.ClusterEvent, _Mapping]]] = ...) -> None: ...

class AbortActionRequest(_message.Message):
    __slots__ = ["action_id", "reason"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    reason: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortActionResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
