from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.admin import cluster_assignment_pb2 as _cluster_assignment_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AttributesSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SOURCE_UNSPECIFIED: _ClassVar[AttributesSource]
    GLOBAL: _ClassVar[AttributesSource]
    DOMAIN: _ClassVar[AttributesSource]
    PROJECT: _ClassVar[AttributesSource]
    PROJECT_DOMAIN: _ClassVar[AttributesSource]
    ORG: _ClassVar[AttributesSource]
SOURCE_UNSPECIFIED: AttributesSource
GLOBAL: AttributesSource
DOMAIN: AttributesSource
PROJECT: AttributesSource
PROJECT_DOMAIN: AttributesSource
ORG: AttributesSource

class ConfigurationID(_message.Message):
    __slots__ = ["org", "domain", "project", "workflow"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    GLOBAL_FIELD_NUMBER: _ClassVar[int]
    org: str
    domain: str
    project: str
    workflow: str
    def __init__(self, org: _Optional[str] = ..., domain: _Optional[str] = ..., project: _Optional[str] = ..., workflow: _Optional[str] = ..., **kwargs) -> None: ...

class AttributeMetadata(_message.Message):
    __slots__ = ["is_mutable"]
    IS_MUTABLE_FIELD_NUMBER: _ClassVar[int]
    is_mutable: AttributeIsMutable
    def __init__(self, is_mutable: _Optional[_Union[AttributeIsMutable, _Mapping]] = ...) -> None: ...

class AttributeIsMutable(_message.Message):
    __slots__ = ["value", "reason"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    value: bool
    reason: str
    def __init__(self, value: bool = ..., reason: _Optional[str] = ...) -> None: ...

class TaskResourceAttributesWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.TaskResourceAttributes
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.TaskResourceAttributes, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ClusterResourceAttributesWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.ClusterResourceAttributes
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.ClusterResourceAttributes, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ExecutionQueueAttributesWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.ExecutionQueueAttributes
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.ExecutionQueueAttributes, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ExecutionClusterLabelWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.ExecutionClusterLabel
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.ExecutionClusterLabel, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class QualityOfServiceWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _execution_pb2.QualityOfService
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class PluginOverridesWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.PluginOverrides
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.PluginOverrides, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class WorkflowExecutionConfigWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.WorkflowExecutionConfig
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.WorkflowExecutionConfig, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ClusterAssignmentWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _cluster_assignment_pb2.ClusterAssignment
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_cluster_assignment_pb2.ClusterAssignment, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ExternalResourceAttributesWithSource(_message.Message):
    __slots__ = ["source", "value", "metadata"]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    source: AttributesSource
    value: _matchable_resource_pb2.ExternalResourceAttributes
    metadata: AttributeMetadata
    def __init__(self, source: _Optional[_Union[AttributesSource, str]] = ..., value: _Optional[_Union[_matchable_resource_pb2.ExternalResourceAttributes, _Mapping]] = ..., metadata: _Optional[_Union[AttributeMetadata, _Mapping]] = ...) -> None: ...

class ConfigurationWithSource(_message.Message):
    __slots__ = ["task_resource_attributes", "cluster_resource_attributes", "execution_queue_attributes", "execution_cluster_label", "quality_of_service", "plugin_overrides", "workflow_execution_config", "cluster_assignment", "external_resource_attributes"]
    TASK_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_QUEUE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_CLUSTER_LABEL_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ASSIGNMENT_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    task_resource_attributes: TaskResourceAttributesWithSource
    cluster_resource_attributes: ClusterResourceAttributesWithSource
    execution_queue_attributes: ExecutionQueueAttributesWithSource
    execution_cluster_label: ExecutionClusterLabelWithSource
    quality_of_service: QualityOfServiceWithSource
    plugin_overrides: PluginOverridesWithSource
    workflow_execution_config: WorkflowExecutionConfigWithSource
    cluster_assignment: ClusterAssignmentWithSource
    external_resource_attributes: ExternalResourceAttributesWithSource
    def __init__(self, task_resource_attributes: _Optional[_Union[TaskResourceAttributesWithSource, _Mapping]] = ..., cluster_resource_attributes: _Optional[_Union[ClusterResourceAttributesWithSource, _Mapping]] = ..., execution_queue_attributes: _Optional[_Union[ExecutionQueueAttributesWithSource, _Mapping]] = ..., execution_cluster_label: _Optional[_Union[ExecutionClusterLabelWithSource, _Mapping]] = ..., quality_of_service: _Optional[_Union[QualityOfServiceWithSource, _Mapping]] = ..., plugin_overrides: _Optional[_Union[PluginOverridesWithSource, _Mapping]] = ..., workflow_execution_config: _Optional[_Union[WorkflowExecutionConfigWithSource, _Mapping]] = ..., cluster_assignment: _Optional[_Union[ClusterAssignmentWithSource, _Mapping]] = ..., external_resource_attributes: _Optional[_Union[ExternalResourceAttributesWithSource, _Mapping]] = ...) -> None: ...

class Configuration(_message.Message):
    __slots__ = ["task_resource_attributes", "cluster_resource_attributes", "execution_queue_attributes", "execution_cluster_label", "quality_of_service", "plugin_overrides", "workflow_execution_config", "cluster_assignment", "external_resource_attributes"]
    TASK_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_QUEUE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_CLUSTER_LABEL_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ASSIGNMENT_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    task_resource_attributes: _matchable_resource_pb2.TaskResourceAttributes
    cluster_resource_attributes: _matchable_resource_pb2.ClusterResourceAttributes
    execution_queue_attributes: _matchable_resource_pb2.ExecutionQueueAttributes
    execution_cluster_label: _matchable_resource_pb2.ExecutionClusterLabel
    quality_of_service: _execution_pb2.QualityOfService
    plugin_overrides: _matchable_resource_pb2.PluginOverrides
    workflow_execution_config: _matchable_resource_pb2.WorkflowExecutionConfig
    cluster_assignment: _cluster_assignment_pb2.ClusterAssignment
    external_resource_attributes: _matchable_resource_pb2.ExternalResourceAttributes
    def __init__(self, task_resource_attributes: _Optional[_Union[_matchable_resource_pb2.TaskResourceAttributes, _Mapping]] = ..., cluster_resource_attributes: _Optional[_Union[_matchable_resource_pb2.ClusterResourceAttributes, _Mapping]] = ..., execution_queue_attributes: _Optional[_Union[_matchable_resource_pb2.ExecutionQueueAttributes, _Mapping]] = ..., execution_cluster_label: _Optional[_Union[_matchable_resource_pb2.ExecutionClusterLabel, _Mapping]] = ..., quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., plugin_overrides: _Optional[_Union[_matchable_resource_pb2.PluginOverrides, _Mapping]] = ..., workflow_execution_config: _Optional[_Union[_matchable_resource_pb2.WorkflowExecutionConfig, _Mapping]] = ..., cluster_assignment: _Optional[_Union[_cluster_assignment_pb2.ClusterAssignment, _Mapping]] = ..., external_resource_attributes: _Optional[_Union[_matchable_resource_pb2.ExternalResourceAttributes, _Mapping]] = ...) -> None: ...

class ConfigurationGetRequest(_message.Message):
    __slots__ = ["id", "only_get_lower_level_configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ONLY_GET_LOWER_LEVEL_CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: ConfigurationID
    only_get_lower_level_configuration: bool
    def __init__(self, id: _Optional[_Union[ConfigurationID, _Mapping]] = ..., only_get_lower_level_configuration: bool = ...) -> None: ...

class ConfigurationGetResponse(_message.Message):
    __slots__ = ["id", "version", "configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: ConfigurationID
    version: str
    configuration: ConfigurationWithSource
    def __init__(self, id: _Optional[_Union[ConfigurationID, _Mapping]] = ..., version: _Optional[str] = ..., configuration: _Optional[_Union[ConfigurationWithSource, _Mapping]] = ...) -> None: ...

class ConfigurationUpdateRequest(_message.Message):
    __slots__ = ["id", "version_to_update", "configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_TO_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: ConfigurationID
    version_to_update: str
    configuration: Configuration
    def __init__(self, id: _Optional[_Union[ConfigurationID, _Mapping]] = ..., version_to_update: _Optional[str] = ..., configuration: _Optional[_Union[Configuration, _Mapping]] = ...) -> None: ...

class ConfigurationUpdateResponse(_message.Message):
    __slots__ = ["id", "version", "configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: ConfigurationID
    version: str
    configuration: ConfigurationWithSource
    def __init__(self, id: _Optional[_Union[ConfigurationID, _Mapping]] = ..., version: _Optional[str] = ..., configuration: _Optional[_Union[ConfigurationWithSource, _Mapping]] = ...) -> None: ...

class ConfigurationDocument(_message.Message):
    __slots__ = ["version", "configurations"]
    class ConfigurationsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Configuration
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Configuration, _Mapping]] = ...) -> None: ...
    VERSION_FIELD_NUMBER: _ClassVar[int]
    CONFIGURATIONS_FIELD_NUMBER: _ClassVar[int]
    version: str
    configurations: _containers.MessageMap[str, Configuration]
    def __init__(self, version: _Optional[str] = ..., configurations: _Optional[_Mapping[str, Configuration]] = ...) -> None: ...
