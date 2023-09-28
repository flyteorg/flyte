from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ResourceType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNSPECIFIED: _ClassVar[ResourceType]
    TASK: _ClassVar[ResourceType]
    WORKFLOW: _ClassVar[ResourceType]
    LAUNCH_PLAN: _ClassVar[ResourceType]
    DATASET: _ClassVar[ResourceType]
UNSPECIFIED: ResourceType
TASK: ResourceType
WORKFLOW: ResourceType
LAUNCH_PLAN: ResourceType
DATASET: ResourceType

class Identifier(_message.Message):
    __slots__ = ["resource_type", "project", "domain", "name", "version"]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    resource_type: ResourceType
    project: str
    domain: str
    name: str
    version: str
    def __init__(self, resource_type: _Optional[_Union[ResourceType, str]] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class WorkflowExecutionIdentifier(_message.Message):
    __slots__ = ["project", "domain", "name"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    name: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class NodeExecutionIdentifier(_message.Message):
    __slots__ = ["node_id", "execution_id"]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    node_id: str
    execution_id: WorkflowExecutionIdentifier
    def __init__(self, node_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionIdentifier(_message.Message):
    __slots__ = ["task_id", "node_execution_id", "retry_attempt"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    task_id: Identifier
    node_execution_id: NodeExecutionIdentifier
    retry_attempt: int
    def __init__(self, task_id: _Optional[_Union[Identifier, _Mapping]] = ..., node_execution_id: _Optional[_Union[NodeExecutionIdentifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ...) -> None: ...

class SignalIdentifier(_message.Message):
    __slots__ = ["signal_id", "execution_id"]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    signal_id: str
    execution_id: WorkflowExecutionIdentifier
    def __init__(self, signal_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

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

class PartitionValue(_message.Message):
    __slots__ = ["static_value", "binding"]
    STATIC_VALUE_FIELD_NUMBER: _ClassVar[int]
    BINDING_FIELD_NUMBER: _ClassVar[int]
    static_value: str
    binding: ArtifactBindingData
    def __init__(self, static_value: _Optional[str] = ..., binding: _Optional[_Union[ArtifactBindingData, _Mapping]] = ...) -> None: ...

class Partitions(_message.Message):
    __slots__ = ["value"]
    class ValueEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: PartitionValue
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[PartitionValue, _Mapping]] = ...) -> None: ...
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: _containers.MessageMap[str, PartitionValue]
    def __init__(self, value: _Optional[_Mapping[str, PartitionValue]] = ...) -> None: ...

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
    value: str
    def __init__(self, artifact_key: _Optional[_Union[ArtifactKey, _Mapping]] = ..., value: _Optional[str] = ...) -> None: ...

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
    trigger_id: Identifier
    triggers: _containers.RepeatedCompositeFieldContainer[ArtifactID]
    def __init__(self, trigger_id: _Optional[_Union[Identifier, _Mapping]] = ..., triggers: _Optional[_Iterable[_Union[ArtifactID, _Mapping]]] = ...) -> None: ...
