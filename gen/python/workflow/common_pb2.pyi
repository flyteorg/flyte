from core import interface_pb2 as _interface_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class NamedParameter(_message.Message):
    __slots__ = ["name", "parameter"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PARAMETER_FIELD_NUMBER: _ClassVar[int]
    name: str
    parameter: _interface_pb2.Parameter
    def __init__(self, name: _Optional[str] = ..., parameter: _Optional[_Union[_interface_pb2.Parameter, _Mapping]] = ...) -> None: ...
