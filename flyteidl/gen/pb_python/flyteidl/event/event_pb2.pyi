from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import compiler_pb2 as _compiler_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import catalog_pb2 as _catalog_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowExecutionEvent(_message.Message):
    __slots__ = ["execution_id", "producer_id", "phase", "occurred_at", "output_uri", "error", "output_data"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    producer_id: str
    phase: _execution_pb2.WorkflowExecution.Phase
    occurred_at: _timestamp_pb2.Timestamp
    output_uri: str
    error: _execution_pb2.ExecutionError
    output_data: _literals_pb2.LiteralMap
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., producer_id: _Optional[str] = ..., phase: _Optional[_Union[_execution_pb2.WorkflowExecution.Phase, str]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class NodeExecutionEvent(_message.Message):
    __slots__ = ["id", "producer_id", "phase", "occurred_at", "input_uri", "input_data", "output_uri", "error", "output_data", "workflow_node_metadata", "task_node_metadata", "parent_task_metadata", "parent_node_metadata", "retry_group", "spec_node_id", "node_name", "event_version", "is_parent", "is_dynamic", "deck_uri", "reported_at", "is_array", "target_entity", "is_in_dynamic_chain", "is_eager"]
    ID_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    TASK_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    PARENT_TASK_METADATA_FIELD_NUMBER: _ClassVar[int]
    PARENT_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    RETRY_GROUP_FIELD_NUMBER: _ClassVar[int]
    SPEC_NODE_ID_FIELD_NUMBER: _ClassVar[int]
    NODE_NAME_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    IS_PARENT_FIELD_NUMBER: _ClassVar[int]
    IS_DYNAMIC_FIELD_NUMBER: _ClassVar[int]
    DECK_URI_FIELD_NUMBER: _ClassVar[int]
    REPORTED_AT_FIELD_NUMBER: _ClassVar[int]
    IS_ARRAY_FIELD_NUMBER: _ClassVar[int]
    TARGET_ENTITY_FIELD_NUMBER: _ClassVar[int]
    IS_IN_DYNAMIC_CHAIN_FIELD_NUMBER: _ClassVar[int]
    IS_EAGER_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.NodeExecutionIdentifier
    producer_id: str
    phase: _execution_pb2.NodeExecution.Phase
    occurred_at: _timestamp_pb2.Timestamp
    input_uri: str
    input_data: _literals_pb2.LiteralMap
    output_uri: str
    error: _execution_pb2.ExecutionError
    output_data: _literals_pb2.LiteralMap
    workflow_node_metadata: WorkflowNodeMetadata
    task_node_metadata: TaskNodeMetadata
    parent_task_metadata: ParentTaskExecutionMetadata
    parent_node_metadata: ParentNodeExecutionMetadata
    retry_group: str
    spec_node_id: str
    node_name: str
    event_version: int
    is_parent: bool
    is_dynamic: bool
    deck_uri: str
    reported_at: _timestamp_pb2.Timestamp
    is_array: bool
    target_entity: _identifier_pb2.Identifier
    is_in_dynamic_chain: bool
    is_eager: bool
    def __init__(self, id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., producer_id: _Optional[str] = ..., phase: _Optional[_Union[_execution_pb2.NodeExecution.Phase, str]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., input_uri: _Optional[str] = ..., input_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., workflow_node_metadata: _Optional[_Union[WorkflowNodeMetadata, _Mapping]] = ..., task_node_metadata: _Optional[_Union[TaskNodeMetadata, _Mapping]] = ..., parent_task_metadata: _Optional[_Union[ParentTaskExecutionMetadata, _Mapping]] = ..., parent_node_metadata: _Optional[_Union[ParentNodeExecutionMetadata, _Mapping]] = ..., retry_group: _Optional[str] = ..., spec_node_id: _Optional[str] = ..., node_name: _Optional[str] = ..., event_version: _Optional[int] = ..., is_parent: bool = ..., is_dynamic: bool = ..., deck_uri: _Optional[str] = ..., reported_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., is_array: bool = ..., target_entity: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., is_in_dynamic_chain: bool = ..., is_eager: bool = ...) -> None: ...

class WorkflowNodeMetadata(_message.Message):
    __slots__ = ["execution_id"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskNodeMetadata(_message.Message):
    __slots__ = ["cache_status", "catalog_key", "reservation_status", "checkpoint_uri", "dynamic_workflow"]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    CATALOG_KEY_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_STATUS_FIELD_NUMBER: _ClassVar[int]
    CHECKPOINT_URI_FIELD_NUMBER: _ClassVar[int]
    DYNAMIC_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    cache_status: _catalog_pb2.CatalogCacheStatus
    catalog_key: _catalog_pb2.CatalogMetadata
    reservation_status: _catalog_pb2.CatalogReservation.Status
    checkpoint_uri: str
    dynamic_workflow: DynamicWorkflowNodeMetadata
    def __init__(self, cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., catalog_key: _Optional[_Union[_catalog_pb2.CatalogMetadata, _Mapping]] = ..., reservation_status: _Optional[_Union[_catalog_pb2.CatalogReservation.Status, str]] = ..., checkpoint_uri: _Optional[str] = ..., dynamic_workflow: _Optional[_Union[DynamicWorkflowNodeMetadata, _Mapping]] = ...) -> None: ...

class DynamicWorkflowNodeMetadata(_message.Message):
    __slots__ = ["id", "compiled_workflow", "dynamic_job_spec_uri"]
    ID_FIELD_NUMBER: _ClassVar[int]
    COMPILED_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    DYNAMIC_JOB_SPEC_URI_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    compiled_workflow: _compiler_pb2.CompiledWorkflowClosure
    dynamic_job_spec_uri: str
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., compiled_workflow: _Optional[_Union[_compiler_pb2.CompiledWorkflowClosure, _Mapping]] = ..., dynamic_job_spec_uri: _Optional[str] = ...) -> None: ...

class ParentTaskExecutionMetadata(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class ParentNodeExecutionMetadata(_message.Message):
    __slots__ = ["node_id"]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    node_id: str
    def __init__(self, node_id: _Optional[str] = ...) -> None: ...

class EventReason(_message.Message):
    __slots__ = ["reason", "occurred_at"]
    REASON_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    reason: str
    occurred_at: _timestamp_pb2.Timestamp
    def __init__(self, reason: _Optional[str] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class TaskExecutionEvent(_message.Message):
    __slots__ = ["task_id", "parent_node_execution_id", "retry_attempt", "phase", "producer_id", "logs", "occurred_at", "input_uri", "input_data", "output_uri", "error", "output_data", "custom_info", "phase_version", "reason", "reasons", "task_type", "metadata", "event_version", "reported_at"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    PHASE_VERSION_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    REASONS_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    REPORTED_AT_FIELD_NUMBER: _ClassVar[int]
    task_id: _identifier_pb2.Identifier
    parent_node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    retry_attempt: int
    phase: _execution_pb2.TaskExecution.Phase
    producer_id: str
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    occurred_at: _timestamp_pb2.Timestamp
    input_uri: str
    input_data: _literals_pb2.LiteralMap
    output_uri: str
    error: _execution_pb2.ExecutionError
    output_data: _literals_pb2.LiteralMap
    custom_info: _struct_pb2.Struct
    phase_version: int
    reason: str
    reasons: _containers.RepeatedCompositeFieldContainer[EventReason]
    task_type: str
    metadata: TaskExecutionMetadata
    event_version: int
    reported_at: _timestamp_pb2.Timestamp
    def __init__(self, task_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., parent_node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., producer_id: _Optional[str] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., input_uri: _Optional[str] = ..., input_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., phase_version: _Optional[int] = ..., reason: _Optional[str] = ..., reasons: _Optional[_Iterable[_Union[EventReason, _Mapping]]] = ..., task_type: _Optional[str] = ..., metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ..., event_version: _Optional[int] = ..., reported_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ExternalResourceInfo(_message.Message):
    __slots__ = ["external_id", "index", "retry_attempt", "phase", "cache_status", "logs", "workflow_node_metadata", "custom_info"]
    EXTERNAL_ID_FIELD_NUMBER: _ClassVar[int]
    INDEX_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    external_id: str
    index: int
    retry_attempt: int
    phase: _execution_pb2.TaskExecution.Phase
    cache_status: _catalog_pb2.CatalogCacheStatus
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    workflow_node_metadata: WorkflowNodeMetadata
    custom_info: _struct_pb2.Struct
    def __init__(self, external_id: _Optional[str] = ..., index: _Optional[int] = ..., retry_attempt: _Optional[int] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., workflow_node_metadata: _Optional[_Union[WorkflowNodeMetadata, _Mapping]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class ResourcePoolInfo(_message.Message):
    __slots__ = ["allocation_token", "namespace"]
    ALLOCATION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    allocation_token: str
    namespace: str
    def __init__(self, allocation_token: _Optional[str] = ..., namespace: _Optional[str] = ...) -> None: ...

class TaskExecutionMetadata(_message.Message):
    __slots__ = ["generated_name", "external_resources", "resource_pool_info", "plugin_identifier", "instance_class"]
    class InstanceClass(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        DEFAULT: _ClassVar[TaskExecutionMetadata.InstanceClass]
        INTERRUPTIBLE: _ClassVar[TaskExecutionMetadata.InstanceClass]
    DEFAULT: TaskExecutionMetadata.InstanceClass
    INTERRUPTIBLE: TaskExecutionMetadata.InstanceClass
    GENERATED_NAME_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_POOL_INFO_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    INSTANCE_CLASS_FIELD_NUMBER: _ClassVar[int]
    generated_name: str
    external_resources: _containers.RepeatedCompositeFieldContainer[ExternalResourceInfo]
    resource_pool_info: _containers.RepeatedCompositeFieldContainer[ResourcePoolInfo]
    plugin_identifier: str
    instance_class: TaskExecutionMetadata.InstanceClass
    def __init__(self, generated_name: _Optional[str] = ..., external_resources: _Optional[_Iterable[_Union[ExternalResourceInfo, _Mapping]]] = ..., resource_pool_info: _Optional[_Iterable[_Union[ResourcePoolInfo, _Mapping]]] = ..., plugin_identifier: _Optional[str] = ..., instance_class: _Optional[_Union[TaskExecutionMetadata.InstanceClass, str]] = ...) -> None: ...
