from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowAttributes(_message.Message):
    __slots__ = ["project", "domain", "workflow", "matching_attributes", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    MATCHING_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    workflow: str
    matching_attributes: _matchable_resource_pb2.MatchingAttributes
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., matching_attributes: _Optional[_Union[_matchable_resource_pb2.MatchingAttributes, _Mapping]] = ..., org: _Optional[str] = ...) -> None: ...

class WorkflowAttributesUpdateRequest(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: WorkflowAttributes
    def __init__(self, attributes: _Optional[_Union[WorkflowAttributes, _Mapping]] = ...) -> None: ...

class WorkflowAttributesUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class WorkflowAttributesGetRequest(_message.Message):
    __slots__ = ["project", "domain", "workflow", "resource_type", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    workflow: str
    resource_type: _matchable_resource_pb2.MatchableResource
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ..., org: _Optional[str] = ...) -> None: ...

class WorkflowAttributesGetResponse(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: WorkflowAttributes
    def __init__(self, attributes: _Optional[_Union[WorkflowAttributes, _Mapping]] = ...) -> None: ...

class WorkflowAttributesDeleteRequest(_message.Message):
    __slots__ = ["project", "domain", "workflow", "resource_type", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    workflow: str
    resource_type: _matchable_resource_pb2.MatchableResource
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ..., org: _Optional[str] = ...) -> None: ...

class WorkflowAttributesDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
