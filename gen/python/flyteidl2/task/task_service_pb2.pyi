from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeployTaskRequest(_message.Message):
    __slots__ = ["task_id", "spec", "triggers"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    TRIGGERS_FIELD_NUMBER: _ClassVar[int]
    task_id: _task_definition_pb2.TaskIdentifier
    spec: _task_definition_pb2.TaskSpec
    triggers: _containers.RepeatedCompositeFieldContainer[_task_definition_pb2.TaskTrigger]
    def __init__(self, task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., triggers: _Optional[_Iterable[_Union[_task_definition_pb2.TaskTrigger, _Mapping]]] = ...) -> None: ...

class DeployTaskResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetTaskDetailsRequest(_message.Message):
    __slots__ = ["task_id"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class GetTaskDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _task_definition_pb2.TaskDetails
    def __init__(self, details: _Optional[_Union[_task_definition_pb2.TaskDetails, _Mapping]] = ...) -> None: ...

class ListTasksRequest(_message.Message):
    __slots__ = ["request", "org", "project_id", "known_filters"]
    class KnownFilter(_message.Message):
        __slots__ = ["deployed_by"]
        DEPLOYED_BY_FIELD_NUMBER: _ClassVar[int]
        deployed_by: str
        def __init__(self, deployed_by: _Optional[str] = ...) -> None: ...
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    KNOWN_FILTERS_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    known_filters: _containers.RepeatedCompositeFieldContainer[ListTasksRequest.KnownFilter]
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., known_filters: _Optional[_Iterable[_Union[ListTasksRequest.KnownFilter, _Mapping]]] = ...) -> None: ...

class ListTasksResponse(_message.Message):
    __slots__ = ["tasks", "token"]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    tasks: _containers.RepeatedCompositeFieldContainer[_task_definition_pb2.Task]
    token: str
    def __init__(self, tasks: _Optional[_Iterable[_Union[_task_definition_pb2.Task, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...
