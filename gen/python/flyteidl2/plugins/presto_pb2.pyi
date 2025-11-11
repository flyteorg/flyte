from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PrestoQuery(_message.Message):
    __slots__ = ["routing_group", "catalog", "schema", "statement"]
    ROUTING_GROUP_FIELD_NUMBER: _ClassVar[int]
    CATALOG_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    STATEMENT_FIELD_NUMBER: _ClassVar[int]
    routing_group: str
    catalog: str
    schema: str
    statement: str
    def __init__(self, routing_group: _Optional[str] = ..., catalog: _Optional[str] = ..., schema: _Optional[str] = ..., statement: _Optional[str] = ...) -> None: ...
