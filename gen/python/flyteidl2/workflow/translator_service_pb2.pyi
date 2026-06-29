from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.core import interface_pb2 as _interface_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LiteralsToLaunchFormJsonRequest(_message.Message):
    __slots__ = ["literals", "variables", "literals_uri", "action_id", "trigger_id"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    VARIABLES_FIELD_NUMBER: _ClassVar[int]
    LITERALS_URI_FIELD_NUMBER: _ClassVar[int]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[_common_pb2.NamedLiteral]
    variables: _interface_pb2.VariableMap
    literals_uri: str
    action_id: _identifier_pb2.ActionIdentifier
    trigger_id: _identifier_pb2.TriggerIdentifier
    def __init__(self, literals: _Optional[_Iterable[_Union[_common_pb2.NamedLiteral, _Mapping]]] = ..., variables: _Optional[_Union[_interface_pb2.VariableMap, _Mapping]] = ..., literals_uri: _Optional[str] = ..., action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., trigger_id: _Optional[_Union[_identifier_pb2.TriggerIdentifier, _Mapping]] = ...) -> None: ...

class LiteralsToLaunchFormJsonResponse(_message.Message):
    __slots__ = ["json"]
    JSON_FIELD_NUMBER: _ClassVar[int]
    json: _struct_pb2.Struct
    def __init__(self, json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class LaunchFormJsonToLiteralsRequest(_message.Message):
    __slots__ = ["json"]
    JSON_FIELD_NUMBER: _ClassVar[int]
    json: _struct_pb2.Struct
    def __init__(self, json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class LaunchFormJsonToLiteralsResponse(_message.Message):
    __slots__ = ["literals"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[_common_pb2.NamedLiteral]
    def __init__(self, literals: _Optional[_Iterable[_Union[_common_pb2.NamedLiteral, _Mapping]]] = ...) -> None: ...

class TaskSpecToLaunchFormJsonRequest(_message.Message):
    __slots__ = ["task_spec"]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    task_spec: _task_definition_pb2.TaskSpec
    def __init__(self, task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ...) -> None: ...

class TaskSpecToLaunchFormJsonResponse(_message.Message):
    __slots__ = ["json"]
    JSON_FIELD_NUMBER: _ClassVar[int]
    json: _struct_pb2.Struct
    def __init__(self, json: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class JsonValuesToLiteralsRequest(_message.Message):
    __slots__ = ["variables", "values"]
    VARIABLES_FIELD_NUMBER: _ClassVar[int]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    variables: _interface_pb2.VariableMap
    values: _struct_pb2.Struct
    def __init__(self, variables: _Optional[_Union[_interface_pb2.VariableMap, _Mapping]] = ..., values: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class JsonValuesToLiteralsResponse(_message.Message):
    __slots__ = ["literals"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[_common_pb2.NamedLiteral]
    def __init__(self, literals: _Optional[_Iterable[_Union[_common_pb2.NamedLiteral, _Mapping]]] = ...) -> None: ...
