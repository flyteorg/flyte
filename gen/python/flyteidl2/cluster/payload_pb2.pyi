from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.app import app_definition_pb2 as _app_definition_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SelectClusterRequest(_message.Message):
    __slots__ = ["org_id", "project_id", "task_id", "action_id", "action_attempt_id", "app_id", "operation"]
    class Operation(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        OPERATION_UNSPECIFIED: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_CREATE_UPLOAD_LOCATION: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_UPLOAD_INPUTS: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_GET_ACTION_DATA: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_QUERY_RANGE_METRICS: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_CREATE_DOWNLOAD_LINK: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_TAIL_LOGS: _ClassVar[SelectClusterRequest.Operation]
        OPERATION_GET_ACTION_ATTEMPT_METRICS: _ClassVar[SelectClusterRequest.Operation]
    OPERATION_UNSPECIFIED: SelectClusterRequest.Operation
    OPERATION_CREATE_UPLOAD_LOCATION: SelectClusterRequest.Operation
    OPERATION_UPLOAD_INPUTS: SelectClusterRequest.Operation
    OPERATION_GET_ACTION_DATA: SelectClusterRequest.Operation
    OPERATION_QUERY_RANGE_METRICS: SelectClusterRequest.Operation
    OPERATION_CREATE_DOWNLOAD_LINK: SelectClusterRequest.Operation
    OPERATION_TAIL_LOGS: SelectClusterRequest.Operation
    OPERATION_GET_ACTION_ATTEMPT_METRICS: SelectClusterRequest.Operation
    ORG_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_ATTEMPT_ID_FIELD_NUMBER: _ClassVar[int]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    OPERATION_FIELD_NUMBER: _ClassVar[int]
    org_id: _identifier_pb2.OrgIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    action_id: _identifier_pb2.ActionIdentifier
    action_attempt_id: _identifier_pb2.ActionAttemptIdentifier
    app_id: _app_definition_pb2.Identifier
    operation: SelectClusterRequest.Operation
    def __init__(self, org_id: _Optional[_Union[_identifier_pb2.OrgIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., action_attempt_id: _Optional[_Union[_identifier_pb2.ActionAttemptIdentifier, _Mapping]] = ..., app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ..., operation: _Optional[_Union[SelectClusterRequest.Operation, str]] = ...) -> None: ...

class SelectClusterResponse(_message.Message):
    __slots__ = ["cluster_endpoint"]
    CLUSTER_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    cluster_endpoint: str
    def __init__(self, cluster_endpoint: _Optional[str] = ...) -> None: ...
