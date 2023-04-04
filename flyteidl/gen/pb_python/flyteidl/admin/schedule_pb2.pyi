from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DAY: FixedRateUnit
DESCRIPTOR: _descriptor.FileDescriptor
HOUR: FixedRateUnit
MINUTE: FixedRateUnit

class CronSchedule(_message.Message):
    __slots__ = ["offset", "schedule"]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    offset: str
    schedule: str
    def __init__(self, schedule: _Optional[str] = ..., offset: _Optional[str] = ...) -> None: ...

class FixedRate(_message.Message):
    __slots__ = ["unit", "value"]
    UNIT_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    unit: FixedRateUnit
    value: int
    def __init__(self, value: _Optional[int] = ..., unit: _Optional[_Union[FixedRateUnit, str]] = ...) -> None: ...

class Schedule(_message.Message):
    __slots__ = ["cron_expression", "cron_schedule", "kickoff_time_input_arg", "rate"]
    CRON_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    KICKOFF_TIME_INPUT_ARG_FIELD_NUMBER: _ClassVar[int]
    RATE_FIELD_NUMBER: _ClassVar[int]
    cron_expression: str
    cron_schedule: CronSchedule
    kickoff_time_input_arg: str
    rate: FixedRate
    def __init__(self, cron_expression: _Optional[str] = ..., rate: _Optional[_Union[FixedRate, _Mapping]] = ..., cron_schedule: _Optional[_Union[CronSchedule, _Mapping]] = ..., kickoff_time_input_arg: _Optional[str] = ...) -> None: ...

class FixedRateUnit(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
