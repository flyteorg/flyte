from flyteidl.admin import matchable_resource_pb2 as _matchable_resource_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.admin import cluster_assignment_pb2 as _cluster_assignment_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class configurationID(_message.Message):
    __slots__ = ["project", "domain", "workflow", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    workflow: str
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., workflow: _Optional[str] = ..., org: _Optional[str] = ...) -> None: ...

class Configuration(_message.Message):
    __slots__ = ["task_resource_attributes", "cluster_resource_attributes", "execution_queue_attributes", "execution_cluster_label", "quality_of_service", "plugin_overrides", "workflow_execution_config", "cluster_assignment"]
    TASK_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_RESOURCE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_QUEUE_ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_CLUSTER_LABEL_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    PLUGIN_OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ASSIGNMENT_FIELD_NUMBER: _ClassVar[int]
    task_resource_attributes: _matchable_resource_pb2.TaskResourceAttributes
    cluster_resource_attributes: _matchable_resource_pb2.ClusterResourceAttributes
    execution_queue_attributes: _matchable_resource_pb2.ExecutionQueueAttributes
    execution_cluster_label: _matchable_resource_pb2.ExecutionClusterLabel
    quality_of_service: _execution_pb2.QualityOfService
    plugin_overrides: _matchable_resource_pb2.PluginOverrides
    workflow_execution_config: _matchable_resource_pb2.WorkflowExecutionConfig
    cluster_assignment: _cluster_assignment_pb2.ClusterAssignment
    def __init__(self, task_resource_attributes: _Optional[_Union[_matchable_resource_pb2.TaskResourceAttributes, _Mapping]] = ..., cluster_resource_attributes: _Optional[_Union[_matchable_resource_pb2.ClusterResourceAttributes, _Mapping]] = ..., execution_queue_attributes: _Optional[_Union[_matchable_resource_pb2.ExecutionQueueAttributes, _Mapping]] = ..., execution_cluster_label: _Optional[_Union[_matchable_resource_pb2.ExecutionClusterLabel, _Mapping]] = ..., quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., plugin_overrides: _Optional[_Union[_matchable_resource_pb2.PluginOverrides, _Mapping]] = ..., workflow_execution_config: _Optional[_Union[_matchable_resource_pb2.WorkflowExecutionConfig, _Mapping]] = ..., cluster_assignment: _Optional[_Union[_cluster_assignment_pb2.ClusterAssignment, _Mapping]] = ...) -> None: ...

class ConfigurationGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: configurationID
    def __init__(self, id: _Optional[_Union[configurationID, _Mapping]] = ...) -> None: ...

class ConfigurationGetResponse(_message.Message):
    __slots__ = ["id", "version", "project_domain_configuration", "project_configuration", "global_configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    PROJECT_DOMAIN_CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    PROJECT_CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    GLOBAL_CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: configurationID
    version: str
    project_domain_configuration: Configuration
    project_configuration: Configuration
    global_configuration: Configuration
    def __init__(self, id: _Optional[_Union[configurationID, _Mapping]] = ..., version: _Optional[str] = ..., project_domain_configuration: _Optional[_Union[Configuration, _Mapping]] = ..., project_configuration: _Optional[_Union[Configuration, _Mapping]] = ..., global_configuration: _Optional[_Union[Configuration, _Mapping]] = ...) -> None: ...

class ConfigurationUpdateRequest(_message.Message):
    __slots__ = ["id", "version_to_update", "configuration"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_TO_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    id: configurationID
    version_to_update: str
    configuration: Configuration
    def __init__(self, id: _Optional[_Union[configurationID, _Mapping]] = ..., version_to_update: _Optional[str] = ..., configuration: _Optional[_Union[Configuration, _Mapping]] = ...) -> None: ...

class ConfigurationUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

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
