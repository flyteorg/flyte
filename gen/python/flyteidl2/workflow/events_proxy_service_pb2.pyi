from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecordRequest(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.ActionEvent]
    def __init__(self, events: _Optional[_Iterable[_Union[_run_definition_pb2.ActionEvent, _Mapping]]] = ...) -> None: ...

class RecordResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
