from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.common import runtime_version_pb2 as _runtime_version_pb2
from flyteidl2.core import artifact_id_pb2 as _artifact_id_pb2
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

class Identifier(_message.Message):
    __slots__ = ["org", "project", "domain", "name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class Meta(_message.Message):
    __slots__ = ["id", "revision", "labels"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    id: Identifier
    revision: int
    labels: _containers.ScalarMap[str, str]
    def __init__(self, id: _Optional[_Union[Identifier, _Mapping]] = ..., revision: _Optional[int] = ..., labels: _Optional[_Mapping[str, str]] = ...) -> None: ...

class AppWrapper(_message.Message):
    __slots__ = ["host", "app", "app_id"]
    HOST_FIELD_NUMBER: _ClassVar[int]
    APP_FIELD_NUMBER: _ClassVar[int]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    host: str
    app: App
    app_id: Identifier
    def __init__(self, host: _Optional[str] = ..., app: _Optional[_Union[App, _Mapping]] = ..., app_id: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class App(_message.Message):
    __slots__ = ["metadata", "spec", "status"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    metadata: Meta
    spec: Spec
    status: Status
    def __init__(self, metadata: _Optional[_Union[Meta, _Mapping]] = ..., spec: _Optional[_Union[Spec, _Mapping]] = ..., status: _Optional[_Union[Status, _Mapping]] = ...) -> None: ...

class Condition(_message.Message):
    __slots__ = ["last_transition_time", "deployment_status", "message", "revision", "actor"]
    LAST_TRANSITION_TIME_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_STATUS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    ACTOR_FIELD_NUMBER: _ClassVar[int]
    last_transition_time: _timestamp_pb2.Timestamp
    deployment_status: Status.DeploymentStatus
    message: str
    revision: int
    actor: _identity_pb2.EnrichedIdentity
    def __init__(self, last_transition_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., deployment_status: _Optional[_Union[Status.DeploymentStatus, str]] = ..., message: _Optional[str] = ..., revision: _Optional[int] = ..., actor: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...

class Status(_message.Message):
    __slots__ = ["assigned_cluster", "current_replicas", "ingress", "created_at", "last_updated_at", "conditions", "lease_expiration", "k8s_metadata", "materialized_inputs"]
    class DeploymentStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        DEPLOYMENT_STATUS_UNSPECIFIED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_UNASSIGNED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_ASSIGNED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_PENDING: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_STOPPED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_STARTED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_FAILED: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_ACTIVE: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_SCALING_UP: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_SCALING_DOWN: _ClassVar[Status.DeploymentStatus]
        DEPLOYMENT_STATUS_DEPLOYING: _ClassVar[Status.DeploymentStatus]
    DEPLOYMENT_STATUS_UNSPECIFIED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_UNASSIGNED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_ASSIGNED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_PENDING: Status.DeploymentStatus
    DEPLOYMENT_STATUS_STOPPED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_STARTED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_FAILED: Status.DeploymentStatus
    DEPLOYMENT_STATUS_ACTIVE: Status.DeploymentStatus
    DEPLOYMENT_STATUS_SCALING_UP: Status.DeploymentStatus
    DEPLOYMENT_STATUS_SCALING_DOWN: Status.DeploymentStatus
    DEPLOYMENT_STATUS_DEPLOYING: Status.DeploymentStatus
    ASSIGNED_CLUSTER_FIELD_NUMBER: _ClassVar[int]
    CURRENT_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    CONDITIONS_FIELD_NUMBER: _ClassVar[int]
    LEASE_EXPIRATION_FIELD_NUMBER: _ClassVar[int]
    K8S_METADATA_FIELD_NUMBER: _ClassVar[int]
    MATERIALIZED_INPUTS_FIELD_NUMBER: _ClassVar[int]
    assigned_cluster: str
    current_replicas: int
    ingress: Ingress
    created_at: _timestamp_pb2.Timestamp
    last_updated_at: _timestamp_pb2.Timestamp
    conditions: _containers.RepeatedCompositeFieldContainer[Condition]
    lease_expiration: _timestamp_pb2.Timestamp
    k8s_metadata: K8sMetadata
    materialized_inputs: MaterializedInputs
    def __init__(self, assigned_cluster: _Optional[str] = ..., current_replicas: _Optional[int] = ..., ingress: _Optional[_Union[Ingress, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., last_updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., conditions: _Optional[_Iterable[_Union[Condition, _Mapping]]] = ..., lease_expiration: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., k8s_metadata: _Optional[_Union[K8sMetadata, _Mapping]] = ..., materialized_inputs: _Optional[_Union[MaterializedInputs, _Mapping]] = ...) -> None: ...

class K8sMetadata(_message.Message):
    __slots__ = ["namespace"]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    def __init__(self, namespace: _Optional[str] = ...) -> None: ...

class Ingress(_message.Message):
    __slots__ = ["public_url", "cname_url", "vpc_url"]
    PUBLIC_URL_FIELD_NUMBER: _ClassVar[int]
    CNAME_URL_FIELD_NUMBER: _ClassVar[int]
    VPC_URL_FIELD_NUMBER: _ClassVar[int]
    public_url: str
    cname_url: str
    vpc_url: str
    def __init__(self, public_url: _Optional[str] = ..., cname_url: _Optional[str] = ..., vpc_url: _Optional[str] = ...) -> None: ...

class Spec(_message.Message):
    __slots__ = ["container", "pod", "autoscaling", "ingress", "desired_state", "cluster_pool", "images", "security_context", "extended_resources", "runtime_metadata", "profile", "creator", "inputs", "links", "timeouts"]
    class DesiredState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        DESIRED_STATE_UNSPECIFIED: _ClassVar[Spec.DesiredState]
        DESIRED_STATE_STOPPED: _ClassVar[Spec.DesiredState]
        DESIRED_STATE_STARTED: _ClassVar[Spec.DesiredState]
        DESIRED_STATE_ACTIVE: _ClassVar[Spec.DesiredState]
    DESIRED_STATE_UNSPECIFIED: Spec.DesiredState
    DESIRED_STATE_STOPPED: Spec.DesiredState
    DESIRED_STATE_STARTED: Spec.DesiredState
    DESIRED_STATE_ACTIVE: Spec.DesiredState
    CONTAINER_FIELD_NUMBER: _ClassVar[int]
    POD_FIELD_NUMBER: _ClassVar[int]
    AUTOSCALING_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    DESIRED_STATE_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_POOL_FIELD_NUMBER: _ClassVar[int]
    IMAGES_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    EXTENDED_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_METADATA_FIELD_NUMBER: _ClassVar[int]
    PROFILE_FIELD_NUMBER: _ClassVar[int]
    CREATOR_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    LINKS_FIELD_NUMBER: _ClassVar[int]
    TIMEOUTS_FIELD_NUMBER: _ClassVar[int]
    container: _tasks_pb2.Container
    pod: _tasks_pb2.K8sPod
    autoscaling: AutoscalingConfig
    ingress: IngressConfig
    desired_state: Spec.DesiredState
    cluster_pool: str
    images: ImageSpecSet
    security_context: SecurityContext
    extended_resources: _tasks_pb2.ExtendedResources
    runtime_metadata: _runtime_version_pb2.RuntimeMetadata
    profile: Profile
    creator: _identity_pb2.EnrichedIdentity
    inputs: InputList
    links: _containers.RepeatedCompositeFieldContainer[Link]
    timeouts: TimeoutConfig
    def __init__(self, container: _Optional[_Union[_tasks_pb2.Container, _Mapping]] = ..., pod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ..., autoscaling: _Optional[_Union[AutoscalingConfig, _Mapping]] = ..., ingress: _Optional[_Union[IngressConfig, _Mapping]] = ..., desired_state: _Optional[_Union[Spec.DesiredState, str]] = ..., cluster_pool: _Optional[str] = ..., images: _Optional[_Union[ImageSpecSet, _Mapping]] = ..., security_context: _Optional[_Union[SecurityContext, _Mapping]] = ..., extended_resources: _Optional[_Union[_tasks_pb2.ExtendedResources, _Mapping]] = ..., runtime_metadata: _Optional[_Union[_runtime_version_pb2.RuntimeMetadata, _Mapping]] = ..., profile: _Optional[_Union[Profile, _Mapping]] = ..., creator: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ..., inputs: _Optional[_Union[InputList, _Mapping]] = ..., links: _Optional[_Iterable[_Union[Link, _Mapping]]] = ..., timeouts: _Optional[_Union[TimeoutConfig, _Mapping]] = ...) -> None: ...

class Link(_message.Message):
    __slots__ = ["path", "title", "is_relative"]
    PATH_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    IS_RELATIVE_FIELD_NUMBER: _ClassVar[int]
    path: str
    title: str
    is_relative: bool
    def __init__(self, path: _Optional[str] = ..., title: _Optional[str] = ..., is_relative: bool = ...) -> None: ...

class Input(_message.Message):
    __slots__ = ["name", "string_value", "artifact_query", "artifact_id", "app_id"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_QUERY_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    string_value: str
    artifact_query: _artifact_id_pb2.ArtifactQuery
    artifact_id: _artifact_id_pb2.ArtifactID
    app_id: Identifier
    def __init__(self, name: _Optional[str] = ..., string_value: _Optional[str] = ..., artifact_query: _Optional[_Union[_artifact_id_pb2.ArtifactQuery, _Mapping]] = ..., artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ..., app_id: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class MaterializedInputs(_message.Message):
    __slots__ = ["items", "revision"]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[MaterializedInput]
    revision: int
    def __init__(self, items: _Optional[_Iterable[_Union[MaterializedInput, _Mapping]]] = ..., revision: _Optional[int] = ...) -> None: ...

class MaterializedInput(_message.Message):
    __slots__ = ["name", "artifact_id"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    artifact_id: _artifact_id_pb2.ArtifactID
    def __init__(self, name: _Optional[str] = ..., artifact_id: _Optional[_Union[_artifact_id_pb2.ArtifactID, _Mapping]] = ...) -> None: ...

class InputList(_message.Message):
    __slots__ = ["items"]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[Input]
    def __init__(self, items: _Optional[_Iterable[_Union[Input, _Mapping]]] = ...) -> None: ...

class Profile(_message.Message):
    __slots__ = ["type", "name", "short_description", "icon_url"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SHORT_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    ICON_URL_FIELD_NUMBER: _ClassVar[int]
    type: str
    name: str
    short_description: str
    icon_url: str
    def __init__(self, type: _Optional[str] = ..., name: _Optional[str] = ..., short_description: _Optional[str] = ..., icon_url: _Optional[str] = ...) -> None: ...

class SecurityContext(_message.Message):
    __slots__ = ["run_as", "secrets", "allow_anonymous"]
    RUN_AS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    ALLOW_ANONYMOUS_FIELD_NUMBER: _ClassVar[int]
    run_as: _security_pb2.Identity
    secrets: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    allow_anonymous: bool
    def __init__(self, run_as: _Optional[_Union[_security_pb2.Identity, _Mapping]] = ..., secrets: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ..., allow_anonymous: bool = ...) -> None: ...

class ImageSpec(_message.Message):
    __slots__ = ["tag", "build_job_url"]
    TAG_FIELD_NUMBER: _ClassVar[int]
    BUILD_JOB_URL_FIELD_NUMBER: _ClassVar[int]
    tag: str
    build_job_url: str
    def __init__(self, tag: _Optional[str] = ..., build_job_url: _Optional[str] = ...) -> None: ...

class ImageSpecSet(_message.Message):
    __slots__ = ["images"]
    IMAGES_FIELD_NUMBER: _ClassVar[int]
    images: _containers.RepeatedCompositeFieldContainer[ImageSpec]
    def __init__(self, images: _Optional[_Iterable[_Union[ImageSpec, _Mapping]]] = ...) -> None: ...

class IngressConfig(_message.Message):
    __slots__ = ["private", "subdomain", "cname"]
    PRIVATE_FIELD_NUMBER: _ClassVar[int]
    SUBDOMAIN_FIELD_NUMBER: _ClassVar[int]
    CNAME_FIELD_NUMBER: _ClassVar[int]
    private: bool
    subdomain: str
    cname: str
    def __init__(self, private: bool = ..., subdomain: _Optional[str] = ..., cname: _Optional[str] = ...) -> None: ...

class AutoscalingConfig(_message.Message):
    __slots__ = ["replicas", "scaledown_period", "scaling_metric"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    SCALEDOWN_PERIOD_FIELD_NUMBER: _ClassVar[int]
    SCALING_METRIC_FIELD_NUMBER: _ClassVar[int]
    replicas: Replicas
    scaledown_period: _duration_pb2.Duration
    scaling_metric: ScalingMetric
    def __init__(self, replicas: _Optional[_Union[Replicas, _Mapping]] = ..., scaledown_period: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., scaling_metric: _Optional[_Union[ScalingMetric, _Mapping]] = ...) -> None: ...

class ScalingMetric(_message.Message):
    __slots__ = ["request_rate", "concurrency"]
    REQUEST_RATE_FIELD_NUMBER: _ClassVar[int]
    CONCURRENCY_FIELD_NUMBER: _ClassVar[int]
    request_rate: RequestRate
    concurrency: Concurrency
    def __init__(self, request_rate: _Optional[_Union[RequestRate, _Mapping]] = ..., concurrency: _Optional[_Union[Concurrency, _Mapping]] = ...) -> None: ...

class Concurrency(_message.Message):
    __slots__ = ["target_value"]
    TARGET_VALUE_FIELD_NUMBER: _ClassVar[int]
    target_value: int
    def __init__(self, target_value: _Optional[int] = ...) -> None: ...

class RequestRate(_message.Message):
    __slots__ = ["target_value"]
    TARGET_VALUE_FIELD_NUMBER: _ClassVar[int]
    target_value: int
    def __init__(self, target_value: _Optional[int] = ...) -> None: ...

class Replicas(_message.Message):
    __slots__ = ["min", "max"]
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    min: int
    max: int
    def __init__(self, min: _Optional[int] = ..., max: _Optional[int] = ...) -> None: ...

class TimeoutConfig(_message.Message):
    __slots__ = ["request_timeout"]
    REQUEST_TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    request_timeout: _duration_pb2.Duration
    def __init__(self, request_timeout: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...
