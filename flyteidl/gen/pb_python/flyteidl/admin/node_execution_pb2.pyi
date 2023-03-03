from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import catalog_pb2 as _catalog_pb2
from flyteidl.core import compiler_pb2 as _compiler_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import containers as _containers
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

class NodeExecution(_message.Message):
    __slots__ = ["closure", "id", "input_uri", "metadata"]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INPUT_URI_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    closure: NodeExecutionClosure
    id: _identifier_pb2.NodeExecutionIdentifier
    input_uri: str
    metadata: NodeExecutionMetaData
    def __init__(self, id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., input_uri: _Optional[str] = ..., closure: _Optional[_Union[NodeExecutionClosure, _Mapping]] = ..., metadata: _Optional[_Union[NodeExecutionMetaData, _Mapping]] = ...) -> None: ...

class NodeExecutionClosure(_message.Message):
    __slots__ = ["created_at", "deck_uri", "duration", "dynamic_job_spec_uri", "error", "output_data", "output_uri", "phase", "started_at", "task_node_metadata", "updated_at", "workflow_node_metadata"]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    DECK_URI_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    DYNAMIC_JOB_SPEC_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    TASK_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_NODE_METADATA_FIELD_NUMBER: _ClassVar[int]
    created_at: _timestamp_pb2.Timestamp
    deck_uri: str
    duration: _duration_pb2.Duration
    dynamic_job_spec_uri: str
    error: _execution_pb2.ExecutionError
    output_data: _literals_pb2.LiteralMap
    output_uri: str
    phase: _execution_pb2.NodeExecution.Phase
    started_at: _timestamp_pb2.Timestamp
    task_node_metadata: TaskNodeMetadata
    updated_at: _timestamp_pb2.Timestamp
    workflow_node_metadata: WorkflowNodeMetadata
    def __init__(self, output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., phase: _Optional[_Union[_execution_pb2.NodeExecution.Phase, str]] = ..., started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., workflow_node_metadata: _Optional[_Union[WorkflowNodeMetadata, _Mapping]] = ..., task_node_metadata: _Optional[_Union[TaskNodeMetadata, _Mapping]] = ..., deck_uri: _Optional[str] = ..., dynamic_job_spec_uri: _Optional[str] = ...) -> None: ...

class NodeExecutionForTaskListRequest(_message.Message):
    __slots__ = ["filters", "limit", "sort_by", "task_execution_id", "token"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    filters: str
    limit: int
    sort_by: _common_pb2.Sort
    task_execution_id: _identifier_pb2.TaskExecutionIdentifier
    token: str
    def __init__(self, task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...

class NodeExecutionGetDataRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.NodeExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ...) -> None: ...

class NodeExecutionGetDataResponse(_message.Message):
    __slots__ = ["dynamic_workflow", "full_inputs", "full_outputs", "inputs", "outputs"]
    DYNAMIC_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    FULL_INPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    dynamic_workflow: DynamicWorkflowNodeMetadata
    full_inputs: _literals_pb2.LiteralMap
    full_outputs: _literals_pb2.LiteralMap
    inputs: _common_pb2.UrlBlob
    outputs: _common_pb2.UrlBlob
    def __init__(self, inputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., full_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., full_outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., dynamic_workflow: _Optional[_Union[DynamicWorkflowNodeMetadata, _Mapping]] = ...) -> None: ...

class NodeExecutionGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.NodeExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ...) -> None: ...

class NodeExecutionList(_message.Message):
    __slots__ = ["node_executions", "token"]
    NODE_EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    node_executions: _containers.RepeatedCompositeFieldContainer[NodeExecution]
    token: str
    def __init__(self, node_executions: _Optional[_Iterable[_Union[NodeExecution, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class NodeExecutionListRequest(_message.Message):
    __slots__ = ["filters", "limit", "sort_by", "token", "unique_parent_id", "workflow_execution_id"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    UNIQUE_PARENT_ID_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    filters: str
    limit: int
    sort_by: _common_pb2.Sort
    token: str
    unique_parent_id: str
    workflow_execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, workflow_execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ..., unique_parent_id: _Optional[str] = ...) -> None: ...

class NodeExecutionMetaData(_message.Message):
    __slots__ = ["is_dynamic", "is_parent_node", "retry_group", "spec_node_id"]
    IS_DYNAMIC_FIELD_NUMBER: _ClassVar[int]
    IS_PARENT_NODE_FIELD_NUMBER: _ClassVar[int]
    RETRY_GROUP_FIELD_NUMBER: _ClassVar[int]
    SPEC_NODE_ID_FIELD_NUMBER: _ClassVar[int]
    is_dynamic: bool
    is_parent_node: bool
    retry_group: str
    spec_node_id: str
    def __init__(self, retry_group: _Optional[str] = ..., is_parent_node: bool = ..., spec_node_id: _Optional[str] = ..., is_dynamic: bool = ...) -> None: ...

class TaskNodeMetadata(_message.Message):
    __slots__ = ["cache_status", "catalog_key", "checkpoint_uri"]
    CACHE_STATUS_FIELD_NUMBER: _ClassVar[int]
    CATALOG_KEY_FIELD_NUMBER: _ClassVar[int]
    CHECKPOINT_URI_FIELD_NUMBER: _ClassVar[int]
    cache_status: _catalog_pb2.CatalogCacheStatus
    catalog_key: _catalog_pb2.CatalogMetadata
    checkpoint_uri: str
    def __init__(self, cache_status: _Optional[_Union[_catalog_pb2.CatalogCacheStatus, str]] = ..., catalog_key: _Optional[_Union[_catalog_pb2.CatalogMetadata, _Mapping]] = ..., checkpoint_uri: _Optional[str] = ...) -> None: ...

class WorkflowNodeMetadata(_message.Message):
    __slots__ = ["executionId"]
    EXECUTIONID_FIELD_NUMBER: _ClassVar[int]
    executionId: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, executionId: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...
