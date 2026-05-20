from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class TestProto(_message.Message):
    __slots__ = ["string_value"]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    string_value: str
    def __init__(self, string_value: _Optional[str] = ...) -> None: ...
