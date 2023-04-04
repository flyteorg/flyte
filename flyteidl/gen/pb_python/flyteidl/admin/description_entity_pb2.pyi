from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.admin import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTION_FORMAT_HTML: DescriptionFormat
DESCRIPTION_FORMAT_MARKDOWN: DescriptionFormat
DESCRIPTION_FORMAT_RST: DescriptionFormat
DESCRIPTION_FORMAT_UNKNOWN: DescriptionFormat
DESCRIPTOR: _descriptor.FileDescriptor

class Description(_message.Message):
    __slots__ = ["format", "icon_link", "uri", "value"]
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    ICON_LINK_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    format: DescriptionFormat
    icon_link: str
    uri: str
    value: str
    def __init__(self, value: _Optional[str] = ..., uri: _Optional[str] = ..., format: _Optional[_Union[DescriptionFormat, str]] = ..., icon_link: _Optional[str] = ...) -> None: ...

class DescriptionEntity(_message.Message):
    __slots__ = ["id", "long_description", "short_description", "source_code", "tags"]
    ID_FIELD_NUMBER: _ClassVar[int]
    LONG_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SHORT_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SOURCE_CODE_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    long_description: Description
    short_description: str
    source_code: SourceCode
    tags: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., short_description: _Optional[str] = ..., long_description: _Optional[_Union[Description, _Mapping]] = ..., source_code: _Optional[_Union[SourceCode, _Mapping]] = ..., tags: _Optional[_Iterable[str]] = ...) -> None: ...

class DescriptionEntityList(_message.Message):
    __slots__ = ["descriptionEntities", "token"]
    DESCRIPTIONENTITIES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    descriptionEntities: _containers.RepeatedCompositeFieldContainer[DescriptionEntity]
    token: str
    def __init__(self, descriptionEntities: _Optional[_Iterable[_Union[DescriptionEntity, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class DescriptionEntityListRequest(_message.Message):
    __slots__ = ["filters", "id", "limit", "resource_type", "sort_by", "token"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    filters: str
    id: _common_pb2.NamedEntityIdentifier
    limit: int
    resource_type: _identifier_pb2.ResourceType
    sort_by: _common_pb2.Sort
    token: str
    def __init__(self, resource_type: _Optional[_Union[_identifier_pb2.ResourceType, str]] = ..., id: _Optional[_Union[_common_pb2.NamedEntityIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...

class SourceCode(_message.Message):
    __slots__ = ["link"]
    LINK_FIELD_NUMBER: _ClassVar[int]
    link: str
    def __init__(self, link: _Optional[str] = ...) -> None: ...

class DescriptionFormat(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
