from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class BooleanExpression(_message.Message):
    __slots__ = ["comparison", "conjunction"]
    COMPARISON_FIELD_NUMBER: _ClassVar[int]
    CONJUNCTION_FIELD_NUMBER: _ClassVar[int]
    comparison: ComparisonExpression
    conjunction: ConjunctionExpression
    def __init__(self, conjunction: _Optional[_Union[ConjunctionExpression, _Mapping]] = ..., comparison: _Optional[_Union[ComparisonExpression, _Mapping]] = ...) -> None: ...

class ComparisonExpression(_message.Message):
    __slots__ = ["left_value", "operator", "right_value"]
    class Operator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    EQ: ComparisonExpression.Operator
    GT: ComparisonExpression.Operator
    GTE: ComparisonExpression.Operator
    LEFT_VALUE_FIELD_NUMBER: _ClassVar[int]
    LT: ComparisonExpression.Operator
    LTE: ComparisonExpression.Operator
    NEQ: ComparisonExpression.Operator
    OPERATOR_FIELD_NUMBER: _ClassVar[int]
    RIGHT_VALUE_FIELD_NUMBER: _ClassVar[int]
    left_value: Operand
    operator: ComparisonExpression.Operator
    right_value: Operand
    def __init__(self, operator: _Optional[_Union[ComparisonExpression.Operator, str]] = ..., left_value: _Optional[_Union[Operand, _Mapping]] = ..., right_value: _Optional[_Union[Operand, _Mapping]] = ...) -> None: ...

class ConjunctionExpression(_message.Message):
    __slots__ = ["left_expression", "operator", "right_expression"]
    class LogicalOperator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    AND: ConjunctionExpression.LogicalOperator
    LEFT_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    OPERATOR_FIELD_NUMBER: _ClassVar[int]
    OR: ConjunctionExpression.LogicalOperator
    RIGHT_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    left_expression: BooleanExpression
    operator: ConjunctionExpression.LogicalOperator
    right_expression: BooleanExpression
    def __init__(self, operator: _Optional[_Union[ConjunctionExpression.LogicalOperator, str]] = ..., left_expression: _Optional[_Union[BooleanExpression, _Mapping]] = ..., right_expression: _Optional[_Union[BooleanExpression, _Mapping]] = ...) -> None: ...

class Operand(_message.Message):
    __slots__ = ["primitive", "var"]
    PRIMITIVE_FIELD_NUMBER: _ClassVar[int]
    VAR_FIELD_NUMBER: _ClassVar[int]
    primitive: _literals_pb2.Primitive
    var: str
    def __init__(self, primitive: _Optional[_Union[_literals_pb2.Primitive, _Mapping]] = ..., var: _Optional[str] = ...) -> None: ...
