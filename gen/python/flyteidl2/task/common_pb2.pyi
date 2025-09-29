from flyteidl2.core import interface_pb2 as _interface_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FixedRateUnit(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    FIXED_RATE_UNIT_UNSPECIFIED: _ClassVar[FixedRateUnit]
    FIXED_RATE_UNIT_MINUTE: _ClassVar[FixedRateUnit]
    FIXED_RATE_UNIT_HOUR: _ClassVar[FixedRateUnit]
    FIXED_RATE_UNIT_DAY: _ClassVar[FixedRateUnit]
FIXED_RATE_UNIT_UNSPECIFIED: FixedRateUnit
FIXED_RATE_UNIT_MINUTE: FixedRateUnit
FIXED_RATE_UNIT_HOUR: FixedRateUnit
FIXED_RATE_UNIT_DAY: FixedRateUnit

class NamedParameter(_message.Message):
    __slots__ = ["name", "parameter"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PARAMETER_FIELD_NUMBER: _ClassVar[int]
    name: str
    parameter: _interface_pb2.Parameter
    def __init__(self, name: _Optional[str] = ..., parameter: _Optional[_Union[_interface_pb2.Parameter, _Mapping]] = ...) -> None: ...

class FixedRate(_message.Message):
    __slots__ = ["value", "unit", "start_time"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    UNIT_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    value: int
    unit: FixedRateUnit
    start_time: _timestamp_pb2.Timestamp
    def __init__(self, value: _Optional[int] = ..., unit: _Optional[_Union[FixedRateUnit, str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class Schedule(_message.Message):
    __slots__ = ["rate", "cron_expression", "kickoff_time_input_arg"]
    RATE_FIELD_NUMBER: _ClassVar[int]
    CRON_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    KICKOFF_TIME_INPUT_ARG_FIELD_NUMBER: _ClassVar[int]
    rate: FixedRate
    cron_expression: str
    kickoff_time_input_arg: str
    def __init__(self, rate: _Optional[_Union[FixedRate, _Mapping]] = ..., cron_expression: _Optional[str] = ..., kickoff_time_input_arg: _Optional[str] = ...) -> None: ...

class TriggerAutomationSpec(_message.Message):
    __slots__ = ["type", "schedule"]
    class Type(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        TYPE_UNSPECIFIED: _ClassVar[TriggerAutomationSpec.Type]
        TYPE_NONE: _ClassVar[TriggerAutomationSpec.Type]
        TYPE_SCHEDULE: _ClassVar[TriggerAutomationSpec.Type]
    TYPE_UNSPECIFIED: TriggerAutomationSpec.Type
    TYPE_NONE: TriggerAutomationSpec.Type
    TYPE_SCHEDULE: TriggerAutomationSpec.Type
    TYPE_FIELD_NUMBER: _ClassVar[int]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    type: TriggerAutomationSpec.Type
    schedule: Schedule
    def __init__(self, type: _Optional[_Union[TriggerAutomationSpec.Type, str]] = ..., schedule: _Optional[_Union[Schedule, _Mapping]] = ...) -> None: ...
