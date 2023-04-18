from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ProjectDomainAttributes(_message.Message):
    __slots__ = ["project", "domain", "matching_attributes"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    MATCHING_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    matching_attributes: _matchable_resource_pb2.MatchingAttributes
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., matching_attributes: _Optional[_Union[_matchable_resource_pb2.MatchingAttributes, _Mapping]] = ...) -> None: ...

class ProjectDomainAttributesUpdateRequest(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: ProjectDomainAttributes
    def __init__(self, attributes: _Optional[_Union[ProjectDomainAttributes, _Mapping]] = ...) -> None: ...

class ProjectDomainAttributesUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ProjectDomainAttributesGetRequest(_message.Message):
    __slots__ = ["project", "domain", "resource_type"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    resource_type: _matchable_resource_pb2.MatchableResource
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class ProjectDomainAttributesGetResponse(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: ProjectDomainAttributes
    def __init__(self, attributes: _Optional[_Union[ProjectDomainAttributes, _Mapping]] = ...) -> None: ...

class ProjectDomainAttributesDeleteRequest(_message.Message):
    __slots__ = ["project", "domain", "resource_type"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    resource_type: _matchable_resource_pb2.MatchableResource
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class ProjectDomainAttributesDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
