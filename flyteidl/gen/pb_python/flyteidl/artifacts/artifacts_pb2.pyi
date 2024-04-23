from google.protobuf import struct_pb2 as _struct_pb2
from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from flyteidl.admin import launch_plan_pb2 as _launch_plan_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import types_pb2 as _types_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import artifact_id_pb2 as _artifact_id_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.event import cloudevents_pb2 as _cloudevents_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Artifact(_message.Message):
    __slots__ = ["artifact_id", "spec", "tags", "source", "metadata"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_id_pb2.ArtifactID
    spec: ArtifactSpec
    tags: _containers.RepeatedScalarFieldContainer[str]
    source: ArtifactSource
    metadata: ArtifactMetadata
    def __init__(self, artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ..., spec: _Optional[_Union[ArtifactSpec, _Mapping]] = ..., tags: _Optional[_Iterable[str]] = ..., source: _Optional[_Union[ArtifactSource, _Mapping]] = ..., metadata: _Optional[_Union[ArtifactMetadata, _Mapping]] = ...) -> None: ...

class ArtifactMetadata(_message.Message):
    __slots__ = ["created_at", "uri"]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    created_at: _timestamp_pb2.Timestamp
    uri: str
    def __init__(self, created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., uri: _Optional[str] = ...) -> None: ...

class CreateArtifactRequest(_message.Message):
    __slots__ = ["artifact_key", "version", "spec", "partitions", "time_partition_value", "source"]
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
    TIME_PARTITION_VALUE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    artifact_key: _artifact_id_pb2.ArtifactKey
    version: str
    spec: ArtifactSpec
    partitions: _containers.ScalarMap[str, str]
    time_partition_value: _timestamp_pb2.Timestamp
    source: ArtifactSource
    def __init__(self, artifact_key: _Optional[_Union[_artifact_id_pb2.ArtifactKey, _Mapping]] = ..., version: _Optional[str] = ..., spec: _Optional[_Union[ArtifactSpec, _Mapping]] = ..., partitions: _Optional[_Mapping[str, str]] = ..., time_partition_value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., source: _Optional[_Union[ArtifactSource, _Mapping]] = ...) -> None: ...

class ArtifactSource(_message.Message):
    __slots__ = ["workflow_execution", "node_id", "task_id", "retry_attempt", "principal"]
    WORKFLOW_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    workflow_execution: _identifier_pb2.WorkflowExecutionIdentifier
    node_id: str
    task_id: _identifier_pb2.Identifier
    retry_attempt: int
    principal: str
    def __init__(self, workflow_execution: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., node_id: _Optional[str] = ..., task_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ..., principal: _Optional[str] = ...) -> None: ...

class Card(_message.Message):
    __slots__ = ["uri", "type", "body", "text_type"]
    class CardType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[Card.CardType]
        MODEL: _ClassVar[Card.CardType]
        DATASET: _ClassVar[Card.CardType]
    UNKNOWN: Card.CardType
    MODEL: Card.CardType
    DATASET: Card.CardType
    class TextType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        MARKDOWN: _ClassVar[Card.TextType]
    MARKDOWN: Card.TextType
    URI_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    TEXT_TYPE_FIELD_NUMBER: _ClassVar[int]
    uri: str
    type: Card.CardType
    body: str
    text_type: Card.TextType
    def __init__(self, uri: _Optional[str] = ..., type: _Optional[_Union[Card.CardType, str]] = ..., body: _Optional[str] = ..., text_type: _Optional[_Union[Card.TextType, str]] = ...) -> None: ...

class ArtifactSpec(_message.Message):
    __slots__ = ["value", "type", "short_description", "user_metadata", "card"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    SHORT_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    USER_METADATA_FIELD_NUMBER: _ClassVar[int]
    CARD_FIELD_NUMBER: _ClassVar[int]
    value: _literals_pb2.Literal
    type: _types_pb2.LiteralType
    short_description: str
    user_metadata: _struct_pb2.Struct
    card: Card
    def __init__(self, value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., short_description: _Optional[str] = ..., user_metadata: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., card: _Optional[_Union[Card, _Mapping]] = ...) -> None: ...

class Trigger(_message.Message):
    __slots__ = ["trigger", "trigger_inputs"]
    TRIGGER_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_INPUTS_FIELD_NUMBER: _ClassVar[int]
    trigger: _artifact_id_pb2.ArtifactID
    trigger_inputs: _interface_pb2.ParameterMap
    def __init__(self, trigger: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ..., trigger_inputs: _Optional[_Union[_interface_pb2.ParameterMap, _Mapping]] = ...) -> None: ...

class CreateArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class GetArtifactRequest(_message.Message):
    __slots__ = ["query", "details"]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    query: _artifact_id_pb2.ArtifactQuery
    details: bool
    def __init__(self, query: _Optional[_Union[_artifact_id_pb2.ArtifactQuery, _Mapping]] = ..., details: bool = ...) -> None: ...

class GetArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class SearchOptions(_message.Message):
    __slots__ = ["strict_partitions", "latest_by_key"]
    STRICT_PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    LATEST_BY_KEY_FIELD_NUMBER: _ClassVar[int]
    strict_partitions: bool
    latest_by_key: bool
    def __init__(self, strict_partitions: bool = ..., latest_by_key: bool = ...) -> None: ...

class SearchArtifactsRequest(_message.Message):
    __slots__ = ["artifact_key", "partitions", "time_partition_value", "principal", "version", "options", "token", "limit", "granularity"]
    ARTIFACT_KEY_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    TIME_PARTITION_VALUE_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    GRANULARITY_FIELD_NUMBER: _ClassVar[int]
    artifact_key: _artifact_id_pb2.ArtifactKey
    partitions: _artifact_id_pb2.Partitions
    time_partition_value: _timestamp_pb2.Timestamp
    principal: str
    version: str
    options: SearchOptions
    token: str
    limit: int
    granularity: _artifact_id_pb2.Granularity
    def __init__(self, artifact_key: _Optional[_Union[_artifact_id_pb2.ArtifactKey, _Mapping]] = ..., partitions: _Optional[_Union[_artifact_id_pb2.Partitions, _Mapping]] = ..., time_partition_value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., principal: _Optional[str] = ..., version: _Optional[str] = ..., options: _Optional[_Union[SearchOptions, _Mapping]] = ..., token: _Optional[str] = ..., limit: _Optional[int] = ..., granularity: _Optional[_Union[_artifact_id_pb2.Granularity, str]] = ...) -> None: ...

class SearchArtifactsResponse(_message.Message):
    __slots__ = ["artifacts", "token"]
    ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    artifacts: _containers.RepeatedCompositeFieldContainer[Artifact]
    token: str
    def __init__(self, artifacts: _Optional[_Iterable[_Union[Artifact, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class FindByWorkflowExecRequest(_message.Message):
    __slots__ = ["exec_id", "direction"]
    class Direction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        INPUTS: _ClassVar[FindByWorkflowExecRequest.Direction]
        OUTPUTS: _ClassVar[FindByWorkflowExecRequest.Direction]
    INPUTS: FindByWorkflowExecRequest.Direction
    OUTPUTS: FindByWorkflowExecRequest.Direction
    EXEC_ID_FIELD_NUMBER: _ClassVar[int]
    DIRECTION_FIELD_NUMBER: _ClassVar[int]
    exec_id: _identifier_pb2.WorkflowExecutionIdentifier
    direction: FindByWorkflowExecRequest.Direction
    def __init__(self, exec_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., direction: _Optional[_Union[FindByWorkflowExecRequest.Direction, str]] = ...) -> None: ...

class AddTagRequest(_message.Message):
    __slots__ = ["artifact_id", "value", "overwrite"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_id_pb2.ArtifactID
    value: str
    overwrite: bool
    def __init__(self, artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ..., value: _Optional[str] = ..., overwrite: bool = ...) -> None: ...

class AddTagResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ActivateTriggerRequest(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: _identifier_pb2.Identifier
    def __init__(self, trigger_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class ActivateTriggerResponse(_message.Message):
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

class DeactivateTriggerRequest(_message.Message):
    __slots__ = ["trigger_id"]
    TRIGGER_ID_FIELD_NUMBER: _ClassVar[int]
    trigger_id: _identifier_pb2.Identifier
    def __init__(self, trigger_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class DeactivateTriggerResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeactivateAllTriggersRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeactivateAllTriggersResponse(_message.Message):
    __slots__ = ["num_deactivated"]
    NUM_DEACTIVATED_FIELD_NUMBER: _ClassVar[int]
    num_deactivated: int
    def __init__(self, num_deactivated: _Optional[int] = ...) -> None: ...

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

class ExecutionInputsRequest(_message.Message):
    __slots__ = ["execution_id", "inputs"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    inputs: _containers.RepeatedCompositeFieldContainer[_artifact_id_pb2.ArtifactID]
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., inputs: _Optional[_Iterable[_Union[_artifact_id_pb2.ArtifactID, _Mapping]]] = ...) -> None: ...

class ExecutionInputsResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListUsageRequest(_message.Message):
    __slots__ = ["artifact_id"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_id_pb2.ArtifactID
    def __init__(self, artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ...) -> None: ...

class ListUsageResponse(_message.Message):
    __slots__ = ["executions"]
    EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    executions: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.WorkflowExecutionIdentifier]
    def __init__(self, executions: _Optional[_Iterable[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]]] = ...) -> None: ...

class GetCardRequest(_message.Message):
    __slots__ = ["artifact_id"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_id_pb2.ArtifactID
    def __init__(self, artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ...) -> None: ...

class GetCardResponse(_message.Message):
    __slots__ = ["card"]
    CARD_FIELD_NUMBER: _ClassVar[int]
    card: Card
    def __init__(self, card: _Optional[_Union[Card, _Mapping]] = ...) -> None: ...

class GetTriggeringArtifactsRequest(_message.Message):
    __slots__ = ["executions"]
    EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    executions: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.WorkflowExecutionIdentifier]
    def __init__(self, executions: _Optional[_Iterable[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]]] = ...) -> None: ...

class GetTriggeringArtifactsResponse(_message.Message):
    __slots__ = ["artifacts"]
    class ArtifactsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _artifact_id_pb2.ArtifactID
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ...) -> None: ...
    ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    artifacts: _containers.MessageMap[str, _artifact_id_pb2.ArtifactID]
    def __init__(self, artifacts: _Optional[_Mapping[str, _artifact_id_pb2.ArtifactID]] = ...) -> None: ...

class GetTriggeredExecutionsByArtifactRequest(_message.Message):
    __slots__ = ["artifact_id"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_id_pb2.ArtifactID
    def __init__(self, artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ...) -> None: ...

class GetTriggeredExecutionsByArtifactResponse(_message.Message):
    __slots__ = ["executions"]
    EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    executions: _containers.RepeatedCompositeFieldContainer[_identifier_pb2.WorkflowExecutionIdentifier]
    def __init__(self, executions: _Optional[_Iterable[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]]] = ...) -> None: ...
