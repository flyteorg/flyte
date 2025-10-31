from flyteidl2.core import execution_pb2 as _execution_pb2
from flyteidl2.core import identifier_pb2 as _identifier_pb2
from flyteidl2.core import metrics_pb2 as _metrics_pb2
from flyteidl2.core import security_pb2 as _security_pb2
from flyteidl2.core import tasks_pb2 as _tasks_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TaskExecutionMetadata(_message.Message):
    __slots__ = ["task_execution_id", "namespace", "labels", "annotations", "k8s_service_account", "environment_variables", "max_attempts", "interruptible", "interruptible_failure_threshold", "identity"]
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
    identity: _security_pb2.Identity
    def __init__(self, task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., namespace: _Optional[str] = ..., labels: _Optional[_Mapping[str, str]] = ..., annotations: _Optional[_Mapping[str, str]] = ..., k8s_service_account: _Optional[str] = ..., environment_variables: _Optional[_Mapping[str, str]] = ..., max_attempts: _Optional[int] = ..., interruptible: bool = ..., interruptible_failure_threshold: _Optional[int] = ..., identity: _Optional[_Union[_security_pb2.Identity, _Mapping]] = ...) -> None: ...

class CreateTaskRequest(_message.Message):
    __slots__ = ["inputs", "template", "output_prefix", "task_execution_metadata", "connection"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_METADATA_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    inputs: _common_pb2.Inputs
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    task_execution_metadata: TaskExecutionMetadata
    connection: _security_pb2.Connection
    def __init__(self, inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ..., task_execution_metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ..., connection: _Optional[_Union[_security_pb2.Connection, _Mapping]] = ...) -> None: ...

class CreateTaskResponse(_message.Message):
    __slots__ = ["resource_meta"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    def __init__(self, resource_meta: _Optional[bytes] = ...) -> None: ...

class CreateRequestHeader(_message.Message):
    __slots__ = ["template", "output_prefix", "task_execution_metadata", "max_dataset_size_bytes", "connection"]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_METADATA_FIELD_NUMBER: _ClassVar[int]
    MAX_DATASET_SIZE_BYTES_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    task_execution_metadata: TaskExecutionMetadata
    max_dataset_size_bytes: int
    connection: _security_pb2.Connection
    def __init__(self, template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ..., task_execution_metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ..., max_dataset_size_bytes: _Optional[int] = ..., connection: _Optional[_Union[_security_pb2.Connection, _Mapping]] = ...) -> None: ...

class GetTaskRequest(_message.Message):
    __slots__ = ["resource_meta", "task_category", "output_prefix", "connection"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    task_category: TaskCategory
    output_prefix: str
    connection: _security_pb2.Connection
    def __init__(self, resource_meta: _Optional[bytes] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ..., output_prefix: _Optional[str] = ..., connection: _Optional[_Union[_security_pb2.Connection, _Mapping]] = ...) -> None: ...

class GetTaskResponse(_message.Message):
    __slots__ = ["resource"]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    resource: Resource
    def __init__(self, resource: _Optional[_Union[Resource, _Mapping]] = ...) -> None: ...

class Resource(_message.Message):
    __slots__ = ["outputs", "message", "log_links", "phase", "custom_info"]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    LOG_LINKS_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_INFO_FIELD_NUMBER: _ClassVar[int]
    outputs: _common_pb2.Outputs
    message: str
    log_links: _containers.RepeatedCompositeFieldContainer[_execution_pb2.TaskLog]
    phase: _execution_pb2.TaskExecution.Phase
    custom_info: _struct_pb2.Struct
    def __init__(self, outputs: _Optional[_Union[_common_pb2.Outputs, _Mapping]] = ..., message: _Optional[str] = ..., log_links: _Optional[_Iterable[_Union[_execution_pb2.TaskLog, _Mapping]]] = ..., phase: _Optional[_Union[_execution_pb2.TaskExecution.Phase, str]] = ..., custom_info: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class DeleteTaskRequest(_message.Message):
    __slots__ = ["resource_meta", "task_category", "connection"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    task_category: TaskCategory
    connection: _security_pb2.Connection
    def __init__(self, resource_meta: _Optional[bytes] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ..., connection: _Optional[_Union[_security_pb2.Connection, _Mapping]] = ...) -> None: ...

class DeleteTaskResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Connector(_message.Message):
    __slots__ = ["name", "supported_task_categories"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SUPPORTED_TASK_CATEGORIES_FIELD_NUMBER: _ClassVar[int]
    name: str
    supported_task_categories: _containers.RepeatedCompositeFieldContainer[TaskCategory]
    def __init__(self, name: _Optional[str] = ..., supported_task_categories: _Optional[_Iterable[_Union[TaskCategory, _Mapping]]] = ...) -> None: ...

class TaskCategory(_message.Message):
    __slots__ = ["name", "version"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: str
    version: int
    def __init__(self, name: _Optional[str] = ..., version: _Optional[int] = ...) -> None: ...

class GetConnectorRequest(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class GetConnectorResponse(_message.Message):
    __slots__ = ["connector"]
    CONNECTOR_FIELD_NUMBER: _ClassVar[int]
    connector: Connector
    def __init__(self, connector: _Optional[_Union[Connector, _Mapping]] = ...) -> None: ...

class ListConnectorsRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListConnectorsResponse(_message.Message):
    __slots__ = ["connectors"]
    CONNECTORS_FIELD_NUMBER: _ClassVar[int]
    connectors: _containers.RepeatedCompositeFieldContainer[Connector]
    def __init__(self, connectors: _Optional[_Iterable[_Union[Connector, _Mapping]]] = ...) -> None: ...

class GetTaskMetricsRequest(_message.Message):
    __slots__ = ["resource_meta", "queries", "start_time", "end_time", "step", "task_category"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    QUERIES_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    STEP_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    queries: _containers.RepeatedScalarFieldContainer[str]
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    step: _duration_pb2.Duration
    task_category: TaskCategory
    def __init__(self, resource_meta: _Optional[bytes] = ..., queries: _Optional[_Iterable[str]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., step: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

class GetTaskMetricsResponse(_message.Message):
    __slots__ = ["results"]
    RESULTS_FIELD_NUMBER: _ClassVar[int]
    results: _containers.RepeatedCompositeFieldContainer[_metrics_pb2.ExecutionMetricResult]
    def __init__(self, results: _Optional[_Iterable[_Union[_metrics_pb2.ExecutionMetricResult, _Mapping]]] = ...) -> None: ...

class GetTaskLogsRequest(_message.Message):
    __slots__ = ["resource_meta", "lines", "token", "task_category"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    TASK_CATEGORY_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    lines: int
    token: str
    task_category: TaskCategory
    def __init__(self, resource_meta: _Optional[bytes] = ..., lines: _Optional[int] = ..., token: _Optional[str] = ..., task_category: _Optional[_Union[TaskCategory, _Mapping]] = ...) -> None: ...

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
