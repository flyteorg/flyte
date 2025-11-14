from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Sort(_message.Message):
    __slots__ = ["key", "direction"]
    class Direction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        DESCENDING: _ClassVar[Sort.Direction]
        ASCENDING: _ClassVar[Sort.Direction]
    DESCENDING: Sort.Direction
    ASCENDING: Sort.Direction
    KEY_FIELD_NUMBER: _ClassVar[int]
    DIRECTION_FIELD_NUMBER: _ClassVar[int]
    key: str
    direction: Sort.Direction
    def __init__(self, key: _Optional[str] = ..., direction: _Optional[_Union[Sort.Direction, str]] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["limit", "token", "sort_by", "filters", "raw_filters", "sort_by_fields"]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    RAW_FILTERS_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELDS_FIELD_NUMBER: _ClassVar[int]
    limit: int
    token: str
    sort_by: Sort
    filters: _containers.RepeatedCompositeFieldContainer[Filter]
    raw_filters: _containers.RepeatedScalarFieldContainer[str]
    sort_by_fields: _containers.RepeatedCompositeFieldContainer[Sort]
    def __init__(self, limit: _Optional[int] = ..., token: _Optional[str] = ..., sort_by: _Optional[_Union[Sort, _Mapping]] = ..., filters: _Optional[_Iterable[_Union[Filter, _Mapping]]] = ..., raw_filters: _Optional[_Iterable[str]] = ..., sort_by_fields: _Optional[_Iterable[_Union[Sort, _Mapping]]] = ...) -> None: ...

class Filter(_message.Message):
    __slots__ = ["function", "field", "values"]
    class Function(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        EQUAL: _ClassVar[Filter.Function]
        NOT_EQUAL: _ClassVar[Filter.Function]
        GREATER_THAN: _ClassVar[Filter.Function]
        GREATER_THAN_OR_EQUAL: _ClassVar[Filter.Function]
        LESS_THAN: _ClassVar[Filter.Function]
        LESS_THAN_OR_EQUAL: _ClassVar[Filter.Function]
        CONTAINS: _ClassVar[Filter.Function]
        VALUE_IN: _ClassVar[Filter.Function]
        ENDS_WITH: _ClassVar[Filter.Function]
        NOT_ENDS_WITH: _ClassVar[Filter.Function]
        CONTAINS_CASE_INSENSITIVE: _ClassVar[Filter.Function]
    EQUAL: Filter.Function
    NOT_EQUAL: Filter.Function
    GREATER_THAN: Filter.Function
    GREATER_THAN_OR_EQUAL: Filter.Function
    LESS_THAN: Filter.Function
    LESS_THAN_OR_EQUAL: Filter.Function
    CONTAINS: Filter.Function
    VALUE_IN: Filter.Function
    ENDS_WITH: Filter.Function
    NOT_ENDS_WITH: Filter.Function
    CONTAINS_CASE_INSENSITIVE: Filter.Function
    FUNCTION_FIELD_NUMBER: _ClassVar[int]
    FIELD_FIELD_NUMBER: _ClassVar[int]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    function: Filter.Function
    field: str
    values: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, function: _Optional[_Union[Filter.Function, str]] = ..., field: _Optional[str] = ..., values: _Optional[_Iterable[str]] = ...) -> None: ...
