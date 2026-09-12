from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
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
        __slots__ = ["deployed_by", "is_entrypoint"]
        DEPLOYED_BY_FIELD_NUMBER: _ClassVar[int]
        IS_ENTRYPOINT_FIELD_NUMBER: _ClassVar[int]
        deployed_by: str
        is_entrypoint: bool
        def __init__(self, deployed_by: _Optional[str] = ..., is_entrypoint: bool = ...) -> None: ...
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
    __slots__ = ["tasks", "token", "metadata"]
    class ListTasksMetadata(_message.Message):
        __slots__ = ["total", "filtered_total"]
        TOTAL_FIELD_NUMBER: _ClassVar[int]
        FILTERED_TOTAL_FIELD_NUMBER: _ClassVar[int]
        total: int
        filtered_total: int
        def __init__(self, total: _Optional[int] = ..., filtered_total: _Optional[int] = ...) -> None: ...
    TASKS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    tasks: _containers.RepeatedCompositeFieldContainer[_task_definition_pb2.Task]
    token: str
    metadata: ListTasksResponse.ListTasksMetadata
    def __init__(self, tasks: _Optional[_Iterable[_Union[_task_definition_pb2.Task, _Mapping]]] = ..., token: _Optional[str] = ..., metadata: _Optional[_Union[ListTasksResponse.ListTasksMetadata, _Mapping]] = ...) -> None: ...

class ListVersionsRequest(_message.Message):
    __slots__ = ["request", "task_name"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    TASK_NAME_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    task_name: _task_definition_pb2.TaskName
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., task_name: _Optional[_Union[_task_definition_pb2.TaskName, _Mapping]] = ...) -> None: ...

class ListVersionsResponse(_message.Message):
    __slots__ = ["versions", "token"]
    class VersionResponse(_message.Message):
        __slots__ = ["version", "deployed_at", "deployed_by", "latest_run"]
        VERSION_FIELD_NUMBER: _ClassVar[int]
        DEPLOYED_AT_FIELD_NUMBER: _ClassVar[int]
        DEPLOYED_BY_FIELD_NUMBER: _ClassVar[int]
        LATEST_RUN_FIELD_NUMBER: _ClassVar[int]
        version: str
        deployed_at: _timestamp_pb2.Timestamp
        deployed_by: _identity_pb2.EnrichedIdentity
        latest_run: _task_definition_pb2.LatestRunSummary
        def __init__(self, version: _Optional[str] = ..., deployed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., deployed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., latest_run: _Optional[_Union[_task_definition_pb2.LatestRunSummary, _Mapping]] = ...) -> None: ...
    VERSIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    versions: _containers.RepeatedCompositeFieldContainer[ListVersionsResponse.VersionResponse]
    token: str
    def __init__(self, versions: _Optional[_Iterable[_Union[ListVersionsResponse.VersionResponse, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class TaskAliasName(_message.Message):
    __slots__ = ["task_name", "alias"]
    TASK_NAME_FIELD_NUMBER: _ClassVar[int]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    task_name: _task_definition_pb2.TaskName
    alias: str
    def __init__(self, task_name: _Optional[_Union[_task_definition_pb2.TaskName, _Mapping]] = ..., alias: _Optional[str] = ...) -> None: ...

class TaskAlias(_message.Message):
    __slots__ = ["name", "version", "set_by", "set_at"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    SET_BY_FIELD_NUMBER: _ClassVar[int]
    SET_AT_FIELD_NUMBER: _ClassVar[int]
    name: TaskAliasName
    version: str
    set_by: _identity_pb2.EnrichedIdentity
    set_at: _timestamp_pb2.Timestamp
    def __init__(self, name: _Optional[_Union[TaskAliasName, _Mapping]] = ..., version: _Optional[str] = ..., set_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., set_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class TaskAliasRevision(_message.Message):
    __slots__ = ["from_version", "to_version", "changed_by", "changed_at"]
    FROM_VERSION_FIELD_NUMBER: _ClassVar[int]
    TO_VERSION_FIELD_NUMBER: _ClassVar[int]
    CHANGED_BY_FIELD_NUMBER: _ClassVar[int]
    CHANGED_AT_FIELD_NUMBER: _ClassVar[int]
    from_version: str
    to_version: str
    changed_by: _identity_pb2.EnrichedIdentity
    changed_at: _timestamp_pb2.Timestamp
    def __init__(self, from_version: _Optional[str] = ..., to_version: _Optional[str] = ..., changed_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., changed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class SetTaskAliasRequest(_message.Message):
    __slots__ = ["name", "version"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: TaskAliasName
    version: str
    def __init__(self, name: _Optional[_Union[TaskAliasName, _Mapping]] = ..., version: _Optional[str] = ...) -> None: ...

class SetTaskAliasResponse(_message.Message):
    __slots__ = ["alias", "previous_version"]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_VERSION_FIELD_NUMBER: _ClassVar[int]
    alias: TaskAlias
    previous_version: str
    def __init__(self, alias: _Optional[_Union[TaskAlias, _Mapping]] = ..., previous_version: _Optional[str] = ...) -> None: ...

class GetTaskAliasRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: TaskAliasName
    def __init__(self, name: _Optional[_Union[TaskAliasName, _Mapping]] = ...) -> None: ...

class GetTaskAliasResponse(_message.Message):
    __slots__ = ["alias"]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    alias: TaskAlias
    def __init__(self, alias: _Optional[_Union[TaskAlias, _Mapping]] = ...) -> None: ...

class ListTaskAliasesRequest(_message.Message):
    __slots__ = ["task_name", "request"]
    TASK_NAME_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    task_name: _task_definition_pb2.TaskName
    request: _list_pb2.ListRequest
    def __init__(self, task_name: _Optional[_Union[_task_definition_pb2.TaskName, _Mapping]] = ..., request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ...) -> None: ...

class ListTaskAliasesResponse(_message.Message):
    __slots__ = ["aliases", "token"]
    ALIASES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    aliases: _containers.RepeatedCompositeFieldContainer[TaskAlias]
    token: str
    def __init__(self, aliases: _Optional[_Iterable[_Union[TaskAlias, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class DeleteTaskAliasRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: TaskAliasName
    def __init__(self, name: _Optional[_Union[TaskAliasName, _Mapping]] = ...) -> None: ...

class DeleteTaskAliasResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetTaskAliasHistoryRequest(_message.Message):
    __slots__ = ["name", "request"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    name: TaskAliasName
    request: _list_pb2.ListRequest
    def __init__(self, name: _Optional[_Union[TaskAliasName, _Mapping]] = ..., request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ...) -> None: ...

class GetTaskAliasHistoryResponse(_message.Message):
    __slots__ = ["revisions", "token"]
    REVISIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    revisions: _containers.RepeatedCompositeFieldContainer[TaskAliasRevision]
    token: str
    def __init__(self, revisions: _Optional[_Iterable[_Union[TaskAliasRevision, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...
