from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.core import security_pb2 as _security_pb2
from flyteidl.admin import schedule_pb2 as _schedule_pb2
from flyteidl.admin import common_pb2 as _common_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LaunchPlanState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    INACTIVE: _ClassVar[LaunchPlanState]
    ACTIVE: _ClassVar[LaunchPlanState]
INACTIVE: LaunchPlanState
ACTIVE: LaunchPlanState

class LaunchPlanCreateRequest(_message.Message):
    __slots__ = ["id", "spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    spec: LaunchPlanSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., spec: _Optional[_Union[LaunchPlanSpec, _Mapping]] = ...) -> None: ...

class LaunchPlanCreateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class LaunchPlan(_message.Message):
    __slots__ = ["id", "spec", "closure"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    spec: LaunchPlanSpec
    closure: LaunchPlanClosure
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., spec: _Optional[_Union[LaunchPlanSpec, _Mapping]] = ..., closure: _Optional[_Union[LaunchPlanClosure, _Mapping]] = ...) -> None: ...

class LaunchPlanList(_message.Message):
    __slots__ = ["launch_plans", "token"]
    LAUNCH_PLANS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    launch_plans: _containers.RepeatedCompositeFieldContainer[LaunchPlan]
    token: str
    def __init__(self, launch_plans: _Optional[_Iterable[_Union[LaunchPlan, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class Auth(_message.Message):
    __slots__ = ["assumable_iam_role", "kubernetes_service_account"]
    ASSUMABLE_IAM_ROLE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    assumable_iam_role: str
    kubernetes_service_account: str
    def __init__(self, assumable_iam_role: _Optional[str] = ..., kubernetes_service_account: _Optional[str] = ...) -> None: ...

class LaunchPlanSpec(_message.Message):
    __slots__ = ["workflow_id", "entity_metadata", "default_inputs", "fixed_inputs", "role", "labels", "annotations", "auth", "auth_role", "security_context", "quality_of_service", "raw_output_data_config", "max_parallelism", "interruptible", "overwrite_cache", "envs"]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    ENTITY_METADATA_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_INPUTS_FIELD_NUMBER: _ClassVar[int]
    FIXED_INPUTS_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    AUTH_FIELD_NUMBER: _ClassVar[int]
    AUTH_ROLE_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    RAW_OUTPUT_DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    MAX_PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    workflow_id: _identifier_pb2.Identifier
    entity_metadata: LaunchPlanMetadata
    default_inputs: _interface_pb2.ParameterMap
    fixed_inputs: _literals_pb2.LiteralMap
    role: str
    labels: _common_pb2.Labels
    annotations: _common_pb2.Annotations
    auth: Auth
    auth_role: _common_pb2.AuthRole
    security_context: _security_pb2.SecurityContext
    quality_of_service: _execution_pb2.QualityOfService
    raw_output_data_config: _common_pb2.RawOutputDataConfig
    max_parallelism: int
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    envs: _common_pb2.Envs
    def __init__(self, workflow_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., entity_metadata: _Optional[_Union[LaunchPlanMetadata, _Mapping]] = ..., default_inputs: _Optional[_Union[_interface_pb2.ParameterMap, _Mapping]] = ..., fixed_inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., role: _Optional[str] = ..., labels: _Optional[_Union[_common_pb2.Labels, _Mapping]] = ..., annotations: _Optional[_Union[_common_pb2.Annotations, _Mapping]] = ..., auth: _Optional[_Union[Auth, _Mapping]] = ..., auth_role: _Optional[_Union[_common_pb2.AuthRole, _Mapping]] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., raw_output_data_config: _Optional[_Union[_common_pb2.RawOutputDataConfig, _Mapping]] = ..., max_parallelism: _Optional[int] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., envs: _Optional[_Union[_common_pb2.Envs, _Mapping]] = ...) -> None: ...

class LaunchPlanClosure(_message.Message):
    __slots__ = ["state", "expected_inputs", "expected_outputs", "created_at", "updated_at"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_INPUTS_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    state: LaunchPlanState
    expected_inputs: _interface_pb2.ParameterMap
    expected_outputs: _interface_pb2.VariableMap
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, state: _Optional[_Union[LaunchPlanState, str]] = ..., expected_inputs: _Optional[_Union[_interface_pb2.ParameterMap, _Mapping]] = ..., expected_outputs: _Optional[_Union[_interface_pb2.VariableMap, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class LaunchPlanMetadata(_message.Message):
    __slots__ = ["schedule", "notifications"]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATIONS_FIELD_NUMBER: _ClassVar[int]
    schedule: _schedule_pb2.Schedule
    notifications: _containers.RepeatedCompositeFieldContainer[_common_pb2.Notification]
    def __init__(self, schedule: _Optional[_Union[_schedule_pb2.Schedule, _Mapping]] = ..., notifications: _Optional[_Iterable[_Union[_common_pb2.Notification, _Mapping]]] = ...) -> None: ...

class LaunchPlanUpdateRequest(_message.Message):
    __slots__ = ["id", "state"]
    ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    state: LaunchPlanState
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., state: _Optional[_Union[LaunchPlanState, str]] = ...) -> None: ...

class LaunchPlanUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ActiveLaunchPlanRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _common_pb2.NamedEntityIdentifier
    def __init__(self, id: _Optional[_Union[_common_pb2.NamedEntityIdentifier, _Mapping]] = ...) -> None: ...

class ActiveLaunchPlanListRequest(_message.Message):
    __slots__ = ["project", "domain", "limit", "token", "sort_by"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    limit: int
    token: str
    sort_by: _common_pb2.Sort
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ...) -> None: ...
