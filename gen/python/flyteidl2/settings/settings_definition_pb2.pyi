from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SettingState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SETTING_STATE_INHERIT: _ClassVar[SettingState]
    SETTING_STATE_UNSET: _ClassVar[SettingState]
    SETTING_STATE_VALUE: _ClassVar[SettingState]

class ScopeLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SCOPE_LEVEL_ORG: _ClassVar[ScopeLevel]
    SCOPE_LEVEL_DOMAIN: _ClassVar[ScopeLevel]
    SCOPE_LEVEL_PROJECT: _ClassVar[ScopeLevel]
SETTING_STATE_INHERIT: SettingState
SETTING_STATE_UNSET: SettingState
SETTING_STATE_VALUE: SettingState
SCOPE_LEVEL_ORG: ScopeLevel
SCOPE_LEVEL_DOMAIN: ScopeLevel
SCOPE_LEVEL_PROJECT: ScopeLevel
DESC_FIELD_NUMBER: _ClassVar[int]
desc: _descriptor.FieldDescriptor

class StringValues(_message.Message):
    __slots__ = ["values"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, values: _Optional[_Iterable[str]] = ...) -> None: ...

class StringMap(_message.Message):
    __slots__ = ["entries"]
    class EntriesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ENTRIES_FIELD_NUMBER: _ClassVar[int]
    entries: _containers.ScalarMap[str, str]
    def __init__(self, entries: _Optional[_Mapping[str, str]] = ...) -> None: ...

class SettingsKey(_message.Message):
    __slots__ = ["org", "domain", "project"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    org: str
    domain: str
    project: str
    def __init__(self, org: _Optional[str] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ...) -> None: ...

class StringSetting(_message.Message):
    __slots__ = ["state", "string_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    string_value: str
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., string_value: _Optional[str] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class Int64Setting(_message.Message):
    __slots__ = ["state", "int_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    INT_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    int_value: int
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., int_value: _Optional[int] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class BoolSetting(_message.Message):
    __slots__ = ["state", "bool_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    BOOL_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    bool_value: bool
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., bool_value: bool = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class StringListSetting(_message.Message):
    __slots__ = ["state", "list_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    LIST_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    list_value: StringValues
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., list_value: _Optional[_Union[StringValues, _Mapping]] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class StringMapSetting(_message.Message):
    __slots__ = ["state", "map_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    MAP_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    map_value: StringMap
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., map_value: _Optional[_Union[StringMap, _Mapping]] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class QuantitySetting(_message.Message):
    __slots__ = ["state", "quantity_value", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_VALUE_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    quantity_value: str
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., quantity_value: _Optional[str] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class RunSettings(_message.Message):
    __slots__ = ["default_queue", "run_concurrency", "action_concurrency"]
    DEFAULT_QUEUE_FIELD_NUMBER: _ClassVar[int]
    RUN_CONCURRENCY_FIELD_NUMBER: _ClassVar[int]
    ACTION_CONCURRENCY_FIELD_NUMBER: _ClassVar[int]
    default_queue: StringSetting
    run_concurrency: Int64Setting
    action_concurrency: Int64Setting
    def __init__(self, default_queue: _Optional[_Union[StringSetting, _Mapping]] = ..., run_concurrency: _Optional[_Union[Int64Setting, _Mapping]] = ..., action_concurrency: _Optional[_Union[Int64Setting, _Mapping]] = ...) -> None: ...

class SecuritySettings(_message.Message):
    __slots__ = ["service_account"]
    SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    service_account: StringSetting
    def __init__(self, service_account: _Optional[_Union[StringSetting, _Mapping]] = ...) -> None: ...

class StorageSettings(_message.Message):
    __slots__ = ["raw_data_path"]
    RAW_DATA_PATH_FIELD_NUMBER: _ClassVar[int]
    raw_data_path: StringSetting
    def __init__(self, raw_data_path: _Optional[_Union[StringSetting, _Mapping]] = ...) -> None: ...

class TaskResourceDefaults(_message.Message):
    __slots__ = ["cpu", "gpu", "memory", "storage"]
    CPU_FIELD_NUMBER: _ClassVar[int]
    GPU_FIELD_NUMBER: _ClassVar[int]
    MEMORY_FIELD_NUMBER: _ClassVar[int]
    STORAGE_FIELD_NUMBER: _ClassVar[int]
    cpu: QuantitySetting
    gpu: QuantitySetting
    memory: QuantitySetting
    storage: QuantitySetting
    def __init__(self, cpu: _Optional[_Union[QuantitySetting, _Mapping]] = ..., gpu: _Optional[_Union[QuantitySetting, _Mapping]] = ..., memory: _Optional[_Union[QuantitySetting, _Mapping]] = ..., storage: _Optional[_Union[QuantitySetting, _Mapping]] = ...) -> None: ...

class TaskResourceSettings(_message.Message):
    __slots__ = ["min", "max", "mirror_limits_request"]
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    MIRROR_LIMITS_REQUEST_FIELD_NUMBER: _ClassVar[int]
    min: TaskResourceDefaults
    max: TaskResourceDefaults
    mirror_limits_request: BoolSetting
    def __init__(self, min: _Optional[_Union[TaskResourceDefaults, _Mapping]] = ..., max: _Optional[_Union[TaskResourceDefaults, _Mapping]] = ..., mirror_limits_request: _Optional[_Union[BoolSetting, _Mapping]] = ...) -> None: ...

class Settings(_message.Message):
    __slots__ = ["run", "security", "storage", "task_resource", "labels", "annotations", "environment_variables"]
    RUN_FIELD_NUMBER: _ClassVar[int]
    SECURITY_FIELD_NUMBER: _ClassVar[int]
    STORAGE_FIELD_NUMBER: _ClassVar[int]
    TASK_RESOURCE_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_VARIABLES_FIELD_NUMBER: _ClassVar[int]
    run: RunSettings
    security: SecuritySettings
    storage: StorageSettings
    task_resource: TaskResourceSettings
    labels: StringListSetting
    annotations: StringMapSetting
    environment_variables: StringMapSetting
    def __init__(self, run: _Optional[_Union[RunSettings, _Mapping]] = ..., security: _Optional[_Union[SecuritySettings, _Mapping]] = ..., storage: _Optional[_Union[StorageSettings, _Mapping]] = ..., task_resource: _Optional[_Union[TaskResourceSettings, _Mapping]] = ..., labels: _Optional[_Union[StringListSetting, _Mapping]] = ..., annotations: _Optional[_Union[StringMapSetting, _Mapping]] = ..., environment_variables: _Optional[_Union[StringMapSetting, _Mapping]] = ...) -> None: ...
