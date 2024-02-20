from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FixedRateUnit(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    MINUTE: _ClassVar[FixedRateUnit]
    HOUR: _ClassVar[FixedRateUnit]
    DAY: _ClassVar[FixedRateUnit]
MINUTE: FixedRateUnit
HOUR: FixedRateUnit
DAY: FixedRateUnit

class FixedRate(_message.Message):
    __slots__ = ["value", "unit"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    UNIT_FIELD_NUMBER: _ClassVar[int]
    value: int
    unit: FixedRateUnit
    def __init__(self, value: _Optional[int] = ..., unit: _Optional[_Union[FixedRateUnit, str]] = ...) -> None: ...

class CronSchedule(_message.Message):
    __slots__ = ["schedule", "offset"]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    schedule: str
    offset: str
    def __init__(self, schedule: _Optional[str] = ..., offset: _Optional[str] = ...) -> None: ...

class Schedule(_message.Message):
    __slots__ = ["cron_expression", "rate", "cron_schedule", "kickoff_time_input_arg"]
    CRON_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    RATE_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    KICKOFF_TIME_INPUT_ARG_FIELD_NUMBER: _ClassVar[int]
    cron_expression: str
    rate: FixedRate
    cron_schedule: CronSchedule
    kickoff_time_input_arg: str
    def __init__(self, cron_expression: _Optional[str] = ..., rate: _Optional[_Union[FixedRate, _Mapping]] = ..., cron_schedule: _Optional[_Union[CronSchedule, _Mapping]] = ..., kickoff_time_input_arg: _Optional[str] = ...) -> None: ...
