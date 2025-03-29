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

class ConcurrencyPolicy(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNSPECIFIED: _ClassVar[ConcurrencyPolicy]
    WAIT: _ClassVar[ConcurrencyPolicy]
    ABORT: _ClassVar[ConcurrencyPolicy]
    REPLACE: _ClassVar[ConcurrencyPolicy]

class ConcurrencyLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    LAUNCH_PLAN: _ClassVar[ConcurrencyLevel]
    LAUNCH_PLAN_VERSION: _ClassVar[ConcurrencyLevel]
MINUTE: FixedRateUnit
HOUR: FixedRateUnit
DAY: FixedRateUnit
UNSPECIFIED: ConcurrencyPolicy
WAIT: ConcurrencyPolicy
ABORT: ConcurrencyPolicy
REPLACE: ConcurrencyPolicy
LAUNCH_PLAN: ConcurrencyLevel
LAUNCH_PLAN_VERSION: ConcurrencyLevel

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
    __slots__ = ["cron_expression", "rate", "cron_schedule", "kickoff_time_input_arg", "scheduler_policy"]
    CRON_EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    RATE_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    KICKOFF_TIME_INPUT_ARG_FIELD_NUMBER: _ClassVar[int]
    SCHEDULER_POLICY_FIELD_NUMBER: _ClassVar[int]
    cron_expression: str
    rate: FixedRate
    cron_schedule: CronSchedule
    kickoff_time_input_arg: str
    scheduler_policy: SchedulerPolicy
    def __init__(self, cron_expression: _Optional[str] = ..., rate: _Optional[_Union[FixedRate, _Mapping]] = ..., cron_schedule: _Optional[_Union[CronSchedule, _Mapping]] = ..., kickoff_time_input_arg: _Optional[str] = ..., scheduler_policy: _Optional[_Union[SchedulerPolicy, _Mapping]] = ...) -> None: ...

class SchedulerPolicy(_message.Message):
    __slots__ = ["max", "policy", "level"]
    MAX_FIELD_NUMBER: _ClassVar[int]
    POLICY_FIELD_NUMBER: _ClassVar[int]
    LEVEL_FIELD_NUMBER: _ClassVar[int]
    max: int
    policy: ConcurrencyPolicy
    level: ConcurrencyLevel
    def __init__(self, max: _Optional[int] = ..., policy: _Optional[_Union[ConcurrencyPolicy, str]] = ..., level: _Optional[_Union[ConcurrencyLevel, str]] = ...) -> None: ...
