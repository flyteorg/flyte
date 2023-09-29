from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import types_pb2 as _types_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SignalGetOrCreateRequest(_message.Message):
    __slots__ = ["id", "type"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.SignalIdentifier
    type: _types_pb2.LiteralType
    def __init__(self, id: _Optional[_Union[_identifier_pb2.SignalIdentifier, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ...) -> None: ...

class SignalListRequest(_message.Message):
    __slots__ = ["workflow_execution_id", "limit", "token", "filters", "sort_by"]
    WORKFLOW_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    workflow_execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    limit: int
    token: str
    filters: str
    sort_by: _common_pb2.Sort
    def __init__(self, workflow_execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...

class SignalList(_message.Message):
    __slots__ = ["signals", "token"]
    SIGNALS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    signals: _containers.RepeatedCompositeFieldContainer[Signal]
    token: str
    def __init__(self, signals: _Optional[_Iterable[_Union[Signal, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class SignalSetRequest(_message.Message):
    __slots__ = ["id", "value"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.SignalIdentifier
    value: _literals_pb2.Literal
    def __init__(self, id: _Optional[_Union[_identifier_pb2.SignalIdentifier, _Mapping]] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...

class SignalSetResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Signal(_message.Message):
    __slots__ = ["id", "type", "value"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.SignalIdentifier
    type: _types_pb2.LiteralType
    value: _literals_pb2.Literal
    def __init__(self, id: _Optional[_Union[_identifier_pb2.SignalIdentifier, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...
