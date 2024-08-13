from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.admin import cluster_assignment_pb2 as _cluster_assignment_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import execution_envs_pb2 as _execution_envs_pb2
from flyteidl.core import security_pb2 as _security_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class MatchableResource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    TASK_RESOURCE: _ClassVar[MatchableResource]
    CLUSTER_RESOURCE: _ClassVar[MatchableResource]
    EXECUTION_QUEUE: _ClassVar[MatchableResource]
    EXECUTION_CLUSTER_LABEL: _ClassVar[MatchableResource]
    QUALITY_OF_SERVICE_SPECIFICATION: _ClassVar[MatchableResource]
    PLUGIN_OVERRIDE: _ClassVar[MatchableResource]
    WORKFLOW_EXECUTION_CONFIG: _ClassVar[MatchableResource]
    CLUSTER_ASSIGNMENT: _ClassVar[MatchableResource]
TASK_RESOURCE: MatchableResource
CLUSTER_RESOURCE: MatchableResource
EXECUTION_QUEUE: MatchableResource
EXECUTION_CLUSTER_LABEL: MatchableResource
QUALITY_OF_SERVICE_SPECIFICATION: MatchableResource
PLUGIN_OVERRIDE: MatchableResource
WORKFLOW_EXECUTION_CONFIG: MatchableResource
CLUSTER_ASSIGNMENT: MatchableResource

class TaskResourceSpec(_message.Message):
    __slots__ = ["cpu", "gpu", "memory", "storage", "ephemeral_storage"]
    CPU_FIELD_NUMBER: _ClassVar[int]
    GPU_FIELD_NUMBER: _ClassVar[int]
    MEMORY_FIELD_NUMBER: _ClassVar[int]
    STORAGE_FIELD_NUMBER: _ClassVar[int]
    EPHEMERAL_STORAGE_FIELD_NUMBER: _ClassVar[int]
    cpu: str
    gpu: str
    memory: str
    storage: str
    ephemeral_storage: str
    def __init__(self, cpu: _Optional[str] = ..., gpu: _Optional[str] = ..., memory: _Optional[str] = ..., storage: _Optional[str] = ..., ephemeral_storage: _Optional[str] = ...) -> None: ...

class TaskResourceAttributes(_message.Message):
    __slots__ = ["defaults", "limits"]
    DEFAULTS_FIELD_NUMBER: _ClassVar[int]
    LIMITS_FIELD_NUMBER: _ClassVar[int]
    defaults: TaskResourceSpec
    limits: TaskResourceSpec
    def __init__(self, defaults: _Optional[_Union[TaskResourceSpec, _Mapping]] = ..., limits: _Optional[_Union[TaskResourceSpec, _Mapping]] = ...) -> None: ...

class ClusterResourceAttributes(_message.Message):
    __slots__ = ["attributes"]
    class AttributesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    attributes: _containers.ScalarMap[str, str]
    def __init__(self, attributes: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ExecutionQueueAttributes(_message.Message):
    __slots__ = ["tags"]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    tags: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, tags: _Optional[_Iterable[str]] = ...) -> None: ...

