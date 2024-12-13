from google.protobuf import timestamp_pb2 as _timestamp_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Granularity(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNSET: _ClassVar[Granularity]
    MINUTE: _ClassVar[Granularity]
    HOUR: _ClassVar[Granularity]
    DAY: _ClassVar[Granularity]
    MONTH: _ClassVar[Granularity]

class Operator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    MINUS: _ClassVar[Operator]
    PLUS: _ClassVar[Operator]
UNSET: Granularity
MINUTE: Granularity
HOUR: Granularity
DAY: Granularity
MONTH: Granularity
MINUS: Operator
PLUS: Operator

class ArtifactKey(_message.Message):
    __slots__ = ["project", "domain", "name", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    name: str
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., org: _Optional[str] = ...) -> None: ...

class ArtifactBindingData(_message.Message):
    __slots__ = ["partition_key", "bind_to_time_partition", "time_transform"]
    PARTITION_KEY_FIELD_NUMBER: _ClassVar[int]
    BIND_TO_TIME_PARTITION_FIELD_NUMBER: _ClassVar[int]
    TIME_TRANSFORM_FIELD_NUMBER: _ClassVar[int]
    partition_key: str
    bind_to_time_partition: bool
    time_transform: TimeTransform
    def __init__(self, partition_key: _Optional[str] = ..., bind_to_time_partition: bool = ..., time_transform: _Optional[_Union[TimeTransform, _Mapping]] = ...) -> None: ...

class TimeTransform(_message.Message):
    __slots__ = ["transform", "op"]
    TRANSFORM_FIELD_NUMBER: _ClassVar[int]
    OP_FIELD_NUMBER: _ClassVar[int]
    transform: str
    op: Operator
    def __init__(self, transform: _Optional[str] = ..., op: _Optional[_Union[Operator, str]] = ...) -> None: ...

class InputBindingData(_message.Message):
    __slots__ = ["var"]
    VAR_FIELD_NUMBER: _ClassVar[int]
    var: str
    def __init__(self, var: _Optional[str] = ...) -> None: ...

class RuntimeBinding(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class LabelValue(_message.Message):
    __slots__ = ["static_value", "time_value", "triggered_binding", "input_binding", "runtime_binding"]
    STATIC_VALUE_FIELD_NUMBER: _ClassVar[int]
    TIME_VALUE_FIELD_NUMBER: _ClassVar[int]
    TRIGGERED_BINDING_FIELD_NUMBER: _ClassVar[int]
    INPUT_BINDING_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_BINDING_FIELD_NUMBER: _ClassVar[int]
    static_value: str
    time_value: _timestamp_pb2.Timestamp
    triggered_binding: ArtifactBindingData
    input_binding: InputBindingData
    runtime_binding: RuntimeBinding
    def __init__(self, static_value: _Optional[str] = ..., time_value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., triggered_binding: _Optional[_Union[ArtifactBindingData, _Mapping]] = ..., input_binding: _Optional[_Union[InputBindingData, _Mapping]] = ..., runtime_binding: _Optional[_Union[RuntimeBinding, _Mapping]] = ...) -> None: ...

class Partitions(_message.Message):
    __slots__ = ["value"]
    class ValueEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: LabelValue
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[LabelValue, _Mapping]] = ...) -> None: ...
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: _containers.MessageMap[str, LabelValue]
    def __init__(self, value: _Optional[_Mapping[str, LabelValue]] = ...) -> None: ...

class TimePartition(_message.Message):
    __slots__ = ["value", "granularity"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    GRANULARITY_FIELD_NUMBER: _ClassVar[int]
    value: LabelValue
    granularity: Granularity
    def __init__(self, value: _Optional[_Union[LabelValue, _Mapping]] = ..., granularity: _Optional[_Union[Granularity, str]] = ...) -> None: ...

class ArtifactID(_message.Message):
    __slots__ = ["artifact_key", "version", "partitions", "time_partition"]
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    TIME_PARTITION_FIELD_NUMBER: _ClassVar[int]
    artifact_key: ArtifactKey
    version: str
    partitions: Partitions
    time_partition: TimePartition
    def __init__(self, artifact_key: _Optional[_Union[ArtifactKey, _Mapping]] = ..., version: _Optional[str] = ..., partitions: _Optional[_Union[Partitions, _Mapping]] = ..., time_partition: _Optional[_Union[TimePartition, _Mapping]] = ...) -> None: ...

class ArtifactTag(_message.Message):
    __slots__ = ["artifact_key", "value"]
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    artifact_key: ArtifactKey
    value: LabelValue
    def __init__(self, artifact_key: _Optional[_Union[ArtifactKey, _Mapping]] = ..., value: _Optional[_Union[LabelValue, _Mapping]] = ...) -> None: ...

class ArtifactQuery(_message.Message):
    __slots__ = ["artifact_id", "artifact_tag", "uri", "binding"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_TAG_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    BINDING_FIELD_NUMBER: _ClassVar[int]
    artifact_id: ArtifactID
    artifact_tag: ArtifactTag
    uri: str
    binding: ArtifactBindingData
    def __init__(self, artifact_id: _Optional[_Union[ArtifactID, _Mapping]] = ..., artifact_tag: _Optional[_Union[ArtifactTag, _Mapping]] = ..., uri: _Optional[str] = ..., binding: _Optional[_Union[ArtifactBindingData, _Mapping]] = ...) -> None: ...
