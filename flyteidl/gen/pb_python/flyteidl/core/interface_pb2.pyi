from flyteidl.core import types_pb2 as _types_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Variable(_message.Message):
    __slots__ = ["type", "description"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    type: _types_pb2.LiteralType
    description: str
    def __init__(self, type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., description: _Optional[str] = ...) -> None: ...

class VariableMap(_message.Message):
    __slots__ = ["variables"]
    class VariablesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Variable
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Variable, _Mapping]] = ...) -> None: ...
    VARIABLES_FIELD_NUMBER: _ClassVar[int]
    variables: _containers.MessageMap[str, Variable]
    def __init__(self, variables: _Optional[_Mapping[str, Variable]] = ...) -> None: ...

class TypedInterface(_message.Message):
    __slots__ = ["inputs", "outputs"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    inputs: VariableMap
    outputs: VariableMap
    def __init__(self, inputs: _Optional[_Union[VariableMap, _Mapping]] = ..., outputs: _Optional[_Union[VariableMap, _Mapping]] = ...) -> None: ...

class Parameter(_message.Message):
    __slots__ = ["var", "default", "required"]
    VAR_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_FIELD_NUMBER: _ClassVar[int]
    REQUIRED_FIELD_NUMBER: _ClassVar[int]
    var: Variable
    default: _literals_pb2.Literal
    required: bool
    def __init__(self, var: _Optional[_Union[Variable, _Mapping]] = ..., default: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ..., required: bool = ...) -> None: ...

class ParameterMap(_message.Message):
    __slots__ = ["parameters"]
    class ParametersEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Parameter
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Parameter, _Mapping]] = ...) -> None: ...
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    parameters: _containers.MessageMap[str, Parameter]
    def __init__(self, parameters: _Optional[_Mapping[str, Parameter]] = ...) -> None: ...
