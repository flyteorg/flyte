from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.logs.dataplane import payload_pb2 as _payload_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TailLogsRequest(_message.Message):
    __slots__ = ["action_id", "attempt"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class TailLogsResponse(_message.Message):
    __slots__ = ["logs"]
    class Logs(_message.Message):
        __slots__ = ["lines"]
        LINES_FIELD_NUMBER: _ClassVar[int]
        lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
        def __init__(self, lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ...) -> None: ...
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[TailLogsResponse.Logs]
    def __init__(self, logs: _Optional[_Iterable[_Union[TailLogsResponse.Logs, _Mapping]]] = ...) -> None: ...
