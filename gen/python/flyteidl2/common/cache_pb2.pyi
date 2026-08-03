from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CacheMetadata(_message.Message):
    __slots__ = ["source_action_attempt"]
    SOURCE_ACTION_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    source_action_attempt: _identifier_pb2.ActionAttemptIdentifier
    def __init__(self, source_action_attempt: _Optional[_Union[_identifier_pb2.ActionAttemptIdentifier, _Mapping]] = ...) -> None: ...
