from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class OffloadedInputData(_message.Message):
    __slots__ = ["uri", "inputs_hash"]
    URI_FIELD_NUMBER: _ClassVar[int]
    INPUTS_HASH_FIELD_NUMBER: _ClassVar[int]
    uri: str
    inputs_hash: str
    def __init__(self, uri: _Optional[str] = ..., inputs_hash: _Optional[str] = ...) -> None: ...
