from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.core import literals_pb2 as _literals_pb2
from flyteidl2.core import security_pb2 as _security_pb2
from flyteidl2.core import tasks_pb2 as _tasks_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EnvironmentClusterState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ENVIRONMENT_CLUSTER_STATE_UNSPECIFIED: _ClassVar[EnvironmentClusterState]
    ENVIRONMENT_CLUSTER_STATE_NOT_READY: _ClassVar[EnvironmentClusterState]
    ENVIRONMENT_CLUSTER_STATE_READY: _ClassVar[EnvironmentClusterState]
    ENVIRONMENT_CLUSTER_STATE_FAILED: _ClassVar[EnvironmentClusterState]

class EnvironmentState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ENVIRONMENT_STATE_UNSPECIFIED: _ClassVar[EnvironmentState]
    ENVIRONMENT_STATE_NOT_READY: _ClassVar[EnvironmentState]
    ENVIRONMENT_STATE_READY: _ClassVar[EnvironmentState]
    ENVIRONMENT_STATE_FAILED: _ClassVar[EnvironmentState]
    ENVIRONMENT_STATE_INACTIVE: _ClassVar[EnvironmentState]

class WorkerState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    WORKER_STATE_UNSPECIFIED: _ClassVar[WorkerState]
    WORKER_STATE_INITIALIZING: _ClassVar[WorkerState]
    WORKER_STATE_READY: _ClassVar[WorkerState]
    WORKER_STATE_DISCONNECTED: _ClassVar[WorkerState]
    WORKER_STATE_FAILED: _ClassVar[WorkerState]

class EnvironmentActionAssignmentState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ENVIRONMENT_ACTION_ASSIGNMENT_STATE_UNSPECIFIED: _ClassVar[EnvironmentActionAssignmentState]
    ENVIRONMENT_ACTION_ASSIGNMENT_STATE_PENDING: _ClassVar[EnvironmentActionAssignmentState]
    ENVIRONMENT_ACTION_ASSIGNMENT_STATE_ASSIGNED: _ClassVar[EnvironmentActionAssignmentState]
ENVIRONMENT_CLUSTER_STATE_UNSPECIFIED: EnvironmentClusterState
ENVIRONMENT_CLUSTER_STATE_NOT_READY: EnvironmentClusterState
ENVIRONMENT_CLUSTER_STATE_READY: EnvironmentClusterState
ENVIRONMENT_CLUSTER_STATE_FAILED: EnvironmentClusterState
ENVIRONMENT_STATE_UNSPECIFIED: EnvironmentState
ENVIRONMENT_STATE_NOT_READY: EnvironmentState
ENVIRONMENT_STATE_READY: EnvironmentState
ENVIRONMENT_STATE_FAILED: EnvironmentState
ENVIRONMENT_STATE_INACTIVE: EnvironmentState
WORKER_STATE_UNSPECIFIED: WorkerState
WORKER_STATE_INITIALIZING: WorkerState
WORKER_STATE_READY: WorkerState
WORKER_STATE_DISCONNECTED: WorkerState
WORKER_STATE_FAILED: WorkerState
ENVIRONMENT_ACTION_ASSIGNMENT_STATE_UNSPECIFIED: EnvironmentActionAssignmentState
ENVIRONMENT_ACTION_ASSIGNMENT_STATE_PENDING: EnvironmentActionAssignmentState
ENVIRONMENT_ACTION_ASSIGNMENT_STATE_ASSIGNED: EnvironmentActionAssignmentState

