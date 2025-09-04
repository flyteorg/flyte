from google.protobuf import struct_pb2 as _struct_pb2
from flyteidl2.core import interface_pb2 as _interface_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LiteralsToJsonRequest(_message.Message):
    __slots__ = ["literals", "variables"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    VARIABLES_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.NamedLiteral]
    variables: _interface_pb2.VariableMap
    def __init__(self, literals: _Optional[_Iterable[_Union[_run_definition_pb2.NamedLiteral, _Mapping]]] = ..., variables: _Optional[_Union[_interface_pb2.VariableMap, _Mapping]] = ...) -> None: ...

class LiteralsToJsonResponse(_message.Message):
    __slots__ = ["json"]
    JSON_FIELD_NUMBER: _ClassVar[int]
    json: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    def __init__(self, json: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ...) -> None: ...

class JsonToLiteralsRequest(_message.Message):
    __slots__ = ["json"]
    JSON_FIELD_NUMBER: _ClassVar[int]
    json: _containers.RepeatedCompositeFieldContainer[_struct_pb2.Struct]
    def __init__(self, json: _Optional[_Iterable[_Union[_struct_pb2.Struct, _Mapping]]] = ...) -> None: ...

class JsonToLiteralsResponse(_message.Message):
    __slots__ = ["literals"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.NamedLiteral]
    def __init__(self, literals: _Optional[_Iterable[_Union[_run_definition_pb2.NamedLiteral, _Mapping]]] = ...) -> None: ...