class ExecutionClusterLabel(_message.Message):
    __slots__ = ["value"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: str
    def __init__(self, value: _Optional[str] = ...) -> None: ...

class PluginOverride(_message.Message):
    __slots__ = ["task_type", "plugin_id", "missing_plugin_behavior"]
    class MissingPluginBehavior(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        FAIL: _ClassVar[PluginOverride.MissingPluginBehavior]
        USE_DEFAULT: _ClassVar[PluginOverride.MissingPluginBehavior]
    FAIL: PluginOverride.MissingPluginBehavior
    USE_DEFAULT: PluginOverride.MissingPluginBehavior
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_ID_FIELD_NUMBER: _ClassVar[int]
    MISSING_PLUGIN_BEHAVIOR_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    plugin_id: _containers.RepeatedScalarFieldContainer[str]
    missing_plugin_behavior: PluginOverride.MissingPluginBehavior
    def __init__(self, task_type: _Optional[str] = ..., plugin_id: _Optional[_Iterable[str]] = ..., missing_plugin_behavior: _Optional[_Union[PluginOverride.MissingPluginBehavior, str]] = ...) -> None: ...

class PluginOverrides(_message.Message):
    __slots__ = ["overrides"]
    OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    overrides: _containers.RepeatedCompositeFieldContainer[PluginOverride]
    def __init__(self, overrides: _Optional[_Iterable[_Union[PluginOverride, _Mapping]]] = ...) -> None: ...

class WorkflowExecutionConfig(_message.Message):
    __slots__ = ["max_parallelism", "security_context", "raw_output_data_config", "labels", "annotations", "interruptible", "overwrite_cache", "envs", "execution_env_assignments"]
    MAX_PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    RAW_OUTPUT_DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ENV_ASSIGNMENTS_FIELD_NUMBER: _ClassVar[int]
    max_parallelism: int
    security_context: _security_pb2.SecurityContext
    raw_output_data_config: _common_pb2.RawOutputDataConfig
    labels: _common_pb2.Labels
    annotations: _common_pb2.Annotations
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    envs: _common_pb2.Envs
    execution_env_assignments: _containers.RepeatedCompositeFieldContainer[_execution_envs_pb2.ExecutionEnvAssignment]
    def __init__(self, max_parallelism: _Optional[int] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., raw_output_data_config: _Optional[_Union[_common_pb2.RawOutputDataConfig, _Mapping]] = ..., labels: _Optional[_Union[_common_pb2.Labels, _Mapping]] = ..., annotations: _Optional[_Union[_common_pb2.Annotations, _Mapping]] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., envs: _Optional[_Union[_common_pb2.Envs, _Mapping]] = ..., execution_env_assignments: _Optional[_Iterable[_Union[_execution_envs_pb2.ExecutionEnvAssignment, _Mapping]]] = ...) -> None: ...

class MatchingAttributes(_message.Message):
    __slots__ = ["task_resource_attributes", "cluster_resource_attributes", "execution_queue_attributes", "execution_cluster_label", "quality_of_service", "plugin_overrides", "workflow_execution_config", "cluster_assignment"]
    TASK_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_QUEUE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_CLUSTER_LABEL_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ASSIGNMENT_FIELD_NUMBER: _ClassVar[int]
    task_resource_attributes: TaskResourceAttributes
    cluster_resource_attributes: ClusterResourceAttributes
    execution_queue_attributes: ExecutionQueueAttributes
    execution_cluster_label: ExecutionClusterLabel
    quality_of_service: _execution_pb2.QualityOfService
    plugin_overrides: PluginOverrides
    workflow_execution_config: WorkflowExecutionConfig
    cluster_assignment: _cluster_assignment_pb2.ClusterAssignment
    def __init__(self, task_resource_attributes: _Optional[_Union[TaskResourceAttributes, _Mapping]] = ..., cluster_resource_attributes: _Optional[_Union[ClusterResourceAttributes, _Mapping]] = ..., execution_queue_attributes: _Optional[_Union[ExecutionQueueAttributes, _Mapping]] = ..., execution_cluster_label: _Optional[_Union[ExecutionClusterLabel, _Mapping]] = ..., quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., plugin_overrides: _Optional[_Union[PluginOverrides, _Mapping]] = ..., workflow_execution_config: _Optional[_Union[WorkflowExecutionConfig, _Mapping]] = ..., cluster_assignment: _Optional[_Union[_cluster_assignment_pb2.ClusterAssignment, _Mapping]] = ...) -> None: ...

class MatchableAttributesConfiguration(_message.Message):
    __slots__ = ["attributes", "domain", "project", "workflow", "launch_plan", "org"]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    LAUNCH_PLAN_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    attributes: MatchingAttributes
    domain: str
    project: str
    workflow: str
    launch_plan: str
    org: str
    def __init__(self, attributes: _Optional[_Union[MatchingAttributes, _Mapping]] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ..., workflow: _Optional[str] = ..., launch_plan: _Optional[str] = ..., org: _Optional[str] = ...) -> None: ...

class ListMatchableAttributesRequest(_message.Message):
    __slots__ = ["resource_type", "org"]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    resource_type: MatchableResource
    org: str
    def __init__(self, resource_type: _Optional[_Union[MatchableResource, str]] = ..., org: _Optional[str] = ...) -> None: ...

class ListMatchableAttributesResponse(_message.Message):
    __slots__ = ["configurations"]
    CONFIGURATIONS_FIELD_NUMBER: _ClassVar[int]
    configurations: _containers.RepeatedCompositeFieldContainer[MatchableAttributesConfiguration]
    def __init__(self, configurations: _Optional[_Iterable[_Union[MatchableAttributesConfiguration, _Mapping]]] = ...) -> None: ...