class EnvironmentId(_message.Message):
    __slots__ = ["organization", "project", "domain", "name", "version"]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    organization: str
    project: str
    domain: str
    name: str
    version: str
    def __init__(self, organization: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class EnvironmentClusterId(_message.Message):
    __slots__ = ["environment", "cluster"]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    environment: EnvironmentId
    cluster: str
    def __init__(self, environment: _Optional[_Union[EnvironmentId, _Mapping]] = ..., cluster: _Optional[str] = ...) -> None: ...

class EnvironmentSpec(_message.Message):
    __slots__ = ["image", "resources", "extended_resources", "security_context", "env_vars", "interruptible", "pod_template_name", "parallelism", "pod_template"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    EXTENDED_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    ENV_VARS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    POD_TEMPLATE_NAME_FIELD_NUMBER: _ClassVar[int]
    PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    POD_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    image: str
    resources: _tasks_pb2.Resources
    extended_resources: _tasks_pb2.ExtendedResources
    security_context: _security_pb2.SecurityContext
    env_vars: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    interruptible: bool
    pod_template_name: str
    parallelism: int
    pod_template: _tasks_pb2.K8sPod
    def __init__(self, image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ..., extended_resources: _Optional[_Union[_tasks_pb2.ExtendedResources, _Mapping]] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., env_vars: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ..., interruptible: bool = ..., pod_template_name: _Optional[str] = ..., parallelism: _Optional[int] = ..., pod_template: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ...) -> None: ...

class EnvironmentScaling(_message.Message):
    __slots__ = ["replicas", "min_replicas", "idle_ttl", "scaledown_ttl"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    IDLE_TTL_FIELD_NUMBER: _ClassVar[int]
    SCALEDOWN_TTL_FIELD_NUMBER: _ClassVar[int]
    replicas: int
    min_replicas: int
    idle_ttl: _duration_pb2.Duration
    scaledown_ttl: _duration_pb2.Duration
    def __init__(self, replicas: _Optional[int] = ..., min_replicas: _Optional[int] = ..., idle_ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., scaledown_ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class Environment(_message.Message):
    __slots__ = ["id", "spec", "scaling"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    SCALING_FIELD_NUMBER: _ClassVar[int]
    id: EnvironmentId
    spec: EnvironmentSpec
    scaling: EnvironmentScaling
    def __init__(self, id: _Optional[_Union[EnvironmentId, _Mapping]] = ..., spec: _Optional[_Union[EnvironmentSpec, _Mapping]] = ..., scaling: _Optional[_Union[EnvironmentScaling, _Mapping]] = ...) -> None: ...

class EnvironmentSummary(_message.Message):
    __slots__ = ["id", "state", "image", "cpu", "memory", "gpu", "gpu_device", "parallelism", "scaling"]
    ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    CPU_FIELD_NUMBER: _ClassVar[int]
    MEMORY_FIELD_NUMBER: _ClassVar[int]
    GPU_FIELD_NUMBER: _ClassVar[int]
    GPU_DEVICE_FIELD_NUMBER: _ClassVar[int]
    PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    SCALING_FIELD_NUMBER: _ClassVar[int]
    id: EnvironmentId
    state: EnvironmentState
    image: str
    cpu: str
    memory: str
    gpu: int
    gpu_device: str
    parallelism: int
    scaling: EnvironmentScaling
    def __init__(self, id: _Optional[_Union[EnvironmentId, _Mapping]] = ..., state: _Optional[_Union[EnvironmentState, str]] = ..., image: _Optional[str] = ..., cpu: _Optional[str] = ..., memory: _Optional[str] = ..., gpu: _Optional[int] = ..., gpu_device: _Optional[str] = ..., parallelism: _Optional[int] = ..., scaling: _Optional[_Union[EnvironmentScaling, _Mapping]] = ...) -> None: ...

class EnvironmentClusterSummary(_message.Message):
    __slots__ = ["cluster", "state", "failure_message", "worker_count", "action_count", "last_updated_at"]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    FAILURE_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    WORKER_COUNT_FIELD_NUMBER: _ClassVar[int]
    ACTION_COUNT_FIELD_NUMBER: _ClassVar[int]
    LAST_UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    cluster: str
    state: EnvironmentClusterState
    failure_message: str
    worker_count: int
    action_count: int
    last_updated_at: _timestamp_pb2.Timestamp
    def __init__(self, cluster: _Optional[str] = ..., state: _Optional[_Union[EnvironmentClusterState, str]] = ..., failure_message: _Optional[str] = ..., worker_count: _Optional[int] = ..., action_count: _Optional[int] = ..., last_updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class EnvironmentListItem(_message.Message):
    __slots__ = ["name", "versions"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSIONS_FIELD_NUMBER: _ClassVar[int]
    name: str
    versions: _containers.RepeatedCompositeFieldContainer[EnvironmentVersionItem]
    def __init__(self, name: _Optional[str] = ..., versions: _Optional[_Iterable[_Union[EnvironmentVersionItem, _Mapping]]] = ...) -> None: ...

class EnvironmentVersionItem(_message.Message):
    __slots__ = ["summary", "cluster_summaries"]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_SUMMARIES_FIELD_NUMBER: _ClassVar[int]
    summary: EnvironmentSummary
    cluster_summaries: _containers.RepeatedCompositeFieldContainer[EnvironmentClusterSummary]
    def __init__(self, summary: _Optional[_Union[EnvironmentSummary, _Mapping]] = ..., cluster_summaries: _Optional[_Iterable[_Union[EnvironmentClusterSummary, _Mapping]]] = ...) -> None: ...

class ListEnvironmentsRequest(_message.Message):
    __slots__ = ["project_id", "request"]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    project_id: _identifier_pb2.ProjectIdentifier
    request: _list_pb2.ListRequest
    def __init__(self, project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ...) -> None: ...

class ListEnvironmentsResponse(_message.Message):
    __slots__ = ["environments", "token"]
    ENVIRONMENTS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    environments: _containers.RepeatedCompositeFieldContainer[EnvironmentListItem]
    token: str
    def __init__(self, environments: _Optional[_Iterable[_Union[EnvironmentListItem, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class GetEnvironmentRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: EnvironmentId
    def __init__(self, id: _Optional[_Union[EnvironmentId, _Mapping]] = ...) -> None: ...

class GetEnvironmentResponse(_message.Message):
    __slots__ = ["environment", "cluster_statuses", "state"]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_STATUSES_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    environment: Environment
    cluster_statuses: _containers.RepeatedCompositeFieldContainer[EnvironmentClusterStatus]
    state: EnvironmentState
    def __init__(self, environment: _Optional[_Union[Environment, _Mapping]] = ..., cluster_statuses: _Optional[_Iterable[_Union[EnvironmentClusterStatus, _Mapping]]] = ..., state: _Optional[_Union[EnvironmentState, str]] = ...) -> None: ...

class EnvironmentClusterStatus(_message.Message):
    __slots__ = ["id", "snapshot", "last_updated_at"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SNAPSHOT_FIELD_NUMBER: _ClassVar[int]
    LAST_UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: EnvironmentClusterId
    snapshot: EnvironmentClusterSnapshot
    last_updated_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[_Union[EnvironmentClusterId, _Mapping]] = ..., snapshot: _Optional[_Union[EnvironmentClusterSnapshot, _Mapping]] = ..., last_updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class EnvironmentClusterSnapshot(_message.Message):
    __slots__ = ["state", "failure_message", "kubernetes_namespace", "workers", "worker_count", "workers_truncated", "actions", "action_count", "actions_truncated", "recent_worker_errors"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    FAILURE_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    WORKER_COUNT_FIELD_NUMBER: _ClassVar[int]
    WORKERS_TRUNCATED_FIELD_NUMBER: _ClassVar[int]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    ACTION_COUNT_FIELD_NUMBER: _ClassVar[int]
    ACTIONS_TRUNCATED_FIELD_NUMBER: _ClassVar[int]
    RECENT_WORKER_ERRORS_FIELD_NUMBER: _ClassVar[int]
    state: EnvironmentClusterState
    failure_message: str
    kubernetes_namespace: str
    workers: _containers.RepeatedCompositeFieldContainer[Worker]
    worker_count: int
    workers_truncated: bool
    actions: _containers.RepeatedCompositeFieldContainer[EnvironmentAction]
    action_count: int
    actions_truncated: bool
    recent_worker_errors: _containers.RepeatedCompositeFieldContainer[WorkerError]
    def __init__(self, state: _Optional[_Union[EnvironmentClusterState, str]] = ..., failure_message: _Optional[str] = ..., kubernetes_namespace: _Optional[str] = ..., workers: _Optional[_Iterable[_Union[Worker, _Mapping]]] = ..., worker_count: _Optional[int] = ..., workers_truncated: bool = ..., actions: _Optional[_Iterable[_Union[EnvironmentAction, _Mapping]]] = ..., action_count: _Optional[int] = ..., actions_truncated: bool = ..., recent_worker_errors: _Optional[_Iterable[_Union[WorkerError, _Mapping]]] = ...) -> None: ...

class Worker(_message.Message):
    __slots__ = ["id", "state", "state_detail", "error", "created_at", "last_heartbeat_at", "active_tasks", "task_capacity"]
    ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    STATE_DETAIL_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_HEARTBEAT_AT_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_TASKS_FIELD_NUMBER: _ClassVar[int]
    TASK_CAPACITY_FIELD_NUMBER: _ClassVar[int]
    id: str
    state: WorkerState
    state_detail: str
    error: str
    created_at: _timestamp_pb2.Timestamp
    last_heartbeat_at: _timestamp_pb2.Timestamp
    active_tasks: int
    task_capacity: int
    def __init__(self, id: _Optional[str] = ..., state: _Optional[_Union[WorkerState, str]] = ..., state_detail: _Optional[str] = ..., error: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., last_heartbeat_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., active_tasks: _Optional[int] = ..., task_capacity: _Optional[int] = ...) -> None: ...

class EnvironmentAction(_message.Message):
    __slots__ = ["action_id", "attempt", "assignment_state", "worker_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    ASSIGNMENT_STATE_FIELD_NUMBER: _ClassVar[int]
    WORKER_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    assignment_state: EnvironmentActionAssignmentState
    worker_id: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., assignment_state: _Optional[_Union[EnvironmentActionAssignmentState, str]] = ..., worker_id: _Optional[str] = ...) -> None: ...

class WorkerError(_message.Message):
    __slots__ = ["worker_id", "error", "observed_at", "exited_at"]
    WORKER_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OBSERVED_AT_FIELD_NUMBER: _ClassVar[int]
    EXITED_AT_FIELD_NUMBER: _ClassVar[int]
    worker_id: str
    error: str
    observed_at: _timestamp_pb2.Timestamp
    exited_at: _timestamp_pb2.Timestamp
    def __init__(self, worker_id: _Optional[str] = ..., error: _Optional[str] = ..., observed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., exited_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
