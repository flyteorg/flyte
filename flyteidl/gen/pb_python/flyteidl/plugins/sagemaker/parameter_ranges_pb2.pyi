from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HyperparameterScalingType(_message.Message):
    __slots__ = []
    class Value(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        AUTO: _ClassVar[HyperparameterScalingType.Value]
        LINEAR: _ClassVar[HyperparameterScalingType.Value]
        LOGARITHMIC: _ClassVar[HyperparameterScalingType.Value]
        REVERSELOGARITHMIC: _ClassVar[HyperparameterScalingType.Value]
    AUTO: HyperparameterScalingType.Value
    LINEAR: HyperparameterScalingType.Value
    LOGARITHMIC: HyperparameterScalingType.Value
    REVERSELOGARITHMIC: HyperparameterScalingType.Value
    def __init__(self) -> None: ...

class ContinuousParameterRange(_message.Message):
    __slots__ = ["max_value", "min_value", "scaling_type"]
    MAX_VALUE_FIELD_NUMBER: _ClassVar[int]
    MIN_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCALING_TYPE_FIELD_NUMBER: _ClassVar[int]
    max_value: float
    min_value: float
    scaling_type: HyperparameterScalingType.Value
    def __init__(self, max_value: _Optional[float] = ..., min_value: _Optional[float] = ..., scaling_type: _Optional[_Union[HyperparameterScalingType.Value, str]] = ...) -> None: ...

class IntegerParameterRange(_message.Message):
    __slots__ = ["max_value", "min_value", "scaling_type"]
    MAX_VALUE_FIELD_NUMBER: _ClassVar[int]
    MIN_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCALING_TYPE_FIELD_NUMBER: _ClassVar[int]
    max_value: int
    min_value: int
    scaling_type: HyperparameterScalingType.Value
    def __init__(self, max_value: _Optional[int] = ..., min_value: _Optional[int] = ..., scaling_type: _Optional[_Union[HyperparameterScalingType.Value, str]] = ...) -> None: ...

class CategoricalParameterRange(_message.Message):
    __slots__ = ["values"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, values: _Optional[_Iterable[str]] = ...) -> None: ...

class ParameterRangeOneOf(_message.Message):
    __slots__ = ["continuous_parameter_range", "integer_parameter_range", "categorical_parameter_range"]
    CONTINUOUS_PARAMETER_RANGE_FIELD_NUMBER: _ClassVar[int]
    INTEGER_PARAMETER_RANGE_FIELD_NUMBER: _ClassVar[int]
    CATEGORICAL_PARAMETER_RANGE_FIELD_NUMBER: _ClassVar[int]
    continuous_parameter_range: ContinuousParameterRange
    integer_parameter_range: IntegerParameterRange
    categorical_parameter_range: CategoricalParameterRange
    def __init__(self, continuous_parameter_range: _Optional[_Union[ContinuousParameterRange, _Mapping]] = ..., integer_parameter_range: _Optional[_Union[IntegerParameterRange, _Mapping]] = ..., categorical_parameter_range: _Optional[_Union[CategoricalParameterRange, _Mapping]] = ...) -> None: ...

class ParameterRanges(_message.Message):
    __slots__ = ["parameter_range_map"]
    class ParameterRangeMapEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: ParameterRangeOneOf
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[ParameterRangeOneOf, _Mapping]] = ...) -> None: ...
    PARAMETER_RANGE_MAP_FIELD_NUMBER: _ClassVar[int]
    parameter_range_map: _containers.MessageMap[str, ParameterRangeOneOf]
    def __init__(self, parameter_range_map: _Optional[_Mapping[str, ParameterRangeOneOf]] = ...) -> None: ...
