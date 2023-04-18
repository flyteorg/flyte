from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ProjectAttributes(_message.Message):
    __slots__ = ["project", "matching_attributes"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    MATCHING_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    project: str
    matching_attributes: _matchable_resource_pb2.MatchingAttributes
    def __init__(self, project: _Optional[str] = ..., matching_attributes: _Optional[_Union[_matchable_resource_pb2.MatchingAttributes, _Mapping]] = ...) -> None: ...

class ProjectAttributesUpdateRequest(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: ProjectAttributes
    def __init__(self, attributes: _Optional[_Union[ProjectAttributes, _Mapping]] = ...) -> None: ...

class ProjectAttributesUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ProjectAttributesGetRequest(_message.Message):
    __slots__ = ["project", "resource_type"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    project: str
    resource_type: _matchable_resource_pb2.MatchableResource
    def __init__(self, project: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class ProjectAttributesGetResponse(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: ProjectAttributes
    def __init__(self, attributes: _Optional[_Union[ProjectAttributes, _Mapping]] = ...) -> None: ...

class ProjectAttributesDeleteRequest(_message.Message):
    __slots__ = ["project", "resource_type"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    project: str
    resource_type: _matchable_resource_pb2.MatchableResource
    def __init__(self, project: _Optional[str] = ..., resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ...) -> None: ...

class ProjectAttributesDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
