from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowAttributes(_message.Message):
    __slots__ = ["domain", "matching_attributes", "project", "workflow"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    MATCHING_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    domain: str
    matching_attributes: _matchable_resource_pb2.MatchingAttributes
    project: str
    workflow: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., matching_attributes: _Optional[_Union[_matchable_resource_pb2.MatchingAttributes, _Mapping]] = ...) -> None: ...

class WorkflowAttributesDeleteRequest(_message.Message):
    __slots__ = ["domain", "project", "resource_type", "workflow"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    domain: str
    project: str
    resource_type: _matchable_resource_pb2.MatchableResource
    workflow: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class WorkflowAttributesDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class WorkflowAttributesGetRequest(_message.Message):
    __slots__ = ["domain", "project", "resource_type", "workflow"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    domain: str
    project: str
    resource_type: _matchable_resource_pb2.MatchableResource
    workflow: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class WorkflowAttributesGetResponse(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: WorkflowAttributes
    def __init__(self, attributes: _Optional[_Union[WorkflowAttributes, _Mapping]] = ...) -> None: ...

class WorkflowAttributesUpdateRequest(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: WorkflowAttributes
    def __init__(self, attributes: _Optional[_Union[WorkflowAttributes, _Mapping]] = ...) -> None: ...

class WorkflowAttributesUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
