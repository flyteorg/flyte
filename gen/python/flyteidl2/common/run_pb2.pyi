from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class OffloadedInputData(_message.Message):
    __slots__ = ["uri", "cache_key"]
    URI_FIELD_NUMBER: _ClassVar[int]
    CACHE_KEY_FIELD_NUMBER: _ClassVar[int]
    uri: str
    cache_key: str
    def __init__(self, uri: _Optional[str] = ..., cache_key: _Optional[str] = ...) -> None: ...
