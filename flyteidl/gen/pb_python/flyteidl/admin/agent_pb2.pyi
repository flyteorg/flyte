from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.core import workflow_pb2 as _workflow_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import metrics_pb2 as _metrics_pb2
from flyteidl.core import security_pb2 as _security_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class State(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    RETRYABLE_FAILURE: _ClassVar[State]
    PERMANENT_FAILURE: _ClassVar[State]
    PENDING: _ClassVar[State]
    RUNNING: _ClassVar[State]
    SUCCEEDED: _ClassVar[State]
RETRYABLE_FAILURE: State
PERMANENT_FAILURE: State
PENDING: State
RUNNING: State
SUCCEEDED: State

class TaskExecutionMetadata(_message.Message):
    __slots__ = ["task_execution_id", "namespace", "labels", "annotations", "k8s_service_account", "environment_variables", "max_attempts", "interruptible", "interruptible_failure_threshold", "overrides", "identity"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class AnnotationsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class EnvironmentVariablesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    TASK_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    K8S_SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_VARIABLES_FIELD_NUMBER: _ClassVar[int]
    MAX_ATTEMPTS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FAILURE_THRESHOLD_FIELD_NUMBER: _ClassVar[int]
    OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    IDENTITY_FIELD_NUMBER: _ClassVar[int]
    task_execution_id: _identifier_pb2.TaskExecutionIdentifier
    namespace: str
    labels: _containers.ScalarMap[str, str]
    annotations: _containers.ScalarMap[str, str]
    k8s_service_account: str
    environment_variables: _containers.ScalarMap[str, str]
    max_attempts: int
    interruptible: bool
    interruptible_failure_threshold: int
    overrides: _workflow_pb2.TaskNodeOverrides
    identity: _security_pb2.Identity
    def __init__(self, task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., namespace: _Optional[str] = ..., labels: _Optional[_Mapping[str, str]] = ..., annotations: _Optional[_Mapping[str, str]] = ..., k8s_service_account: _Optional[str] = ..., environment_variables: _Optional[_Mapping[str, str]] = ..., max_attempts: _Optional[int] = ..., interruptible: bool = ..., interruptible_failure_threshold: _Optional[int] = ..., overrides: _Optional[_Union[_workflow_pb2.TaskNodeOverrides, _Mapping]] = ..., identity: _Optional[_Union[_security_pb2.Identity, _Mapping]] = ...) -> None: ...

class CreateTaskRequest(_message.Message):
    __slots__ = ["inputs", "template", "output_prefix", "task_execution_metadata"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_METADATA_FIELD_NUMBER: _ClassVar[int]
    inputs: _literals_pb2.LiteralMap
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    task_execution_metadata: TaskExecutionMetadata
    def __init__(self, inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ..., task_execution_metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ...) -> None: ...

class CreateTaskResponse(_message.Message):
    __slots__ = ["resource_meta"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    def __init__(self, resource_meta: _Optional[bytes] = ...) -> None: ...

class CreateRequestHeader(_message.Message):
    __slots__ = ["template", "output_prefix", "task_execution_metadata", "max_dataset_size_bytes"]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_METADATA_FIELD_NUMBER: _ClassVar[int]
    MAX_DATASET_SIZE_BYTES_FIELD_NUMBER: _ClassVar[int]
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    task_execution_metadata: TaskExecutionMetadata
    max_dataset_size_bytes: int
    def __init__(self, template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ..., task_execution_metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ..., max_dataset_size_bytes: _Optional[int] = ...) -> None: ...

class ExecuteTaskSyncRequest(_message.Message):
    __slots__ = ["header", "inputs"]
    HEADER_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    header: CreateRequestHeader
    inputs: _literals_pb2.LiteralMap
    def __init__(self, header: _Optional[_Union[CreateRequestHeader, _Mapping]] = ..., inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class ExecuteTaskSyncResponseHeader(_message.Message):
    __slots__ = ["resource"]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    resource: Resource
    def __init__(self, resource: _Optional[_Union[Resource, _Mapping]] = ...) -> None: ...

class ExecuteTaskSyncResponse(_message.Message):
    __slots__ = ["header", "outputs"]
    HEADER_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    header: ExecuteTaskSyncResponseHeader
    outputs: _literals_pb2.LiteralMap
    def __init__(self, header: _Optional[_Union[ExecuteTaskSyncResponseHeader, _Mapping]] = ..., outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class GetTaskRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta", "task_category"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    task_category: TaskCategory
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

class GetTaskResponse(_message.Message):
    __slots__ = ["resource"]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    resource: Resource
    def __init__(self, resource: _Optional[_Union[Resource, _Mapping]] = ...) -> None: ...

class Resource(_message.Message):
    __slots__ = ["state", "outputs", "message", "log_links", "phase", "custom_info"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    LOG_LINKS_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    state: State
    outputs: _literals_pb2.LiteralMap
    message: str
    log_links: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    phase: _execution_pb2.TaskExecution.Phase
    custom_info: _struct_pb2.Struct
    def __init__(self, state: _Optional[_Union[State, str]] = ..., outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., message: _Optional[str] = ..., log_links: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class DeleteTaskRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta", "task_category"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    task_category: TaskCategory
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

class DeleteTaskResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Agent(_message.Message):
    __slots__ = ["name", "supported_task_types", "is_sync", "supported_task_categories"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SUPPORTED_TASK_TYPES_FIELD_NUMBER: _ClassVar[int]
    IS_SYNC_FIELD_NUMBER: _ClassVar[int]
    SUPPORTED_TASK_CATEGORIES_FIELD_NUMBER: _ClassVar[int]
    name: str
    supported_task_types: _containers.RepeatedScalarFieldContainer[str]
    is_sync: bool
    supported_task_categories: _containers.RepeatedCompositeFieldContainer[TaskCategory]
    def __init__(self, name: _Optional[str] = ..., supported_task_types: _Optional[_Iterable[str]] = ..., is_sync: bool = ..., supported_task_categories: _Optional[_Iterable[_Union[TaskCategory, _Mapping]]] = ...) -> None: ...

class TaskCategory(_message.Message):
    __slots__ = ["name", "version"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: str
    version: int
    def __init__(self, name: _Optional[str] = ..., version: _Optional[int] = ...) -> None: ...

class GetAgentRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class GetAgentResponse(_message.Message):
    __slots__ = ["agent"]
    AGENT_FIELD_NUMBER: _ClassVar[int]
    agent: Agent
    def __init__(self, agent: _Optional[_Union[Agent, _Mapping]] = ...) -> None: ...

class ListAgentsRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListAgentsResponse(_message.Message):
    __slots__ = ["agents"]
    AGENTS_FIELD_NUMBER: _ClassVar[int]
    agents: _containers.RepeatedCompositeFieldContainer[Agent]
    def __init__(self, agents: _Optional[_Iterable[_Union[Agent, _Mapping]]] = ...) -> None: ...

class GetTaskMetricsRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta", "queries", "start_time", "end_time", "step", "task_category"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    QUERIES_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    STEP_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    queries: _containers.RepeatedScalarFieldContainer[str]
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    step: _duration_pb2.Duration
    task_category: TaskCategory
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ..., queries: _Optional[_Iterable[str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., step: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

class GetTaskMetricsResponse(_message.Message):
    __slots__ = ["results"]
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[_metrics_pb2.ExecutionMetricResult]
    def __init__(self, results: _Optional[_Iterable[_Union[_metrics_pb2.ExecutionMetricResult, _Mapping]]] = ...) -> None: ...

class GetTaskLogsRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta", "lines", "token", "task_category"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    lines: int
    token: str
    task_category: TaskCategory
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ..., lines: _Optional[int] = ..., token: _Optional[str] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

class GetTaskLogsResponseHeader(_message.Message):
    __slots__ = ["token"]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    token: str
    def __init__(self, token: _Optional[str] = ...) -> None: ...

class GetTaskLogsResponseBody(_message.Message):
    __slots__ = ["results"]
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, results: _Optional[_Iterable[str]] = ...) -> None: ...

class GetTaskLogsResponse(_message.Message):
    __slots__ = ["header", "body"]
    HEADER_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    header: GetTaskLogsResponseHeader
    body: GetTaskLogsResponseBody
    def __init__(self, header: _Optional[_Union[GetTaskLogsResponseHeader, _Mapping]] = ..., body: _Optional[_Union[GetTaskLogsResponseBody, _Mapping]] = ...) -> None: ...
