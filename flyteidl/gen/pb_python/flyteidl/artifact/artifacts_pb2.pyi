from google.protobuf import any_pb2 as _any_pb2
from flyteidl.admin import launch_plan_pb2 as _launch_plan_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import types_pb2 as _types_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Artifact(_message.Message):
    __slots__ = ["artifact_id", "spec", "tags"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _identifier_pb2.ArtifactID
    spec: ArtifactSpec
    tags: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, artifact_id: _Optional[_Union[_identifier_pb2.ArtifactID, _Mapping]] = ..., spec: _Optional[_Union[ArtifactSpec, _Mapping]] = ..., tags: _Optional[_Iterable[str]] = ...) -> None: ...

class CreateArtifactRequest(_message.Message):
    __slots__ = ["artifact_key", "version", "spec", "partitions", "tag"]
    class PartitionsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    TAG_FIELD_NUMBER: _ClassVar[int]
    artifact_key: _identifier_pb2.ArtifactKey
    version: str
    spec: ArtifactSpec
    partitions: _containers.ScalarMap[str, str]
    tag: str
    def __init__(self, artifact_key: _Optional[_Union[_identifier_pb2.ArtifactKey, _Mapping]] = ..., version: _Optional[str] = ..., spec: _Optional[_Union[ArtifactSpec, _Mapping]] = ..., partitions: _Optional[_Mapping[str, str]] = ..., tag: _Optional[str] = ...) -> None: ...

class ArtifactSpec(_message.Message):
    __slots__ = ["value", "type", "task_execution", "execution", "principal", "short_description", "long_description", "user_metadata"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    SHORT_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    LONG_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    USER_METADATA_FIELD_NUMBER: _ClassVar[int]
    value: _literals_pb2.Literal
    type: _types_pb2.LiteralType
    task_execution: _identifier_pb2.TaskExecutionIdentifier
    execution: _identifier_pb2.WorkflowExecutionIdentifier
    principal: str
    short_description: str
    long_description: str
    user_metadata: _any_pb2.Any
    def __init__(self, value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., task_execution: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., execution: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., principal: _Optional[str] = ..., short_description: _Optional[str] = ..., long_description: _Optional[str] = ..., user_metadata: _Optional[_Union[_any_pb2.Any, _Mapping]] = ...) -> None: ...

class CreateArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class GetArtifactRequest(_message.Message):
    __slots__ = ["query", "details"]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    query: _identifier_pb2.ArtifactQuery
    details: bool
    def __init__(self, query: _Optional[_Union[_identifier_pb2.ArtifactQuery, _Mapping]] = ..., details: bool = ...) -> None: ...

class GetArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class ListArtifactNamesRequest(_message.Message):
    __slots__ = ["project", "domain"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ...) -> None: ...

class ListArtifactNamesResponse(_message.Message):
    __slots__ = ["artifact_keys"]
    ARTIFACT_KEYS_FIELD_NUMBER: _ClassVar[int]
    artifact_keys: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.ArtifactKey]
    def __init__(self, artifact_keys: _Optional[_Iterable[_Union[_identifier_pb2.ArtifactKey, _Mapping]]] = ...) -> None: ...

class ListArtifactsRequest(_message.Message):
    __slots__ = ["artifact_key"]
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    artifact_key: _identifier_pb2.ArtifactKey
    def __init__(self, artifact_key: _Optional[_Union[_identifier_pb2.ArtifactKey, _Mapping]] = ...) -> None: ...

class ListArtifactsResponse(_message.Message):
    __slots__ = ["artifacts"]
    ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    artifacts: _containers.RepeatedCompositeFieldContainer[Artifact]
    def __init__(self, artifacts: _Optional[_Iterable[_Union[Artifact, _Mapping]]] = ...) -> None: ...

class AddTagRequest(_message.Message):
    __slots__ = ["artifact_id", "value", "overwrite"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _identifier_pb2.ArtifactID
    value: str
    overwrite: bool
    def __init__(self, artifact_id: _Optional[_Union[_identifier_pb2.ArtifactID, _Mapping]] = ..., value: _Optional[str] = ..., overwrite: bool = ...) -> None: ...

class AddTagResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class CreateTriggerRequest(_message.Message):
    __slots__ = ["trigger_launch_plan"]
    TRIGGER_LAUNCH_PLAN_FIELD_NUMBER: _ClassVar[int]
    trigger_launch_plan: _launch_plan_pb2.LaunchPlan
    def __init__(self, trigger_launch_plan: _Optional[_Union[_launch_plan_pb2.LaunchPlan, _Mapping]] = ...) -> None: ...

class CreateTriggerResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteTriggerRequest(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: _identifier_pb2.Identifier
    def __init__(self, trigger_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class DeleteTriggerResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ArtifactProducer(_message.Message):
    __slots__ = ["entity_id", "outputs"]
    ENTITY_ID_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    entity_id: _identifier_pb2.Identifier
    outputs: _interface_pb2.VariableMap
    def __init__(self, entity_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., outputs: _Optional[_Union[_interface_pb2.VariableMap, _Mapping]] = ...) -> None: ...

class RegisterProducerRequest(_message.Message):
    __slots__ = ["producers"]
    PRODUCERS_FIELD_NUMBER: _ClassVar[int]
    producers: _containers.RepeatedCompositeFieldContainer[ArtifactProducer]
    def __init__(self, producers: _Optional[_Iterable[_Union[ArtifactProducer, _Mapping]]] = ...) -> None: ...

class ArtifactConsumer(_message.Message):
    __slots__ = ["entity_id", "inputs"]
    ENTITY_ID_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    entity_id: _identifier_pb2.Identifier
    inputs: _interface_pb2.ParameterMap
    def __init__(self, entity_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., inputs: _Optional[_Union[_interface_pb2.ParameterMap, _Mapping]] = ...) -> None: ...

class RegisterConsumerRequest(_message.Message):
    __slots__ = ["consumers"]
    CONSUMERS_FIELD_NUMBER: _ClassVar[int]
    consumers: _containers.RepeatedCompositeFieldContainer[ArtifactConsumer]
    def __init__(self, consumers: _Optional[_Iterable[_Union[ArtifactConsumer, _Mapping]]] = ...) -> None: ...

class RegisterResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
