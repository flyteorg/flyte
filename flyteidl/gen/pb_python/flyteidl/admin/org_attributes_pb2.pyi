from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class OrgAttributes(_message.Message):
    __slots__ = ["matching_attributes", "org"]
    MATCHING_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    matching_attributes: _matchable_resource_pb2.MatchingAttributes
    org: str
    def __init__(self, matching_attributes: _Optional[_Union[_matchable_resource_pb2.MatchingAttributes, _Mapping]] = ..., org: _Optional[str] = ...) -> None: ...

class OrgAttributesUpdateRequest(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: OrgAttributes
    def __init__(self, attributes: _Optional[_Union[OrgAttributes, _Mapping]] = ...) -> None: ...

class OrgAttributesUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class OrgAttributesGetRequest(_message.Message):
    __slots__ = ["resource_type", "org"]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    resource_type: _matchable_resource_pb2.MatchableResource
    org: str
    def __init__(self, resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ..., org: _Optional[str] = ...) -> None: ...

class OrgAttributesGetResponse(_message.Message):
    __slots__ = ["attributes"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: OrgAttributes
    def __init__(self, attributes: _Optional[_Union[OrgAttributes, _Mapping]] = ...) -> None: ...

class OrgAttributesDeleteRequest(_message.Message):
    __slots__ = ["resource_type", "org"]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    resource_type: _matchable_resource_pb2.MatchableResource
    org: str
    def __init__(self, resource_type: _Optional[_Union[_matchable_resource_pb2.MatchableResource, str]] = ..., org: _Optional[str] = ...) -> None: ...

class OrgAttributesDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
