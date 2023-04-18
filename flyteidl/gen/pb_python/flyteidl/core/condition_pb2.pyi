from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ComparisonExpression(_message.Message):
    __slots__ = ["operator", "left_value", "right_value"]
    class Operator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        EQ: _ClassVar[ComparisonExpression.Operator]
        NEQ: _ClassVar[ComparisonExpression.Operator]
        GT: _ClassVar[ComparisonExpression.Operator]
        GTE: _ClassVar[ComparisonExpression.Operator]
        LT: _ClassVar[ComparisonExpression.Operator]
        LTE: _ClassVar[ComparisonExpression.Operator]
    EQ: ComparisonExpression.Operator
    NEQ: ComparisonExpression.Operator
    GT: ComparisonExpression.Operator
    GTE: ComparisonExpression.Operator
    LT: ComparisonExpression.Operator
    LTE: ComparisonExpression.Operator
    OPERATOR_FIELD_NUMBER: _ClassVar[int]
    LEFT_VALUE_FIELD_NUMBER: _ClassVar[int]
    RIGHT_VALUE_FIELD_NUMBER: _ClassVar[int]
    operator: ComparisonExpression.Operator
    left_value: Operand
    right_value: Operand
    def __init__(self, operator: _Optional[_Union[ComparisonExpression.Operator, str]] = ..., left_value: _Optional[_Union[Operand, _Mapping]] = ..., right_value: _Optional[_Union[Operand, _Mapping]] = ...) -> None: ...

class Operand(_message.Message):
    __slots__ = ["primitive", "var"]
    PRIMITIVE_FIELD_NUMBER: _ClassVar[int]
    VAR_FIELD_NUMBER: _ClassVar[int]
    primitive: _literals_pb2.Primitive
    var: str
    def __init__(self, primitive: _Optional[_Union[_literals_pb2.Primitive, _Mapping]] = ..., var: _Optional[str] = ...) -> None: ...

class BooleanExpression(_message.Message):
    __slots__ = ["conjunction", "comparison"]
    CONJUNCTION_FIELD_NUMBER: _ClassVar[int]
    COMPARISON_FIELD_NUMBER: _ClassVar[int]
    conjunction: ConjunctionExpression
    comparison: ComparisonExpression
    def __init__(self, conjunction: _Optional[_Union[ConjunctionExpression, _Mapping]] = ..., comparison: _Optional[_Union[ComparisonExpression, _Mapping]] = ...) -> None: ...

class ConjunctionExpression(_message.Message):
    __slots__ = ["operator", "left_expression", "right_expression"]
    class LogicalOperator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        AND: _ClassVar[ConjunctionExpression.LogicalOperator]
        OR: _ClassVar[ConjunctionExpression.LogicalOperator]
    AND: ConjunctionExpression.LogicalOperator
    OR: ConjunctionExpression.LogicalOperator
    OPERATOR_FIELD_NUMBER: _ClassVar[int]
    LEFT_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    RIGHT_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    operator: ConjunctionExpression.LogicalOperator
    left_expression: BooleanExpression
    right_expression: BooleanExpression
    def __init__(self, operator: _Optional[_Union[ConjunctionExpression.LogicalOperator, str]] = ..., left_expression: _Optional[_Union[BooleanExpression, _Mapping]] = ..., right_expression: _Optional[_Union[BooleanExpression, _Mapping]] = ...) -> None: ...
