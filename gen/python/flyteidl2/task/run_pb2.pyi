from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import phase_pb2 as _phase_pb2
from flyteidl2.core import literals_pb2 as _literals_pb2
from flyteidl2.core import security_pb2 as _security_pb2
from flyteidl2.notification import definition_pb2 as _definition_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CacheLookupScope(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    CACHE_LOOKUP_SCOPE_UNSPECIFIED: _ClassVar[CacheLookupScope]
    CACHE_LOOKUP_SCOPE_GLOBAL: _ClassVar[CacheLookupScope]
    CACHE_LOOKUP_SCOPE_PROJECT_DOMAIN: _ClassVar[CacheLookupScope]
CACHE_LOOKUP_SCOPE_UNSPECIFIED: CacheLookupScope
CACHE_LOOKUP_SCOPE_GLOBAL: CacheLookupScope
CACHE_LOOKUP_SCOPE_PROJECT_DOMAIN: CacheLookupScope

class Labels(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Annotations(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Envs(_message.Message):
    __slots__ = ["values"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    def __init__(self, values: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ...) -> None: ...

class RawDataStorage(_message.Message):
    __slots__ = ["raw_data_prefix"]
    RAW_DATA_PREFIX_FIELD_NUMBER: _ClassVar[int]
    raw_data_prefix: str
    def __init__(self, raw_data_prefix: _Optional[str] = ...) -> None: ...

class CacheConfig(_message.Message):
    __slots__ = ["overwrite_cache", "cache_lookup_scope"]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    CACHE_LOOKUP_SCOPE_FIELD_NUMBER: _ClassVar[int]
    overwrite_cache: bool
    cache_lookup_scope: CacheLookupScope
    def __init__(self, overwrite_cache: bool = ..., cache_lookup_scope: _Optional[_Union[CacheLookupScope, str]] = ...) -> None: ...

class RunSpec(_message.Message):
    __slots__ = ["labels", "annotations", "envs", "interruptible", "overwrite_cache", "cluster", "raw_data_storage", "security_context", "cache_config", "notification_rule_name", "notification_rules"]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    RAW_DATA_STORAGE_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    CACHE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATION_RULE_NAME_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATION_RULES_FIELD_NUMBER: _ClassVar[int]
    labels: Labels
    annotations: Annotations
    envs: Envs
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    cluster: str
    raw_data_storage: RawDataStorage
    security_context: _security_pb2.SecurityContext
    cache_config: CacheConfig
    notification_rule_name: str
    notification_rules: InlineRuleList
    def __init__(self, labels: _Optional[_Union[Labels, _Mapping]] = ..., annotations: _Optional[_Union[Annotations, _Mapping]] = ..., envs: _Optional[_Union[Envs, _Mapping]] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., cluster: _Optional[str] = ..., raw_data_storage: _Optional[_Union[RawDataStorage, _Mapping]] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., cache_config: _Optional[_Union[CacheConfig, _Mapping]] = ..., notification_rule_name: _Optional[str] = ..., notification_rules: _Optional[_Union[InlineRuleList, _Mapping]] = ...) -> None: ...

class InlineRuleList(_message.Message):
    __slots__ = ["rules"]
    RULES_FIELD_NUMBER: _ClassVar[int]
    rules: _containers.RepeatedCompositeFieldContainer[InlineRule]
    def __init__(self, rules: _Optional[_Iterable[_Union[InlineRule, _Mapping]]] = ...) -> None: ...

class InlineRule(_message.Message):
    __slots__ = ["on_phases", "delivery_config_name", "delivery_template"]
    ON_PHASES_FIELD_NUMBER: _ClassVar[int]
    DELIVERY_CONFIG_NAME_FIELD_NUMBER: _ClassVar[int]
    DELIVERY_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    on_phases: _containers.RepeatedScalarFieldContainer[_phase_pb2.ActionPhase]
    delivery_config_name: str
    delivery_template: _definition_pb2.DeliveryConfigTemplate
    def __init__(self, on_phases: _Optional[_Iterable[_Union[_phase_pb2.ActionPhase, str]]] = ..., delivery_config_name: _Optional[str] = ..., delivery_template: _Optional[_Union[_definition_pb2.DeliveryConfigTemplate, _Mapping]] = ...) -> None: ...
