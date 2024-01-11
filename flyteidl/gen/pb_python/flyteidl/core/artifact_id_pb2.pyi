from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ArtifactKey(_message.Message):
    __slots__ = ["project", "domain", "name"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    name: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ArtifactBindingData(_message.Message):
    __slots__ = ["index", "partition_key", "transform"]
    INDEX_FIELD_NUMBER: _ClassVar[int]
    PARTITION_KEY_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_FIELD_NUMBER: _ClassVar[int]
    index: int
    partition_key: str
    transform: str
    def __init__(self, index: _Optional[int] = ..., partition_key: _Optional[str] = ..., transform: _Optional[str] = ...) -> None: ...

class InputBindingData(_message.Message):
    __slots__ = ["var"]
    VAR_FIELD_NUMBER: _ClassVar[int]
    var: str
    def __init__(self, var: _Optional[str] = ...) -> None: ...

class LabelValue(_message.Message):
    __slots__ = ["static_value", "triggered_binding", "input_binding"]
    STATIC_VALUE_FIELD_NUMBER: _ClassVar[int]
    TRIGGERED_BINDING_FIELD_NUMBER: _ClassVar[int]
    INPUT_BINDING_FIELD_NUMBER: _ClassVar[int]
    static_value: str
    triggered_binding: ArtifactBindingData
    input_binding: InputBindingData
    def __init__(self, static_value: _Optional[str] = ..., triggered_binding: _Optional[_Union[ArtifactBindingData, _Mapping]] = ..., input_binding: _Optional[_Union[InputBindingData, _Mapping]] = ...) -> None: ...

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

class ArtifactID(_message.Message):
    __slots__ = ["artifact_key", "version", "partitions"]
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    artifact_key: ArtifactKey
    version: str
    partitions: Partitions
    def __init__(self, artifact_key: _Optional[_Union[ArtifactKey, _Mapping]] = ..., version: _Optional[str] = ..., partitions: _Optional[_Union[Partitions, _Mapping]] = ...) -> None: ...

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

class Trigger(_message.Message):
    __slots__ = ["trigger_id", "triggers"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    TRIGGERS_FIELD_NUMBER: _ClassVar[int]
    trigger_id: _identifier_pb2.Identifier
    triggers: _containers.RepeatedCompositeFieldContainer[ArtifactID]
    def __init__(self, trigger_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., triggers: _Optional[_Iterable[_Union[ArtifactID, _Mapping]]] = ...) -> None: ...
