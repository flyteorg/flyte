from flyteidl.admin import cluster_assignment_pb2 as _cluster_assignment_pb2
from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import metrics_pb2 as _metrics_pb2
from flyteidl.core import security_pb2 as _security_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ExecutionState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    EXECUTION_ACTIVE: _ClassVar[ExecutionState]
    EXECUTION_ARCHIVED: _ClassVar[ExecutionState]
EXECUTION_ACTIVE: ExecutionState
EXECUTION_ARCHIVED: ExecutionState

class ExecutionCreateRequest(_message.Message):
    __slots__ = ["project", "domain", "name", "spec", "inputs"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    name: str
    spec: ExecutionSpec
    inputs: _literals_pb2.LiteralMap
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., spec: _Optional[_Union[ExecutionSpec, _Mapping]] = ..., inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class ExecutionRelaunchRequest(_message.Message):
    __slots__ = ["id", "name", "overwrite_cache"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    name: str
    overwrite_cache: bool
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., name: _Optional[str] = ..., overwrite_cache: bool = ...) -> None: ...

class ExecutionRecoverRequest(_message.Message):
    __slots__ = ["id", "name", "metadata"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    name: str
    metadata: ExecutionMetadata
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., name: _Optional[str] = ..., metadata: _Optional[_Union[ExecutionMetadata, _Mapping]] = ...) -> None: ...

class ExecutionCreateResponse(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class WorkflowExecutionGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class Execution(_message.Message):
    __slots__ = ["id", "spec", "closure"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    spec: ExecutionSpec
    closure: ExecutionClosure
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., spec: _Optional[_Union[ExecutionSpec, _Mapping]] = ..., closure: _Optional[_Union[ExecutionClosure, _Mapping]] = ...) -> None: ...

class ExecutionList(_message.Message):
    __slots__ = ["executions", "token"]
    EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    executions: _containers.RepeatedCompositeFieldContainer[Execution]
    token: str
    def __init__(self, executions: _Optional[_Iterable[_Union[Execution, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class LiteralMapBlob(_message.Message):
    __slots__ = ["values", "uri"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    values: _literals_pb2.LiteralMap
    uri: str
    def __init__(self, values: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., uri: _Optional[str] = ...) -> None: ...

class AbortMetadata(_message.Message):
    __slots__ = ["cause", "principal"]
    CAUSE_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    cause: str
    principal: str
    def __init__(self, cause: _Optional[str] = ..., principal: _Optional[str] = ...) -> None: ...

class ExecutionClosure(_message.Message):
    __slots__ = ["outputs", "error", "abort_cause", "abort_metadata", "output_data", "computed_inputs", "phase", "started_at", "duration", "created_at", "updated_at", "notifications", "workflow_id", "state_change_details"]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    ABORT_CAUSE_FIELD_NUMBER: _ClassVar[int]
    ABORT_METADATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    COMPUTED_INPUTS_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATIONS_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_CHANGE_DETAILS_FIELD_NUMBER: _ClassVar[int]
    outputs: LiteralMapBlob
    error: _execution_pb2.ExecutionError
    abort_cause: str
    abort_metadata: AbortMetadata
    output_data: _literals_pb2.LiteralMap
    computed_inputs: _literals_pb2.LiteralMap
    phase: _execution_pb2.WorkflowExecution.Phase
    started_at: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    notifications: _containers.RepeatedCompositeFieldContainer[_common_pb2.Notification]
    workflow_id: _identifier_pb2.Identifier
    state_change_details: ExecutionStateChangeDetails
    def __init__(self, outputs: _Optional[_Union[LiteralMapBlob, _Mapping]] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ..., abort_cause: _Optional[str] = ..., abort_metadata: _Optional[_Union[AbortMetadata, _Mapping]] = ..., output_data: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., computed_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., phase: _Optional[_Union[_execution_pb2.WorkflowExecution.Phase, str]] = ..., started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., notifications: _Optional[_Iterable[_Union[_common_pb2.Notification, _Mapping]]] = ..., workflow_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., state_change_details: _Optional[_Union[ExecutionStateChangeDetails, _Mapping]] = ...) -> None: ...

class SystemMetadata(_message.Message):
    __slots__ = ["execution_cluster", "namespace"]
    EXECUTION_CLUSTER_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    execution_cluster: str
    namespace: str
    def __init__(self, execution_cluster: _Optional[str] = ..., namespace: _Optional[str] = ...) -> None: ...

class ExecutionMetadata(_message.Message):
    __slots__ = ["mode", "principal", "nesting", "scheduled_at", "parent_node_execution", "reference_execution", "system_metadata"]
    class ExecutionMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        MANUAL: _ClassVar[ExecutionMetadata.ExecutionMode]
        SCHEDULED: _ClassVar[ExecutionMetadata.ExecutionMode]
        SYSTEM: _ClassVar[ExecutionMetadata.ExecutionMode]
        RELAUNCH: _ClassVar[ExecutionMetadata.ExecutionMode]
        CHILD_WORKFLOW: _ClassVar[ExecutionMetadata.ExecutionMode]
        RECOVERED: _ClassVar[ExecutionMetadata.ExecutionMode]
    MANUAL: ExecutionMetadata.ExecutionMode
    SCHEDULED: ExecutionMetadata.ExecutionMode
    SYSTEM: ExecutionMetadata.ExecutionMode
    RELAUNCH: ExecutionMetadata.ExecutionMode
    CHILD_WORKFLOW: ExecutionMetadata.ExecutionMode
    RECOVERED: ExecutionMetadata.ExecutionMode
    MODE_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    NESTING_FIELD_NUMBER: _ClassVar[int]
    SCHEDULED_AT_FIELD_NUMBER: _ClassVar[int]
    PARENT_NODE_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    REFERENCE_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    SYSTEM_METADATA_FIELD_NUMBER: _ClassVar[int]
    mode: ExecutionMetadata.ExecutionMode
    principal: str
    nesting: int
    scheduled_at: _timestamp_pb2.Timestamp
    parent_node_execution: _identifier_pb2.NodeExecutionIdentifier
    reference_execution: _identifier_pb2.WorkflowExecutionIdentifier
    system_metadata: SystemMetadata
    def __init__(self, mode: _Optional[_Union[ExecutionMetadata.ExecutionMode, str]] = ..., principal: _Optional[str] = ..., nesting: _Optional[int] = ..., scheduled_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., parent_node_execution: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., reference_execution: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., system_metadata: _Optional[_Union[SystemMetadata, _Mapping]] = ...) -> None: ...

class NotificationList(_message.Message):
    __slots__ = ["notifications"]
    NOTIFICATIONS_FIELD_NUMBER: _ClassVar[int]
    notifications: _containers.RepeatedCompositeFieldContainer[_common_pb2.Notification]
    def __init__(self, notifications: _Optional[_Iterable[_Union[_common_pb2.Notification, _Mapping]]] = ...) -> None: ...

class ExecutionSpec(_message.Message):
    __slots__ = ["launch_plan", "inputs", "metadata", "notifications", "disable_all", "labels", "annotations", "security_context", "auth_role", "quality_of_service", "max_parallelism", "raw_output_data_config", "cluster_assignment", "interruptible", "overwrite_cache", "envs"]
    LAUNCH_PLAN_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATIONS_FIELD_NUMBER: _ClassVar[int]
    DISABLE_ALL_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    AUTH_ROLE_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    MAX_PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    RAW_OUTPUT_DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ASSIGNMENT_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    launch_plan: _identifier_pb2.Identifier
    inputs: _literals_pb2.LiteralMap
    metadata: ExecutionMetadata
    notifications: NotificationList
    disable_all: bool
    labels: _common_pb2.Labels
    annotations: _common_pb2.Annotations
    security_context: _security_pb2.SecurityContext
    auth_role: _common_pb2.AuthRole
    quality_of_service: _execution_pb2.QualityOfService
    max_parallelism: int
    raw_output_data_config: _common_pb2.RawOutputDataConfig
    cluster_assignment: _cluster_assignment_pb2.ClusterAssignment
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    envs: _common_pb2.Envs
    def __init__(self, launch_plan: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., metadata: _Optional[_Union[ExecutionMetadata, _Mapping]] = ..., notifications: _Optional[_Union[NotificationList, _Mapping]] = ..., disable_all: bool = ..., labels: _Optional[_Union[_common_pb2.Labels, _Mapping]] = ..., annotations: _Optional[_Union[_common_pb2.Annotations, _Mapping]] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., auth_role: _Optional[_Union[_common_pb2.AuthRole, _Mapping]] = ..., quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., max_parallelism: _Optional[int] = ..., raw_output_data_config: _Optional[_Union[_common_pb2.RawOutputDataConfig, _Mapping]] = ..., cluster_assignment: _Optional[_Union[_cluster_assignment_pb2.ClusterAssignment, _Mapping]] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., envs: _Optional[_Union[_common_pb2.Envs, _Mapping]] = ...) -> None: ...

class ExecutionTerminateRequest(_message.Message):
    __slots__ = ["id", "cause"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CAUSE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    cause: str
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., cause: _Optional[str] = ...) -> None: ...

class ExecutionTerminateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class WorkflowExecutionGetDataRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class WorkflowExecutionGetDataResponse(_message.Message):
    __slots__ = ["outputs", "inputs", "full_inputs", "full_outputs"]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_INPUTS_FIELD_NUMBER: _ClassVar[int]
    FULL_OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    outputs: _common_pb2.UrlBlob
    inputs: _common_pb2.UrlBlob
    full_inputs: _literals_pb2.LiteralMap
    full_outputs: _literals_pb2.LiteralMap
    def __init__(self, outputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., inputs: _Optional[_Union[_common_pb2.UrlBlob, _Mapping]] = ..., full_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., full_outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class ExecutionUpdateRequest(_message.Message):
    __slots__ = ["id", "state"]
    ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    state: ExecutionState
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., state: _Optional[_Union[ExecutionState, str]] = ...) -> None: ...

class ExecutionStateChangeDetails(_message.Message):
    __slots__ = ["state", "occurred_at", "principal"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    PRINCIPAL_FIELD_NUMBER: _ClassVar[int]
    state: ExecutionState
    occurred_at: _timestamp_pb2.Timestamp
    principal: str
    def __init__(self, state: _Optional[_Union[ExecutionState, str]] = ..., occurred_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., principal: _Optional[str] = ...) -> None: ...

class ExecutionUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class WorkflowExecutionGetMetricsRequest(_message.Message):
    __slots__ = ["id", "depth"]
    ID_FIELD_NUMBER: _ClassVar[int]
    DEPTH_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    depth: int
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., depth: _Optional[int] = ...) -> None: ...

class WorkflowExecutionGetMetricsResponse(_message.Message):
    __slots__ = ["span"]
    SPAN_FIELD_NUMBER: _ClassVar[int]
    span: _metrics_pb2.Span
    def __init__(self, span: _Optional[_Union[_metrics_pb2.Span, _Mapping]] = ...) -> None: ...
