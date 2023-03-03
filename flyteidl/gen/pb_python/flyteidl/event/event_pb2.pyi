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

class DynamicWorkflowNodeMetadata(_message.Message):
    __slots__ = ["compiled_workflow", "dynamic_job_spec_uri", "id"]
    COMPILED_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    DYNAMIC_JOB_SPEC_URI_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    compiled_workflow: _compiler_pb2.CompiledWorkflowClosure
    dynamic_job_spec_uri: str
    id: _identifier_pb2.Identifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., compiled_workflow: _Optional[_Union[_compiler_pb2.CompiledWorkflowClosure, _Mapping]] = ..., dynamic_job_spec_uri: _Optional[str] = ...) -> None: ...

class ExternalResourceInfo(_message.Message):
    __slots__ = ["cache_status", "external_id", "index", "logs", "phase", "retry_attempt"]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_ID_FIELD_NUMBER: _ClassVar[int]
    INDEX_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    cache_status: _catalog_pb2.CatalogCacheStatus
    external_id: str
    index: int
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    phase: _execution_pb2.TaskExecution.Phase
    retry_attempt: int
    def __init__(self, external_id: _Optional[str] = ..., index: _Optional[int] = ..., retry_attempt: _Optional[int] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ...) -> None: ...

class NodeExecutionEvent(_message.Message):
    __slots__ = ["deck_uri", "error", "event_version", "id", "input_data", "input_uri", "is_dynamic", "is_parent", "node_name", "occurred_at", "output_data", "output_uri", "parent_node_metadata", "parent_task_metadata", "phase", "producer_id", "retry_group", "spec_node_id", "task_node_metadata", "workflow_node_metadata"]
    DECK_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    IS_DYNAMIC_FIELD_NUMBER: _ClassVar[int]
    IS_PARENT_FIELD_NUMBER: _ClassVar[int]
    NODE_NAME_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    PARENT_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    PARENT_TASK_METADATA_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_GROUP_FIELD_NUMBER: _ClassVar[int]
    SPEC_NODE_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    deck_uri: str
    error: _execution_pb2.ExecutionError
    event_version: int
    id: _identifier_pb2.NodeExecutionIdentifier
    input_data: _literals_pb2.LiteralMap
    input_uri: str
    is_dynamic: bool
    is_parent: bool
    node_name: str
    occurred_at: _timestamp_pb2.Timestamp
    output_data: _literals_pb2.LiteralMap
    output_uri: str
    parent_node_metadata: ParentNodeExecutionMetadata
    parent_task_metadata: ParentTaskExecutionMetadata
    phase: _execution_pb2.NodeExecution.Phase
    producer_id: str
    retry_group: str
    spec_node_id: str
    task_node_metadata: TaskNodeMetadata
    workflow_node_metadata: WorkflowNodeMetadata
    def __init__(self, id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., producer_id: _Optional[str] = ..., phase: _Optional[_Union[_execution_pb2.NodeExecution.Phase, str]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., input_uri: _Optional[str] = ..., input_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., workflow_node_metadata: _Optional[_Union[WorkflowNodeMetadata, _Mapping]] = ..., task_node_metadata: _Optional[_Union[TaskNodeMetadata, _Mapping]] = ..., parent_task_metadata: _Optional[_Union[ParentTaskExecutionMetadata, _Mapping]] = ..., parent_node_metadata: _Optional[_Union[ParentNodeExecutionMetadata, _Mapping]] = ..., retry_group: _Optional[str] = ..., spec_node_id: _Optional[str] = ..., node_name: _Optional[str] = ..., event_version: _Optional[int] = ..., is_parent: bool = ..., is_dynamic: bool = ..., deck_uri: _Optional[str] = ...) -> None: ...

class ParentNodeExecutionMetadata(_message.Message):
    __slots__ = ["node_id"]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    node_id: str
    def __init__(self, node_id: _Optional[str] = ...) -> None: ...

class ParentTaskExecutionMetadata(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class ResourcePoolInfo(_message.Message):
    __slots__ = ["allocation_token", "namespace"]
    ALLOCATION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    allocation_token: str
    namespace: str
    def __init__(self, allocation_token: _Optional[str] = ..., namespace: _Optional[str] = ...) -> None: ...

class TaskExecutionEvent(_message.Message):
    __slots__ = ["custom_info", "error", "event_version", "input_data", "input_uri", "logs", "metadata", "occurred_at", "output_data", "output_uri", "parent_node_execution_id", "phase", "phase_version", "producer_id", "reason", "retry_attempt", "task_id", "task_type"]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    EVENT_VERSION_FIELD_NUMBER: _ClassVar[int]
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    PARENT_NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    PHASE_VERSION_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    custom_info: _struct_pb2.Struct
    error: _execution_pb2.ExecutionError
    event_version: int
    input_data: _literals_pb2.LiteralMap
    input_uri: str
    logs: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    metadata: TaskExecutionMetadata
    occurred_at: _timestamp_pb2.Timestamp
    output_data: _literals_pb2.LiteralMap
    output_uri: str
    parent_node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    phase: _execution_pb2.TaskExecution.Phase
    phase_version: int
    producer_id: str
    reason: str
    retry_attempt: int
    task_id: _identifier_pb2.Identifier
    task_type: str
    def __init__(self, task_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., parent_node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., producer_id: _Optional[str] = ..., logs: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., input_uri: _Optional[str] = ..., input_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., phase_version: _Optional[int] = ..., reason: _Optional[str] = ..., task_type: _Optional[str] = ..., metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ..., event_version: _Optional[int] = ...) -> None: ...

class TaskExecutionMetadata(_message.Message):
    __slots__ = ["external_resources", "generated_name", "instance_class", "plugin_identifier", "resource_pool_info"]
    class InstanceClass(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    DEFAULT: TaskExecutionMetadata.InstanceClass
    EXTERNAL_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    GENERATED_NAME_FIELD_NUMBER: _ClassVar[int]
    INSTANCE_CLASS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE: TaskExecutionMetadata.InstanceClass
    PLUGIN_IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_POOL_INFO_FIELD_NUMBER: _ClassVar[int]
    external_resources: _containers.RepeatedCompositeFieldContainer[ExternalResourceInfo]
    generated_name: str
    instance_class: TaskExecutionMetadata.InstanceClass
    plugin_identifier: str
    resource_pool_info: _containers.RepeatedCompositeFieldContainer[ResourcePoolInfo]
    def __init__(self, generated_name: _Optional[str] = ..., external_resources: _Optional[_Iterable[_Union[ExternalResourceInfo, _Mapping]]] = ..., resource_pool_info: _Optional[_Iterable[_Union[ResourcePoolInfo, _Mapping]]] = ..., plugin_identifier: _Optional[str] = ..., instance_class: _Optional[_Union[TaskExecutionMetadata.InstanceClass, str]] = ...) -> None: ...

class TaskNodeMetadata(_message.Message):
    __slots__ = ["cache_status", "catalog_key", "checkpoint_uri", "dynamic_workflow", "reservation_status"]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    CATALOG_KEY_FIELD_NUMBER: _ClassVar[int]
    CHECKPOINT_URI_FIELD_NUMBER: _ClassVar[int]
    DYNAMIC_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_STATUS_FIELD_NUMBER: _ClassVar[int]
    cache_status: _catalog_pb2.CatalogCacheStatus
    catalog_key: _catalog_pb2.CatalogMetadata
    checkpoint_uri: str
    dynamic_workflow: DynamicWorkflowNodeMetadata
    reservation_status: _catalog_pb2.CatalogReservation.Status
    def __init__(self, cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., catalog_key: _Optional[_Union[_catalog_pb2.CatalogMetadata, _Mapping]] = ..., reservation_status: _Optional[_Union[_catalog_pb2.CatalogReservation.Status, str]] = ..., checkpoint_uri: _Optional[str] = ..., dynamic_workflow: _Optional[_Union[DynamicWorkflowNodeMetadata, _Mapping]] = ...) -> None: ...

class WorkflowExecutionEvent(_message.Message):
    __slots__ = ["error", "execution_id", "occurred_at", "output_data", "output_uri", "phase", "producer_id"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_ID_FIELD_NUMBER: _ClassVar[int]
    error: _execution_pb2.ExecutionError
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    occurred_at: _timestamp_pb2.Timestamp
    output_data: _literals_pb2.LiteralMap
    output_uri: str
    phase: _execution_pb2.WorkflowExecution.Phase
    producer_id: str
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., producer_id: _Optional[str] = ..., phase: _Optional[_Union[_execution_pb2.WorkflowExecution.Phase, str]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class WorkflowNodeMetadata(_message.Message):
    __slots__ = ["execution_id"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...
