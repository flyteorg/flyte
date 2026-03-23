from buf.validate import validate_pb2 as _validate_pb2
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

class ScopeLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SCOPE_LEVEL_ORG: _ClassVar[ScopeLevel]
    SCOPE_LEVEL_DOMAIN: _ClassVar[ScopeLevel]
    SCOPE_LEVEL_PROJECT: _ClassVar[ScopeLevel]
SETTING_STATE_INHERIT: SettingState
SETTING_STATE_UNSET: SettingState
SCOPE_LEVEL_ORG: ScopeLevel
SCOPE_LEVEL_DOMAIN: ScopeLevel
SCOPE_LEVEL_PROJECT: ScopeLevel

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

class SettingValue(_message.Message):
    __slots__ = ["state", "string_value", "int_value", "bool_value", "list_value", "map_value", "description", "scope_level"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    INT_VALUE_FIELD_NUMBER: _ClassVar[int]
    BOOL_VALUE_FIELD_NUMBER: _ClassVar[int]
    LIST_VALUE_FIELD_NUMBER: _ClassVar[int]
    MAP_VALUE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    SCOPE_LEVEL_FIELD_NUMBER: _ClassVar[int]
    state: SettingState
    string_value: str
    int_value: int
    bool_value: bool
    list_value: StringValues
    map_value: StringMap
    description: str
    scope_level: ScopeLevel
    def __init__(self, state: _Optional[_Union[SettingState, str]] = ..., string_value: _Optional[str] = ..., int_value: _Optional[int] = ..., bool_value: bool = ..., list_value: _Optional[_Union[StringValues, _Mapping]] = ..., map_value: _Optional[_Union[StringMap, _Mapping]] = ..., description: _Optional[str] = ..., scope_level: _Optional[_Union[ScopeLevel, str]] = ...) -> None: ...

class SettingsKey(_message.Message):
    __slots__ = ["org", "domain", "project"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    org: str
    domain: str
    project: str
    def __init__(self, org: _Optional[str] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ...) -> None: ...

class RunSettings(_message.Message):
    __slots__ = ["default_queue", "run_concurrency", "action_concurrency"]
    DEFAULT_QUEUE_FIELD_NUMBER: _ClassVar[int]
    RUN_CONCURRENCY_FIELD_NUMBER: _ClassVar[int]
    ACTION_CONCURRENCY_FIELD_NUMBER: _ClassVar[int]
    default_queue: SettingValue
    run_concurrency: SettingValue
    action_concurrency: SettingValue
    def __init__(self, default_queue: _Optional[_Union[SettingValue, _Mapping]] = ..., run_concurrency: _Optional[_Union[SettingValue, _Mapping]] = ..., action_concurrency: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...

class SecuritySettings(_message.Message):
    __slots__ = ["service_account"]
    SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    service_account: SettingValue
    def __init__(self, service_account: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...

class StorageSettings(_message.Message):
    __slots__ = ["raw_data_path"]
    RAW_DATA_PATH_FIELD_NUMBER: _ClassVar[int]
    raw_data_path: SettingValue
    def __init__(self, raw_data_path: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...

class TaskResourceDefaults(_message.Message):
    __slots__ = ["min_cpu", "min_gpu", "min_memory", "min_storage", "max_cpu", "max_gpu", "max_memory", "max_storage"]
    MIN_CPU_FIELD_NUMBER: _ClassVar[int]
    MIN_GPU_FIELD_NUMBER: _ClassVar[int]
    MIN_MEMORY_FIELD_NUMBER: _ClassVar[int]
    MIN_STORAGE_FIELD_NUMBER: _ClassVar[int]
    MAX_CPU_FIELD_NUMBER: _ClassVar[int]
    MAX_GPU_FIELD_NUMBER: _ClassVar[int]
    MAX_MEMORY_FIELD_NUMBER: _ClassVar[int]
    MAX_STORAGE_FIELD_NUMBER: _ClassVar[int]
    min_cpu: SettingValue
    min_gpu: SettingValue
    min_memory: SettingValue
    min_storage: SettingValue
    max_cpu: SettingValue
    max_gpu: SettingValue
    max_memory: SettingValue
    max_storage: SettingValue
    def __init__(self, min_cpu: _Optional[_Union[SettingValue, _Mapping]] = ..., min_gpu: _Optional[_Union[SettingValue, _Mapping]] = ..., min_memory: _Optional[_Union[SettingValue, _Mapping]] = ..., min_storage: _Optional[_Union[SettingValue, _Mapping]] = ..., max_cpu: _Optional[_Union[SettingValue, _Mapping]] = ..., max_gpu: _Optional[_Union[SettingValue, _Mapping]] = ..., max_memory: _Optional[_Union[SettingValue, _Mapping]] = ..., max_storage: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...

class TaskResourceSettings(_message.Message):
    __slots__ = ["defaults", "mirror_limits_request"]
    DEFAULTS_FIELD_NUMBER: _ClassVar[int]
    MIRROR_LIMITS_REQUEST_FIELD_NUMBER: _ClassVar[int]
    defaults: TaskResourceDefaults
    mirror_limits_request: SettingValue
    def __init__(self, defaults: _Optional[_Union[TaskResourceDefaults, _Mapping]] = ..., mirror_limits_request: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...

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
    labels: SettingValue
    annotations: SettingValue
    environment_variables: SettingValue
    def __init__(self, run: _Optional[_Union[RunSettings, _Mapping]] = ..., security: _Optional[_Union[SecuritySettings, _Mapping]] = ..., storage: _Optional[_Union[StorageSettings, _Mapping]] = ..., task_resource: _Optional[_Union[TaskResourceSettings, _Mapping]] = ..., labels: _Optional[_Union[SettingValue, _Mapping]] = ..., annotations: _Optional[_Union[SettingValue, _Mapping]] = ..., environment_variables: _Optional[_Union[SettingValue, _Mapping]] = ...) -> None: ...
