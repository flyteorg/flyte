from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import security_pb2 as _security_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Container(_message.Message):
    __slots__ = ["architecture", "args", "command", "config", "data_config", "env", "image", "ports", "resources"]
    class Architecture(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    AMD64: Container.Architecture
    ARCHITECTURE_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    ARM64: Container.Architecture
    ARM_V6: Container.Architecture
    ARM_V7: Container.Architecture
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    PORTS_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    UNKNOWN: Container.Architecture
    architecture: Container.Architecture
    args: _containers.RepeatedScalarFieldContainer[str]
    command: _containers.RepeatedScalarFieldContainer[str]
    config: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    data_config: DataLoadingConfig
    env: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    image: str
    ports: _containers.RepeatedCompositeFieldContainer[ContainerPort]
    resources: Resources
    def __init__(self, image: _Optional[str] = ..., command: _Optional[_Iterable[str]] = ..., args: _Optional[_Iterable[str]] = ..., resources: _Optional[_Union[Resources, _Mapping]] = ..., env: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ..., config: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ..., ports: _Optional[_Iterable[_Union[ContainerPort, _Mapping]]] = ..., data_config: _Optional[_Union[DataLoadingConfig, _Mapping]] = ..., architecture: _Optional[_Union[Container.Architecture, str]] = ...) -> None: ...

class ContainerPort(_message.Message):
    __slots__ = ["container_port"]
    CONTAINER_PORT_FIELD_NUMBER: _ClassVar[int]
    container_port: int
    def __init__(self, container_port: _Optional[int] = ...) -> None: ...

class DataLoadingConfig(_message.Message):
    __slots__ = ["enabled", "format", "input_path", "io_strategy", "output_path"]
    class LiteralMapFormat(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    INPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    IO_STRATEGY_FIELD_NUMBER: _ClassVar[int]
    JSON: DataLoadingConfig.LiteralMapFormat
    OUTPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    PROTO: DataLoadingConfig.LiteralMapFormat
    YAML: DataLoadingConfig.LiteralMapFormat
    enabled: bool
    format: DataLoadingConfig.LiteralMapFormat
    input_path: str
    io_strategy: IOStrategy
    output_path: str
    def __init__(self, enabled: bool = ..., input_path: _Optional[str] = ..., output_path: _Optional[str] = ..., format: _Optional[_Union[DataLoadingConfig.LiteralMapFormat, str]] = ..., io_strategy: _Optional[_Union[IOStrategy, _Mapping]] = ...) -> None: ...

class IOStrategy(_message.Message):
    __slots__ = ["download_mode", "upload_mode"]
    class DownloadMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    class UploadMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    DOWNLOAD_EAGER: IOStrategy.DownloadMode
    DOWNLOAD_MODE_FIELD_NUMBER: _ClassVar[int]
    DOWNLOAD_STREAM: IOStrategy.DownloadMode
    DO_NOT_DOWNLOAD: IOStrategy.DownloadMode
    DO_NOT_UPLOAD: IOStrategy.UploadMode
    UPLOAD_EAGER: IOStrategy.UploadMode
    UPLOAD_MODE_FIELD_NUMBER: _ClassVar[int]
    UPLOAD_ON_EXIT: IOStrategy.UploadMode
    download_mode: IOStrategy.DownloadMode
    upload_mode: IOStrategy.UploadMode
    def __init__(self, download_mode: _Optional[_Union[IOStrategy.DownloadMode, str]] = ..., upload_mode: _Optional[_Union[IOStrategy.UploadMode, str]] = ...) -> None: ...

class K8sObjectMetadata(_message.Message):
    __slots__ = ["annotations", "labels"]
    class AnnotationsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    annotations: _containers.ScalarMap[str, str]
    labels: _containers.ScalarMap[str, str]
    def __init__(self, labels: _Optional[_Mapping[str, str]] = ..., annotations: _Optional[_Mapping[str, str]] = ...) -> None: ...

class K8sPod(_message.Message):
    __slots__ = ["metadata", "pod_spec"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    POD_SPEC_FIELD_NUMBER: _ClassVar[int]
    metadata: K8sObjectMetadata
    pod_spec: _struct_pb2.Struct
    def __init__(self, metadata: _Optional[_Union[K8sObjectMetadata, _Mapping]] = ..., pod_spec: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class Resources(_message.Message):
    __slots__ = ["limits", "requests"]
    class ResourceName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    class ResourceEntry(_message.Message):
        __slots__ = ["name", "value"]
        NAME_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        name: Resources.ResourceName
        value: str
        def __init__(self, name: _Optional[_Union[Resources.ResourceName, str]] = ..., value: _Optional[str] = ...) -> None: ...
    CPU: Resources.ResourceName
    EPHEMERAL_STORAGE: Resources.ResourceName
    GPU: Resources.ResourceName
    LIMITS_FIELD_NUMBER: _ClassVar[int]
    MEMORY: Resources.ResourceName
    REQUESTS_FIELD_NUMBER: _ClassVar[int]
    STORAGE: Resources.ResourceName
    UNKNOWN: Resources.ResourceName
    limits: _containers.RepeatedCompositeFieldContainer[Resources.ResourceEntry]
    requests: _containers.RepeatedCompositeFieldContainer[Resources.ResourceEntry]
    def __init__(self, requests: _Optional[_Iterable[_Union[Resources.ResourceEntry, _Mapping]]] = ..., limits: _Optional[_Iterable[_Union[Resources.ResourceEntry, _Mapping]]] = ...) -> None: ...

class RuntimeMetadata(_message.Message):
    __slots__ = ["flavor", "type", "version"]
    class RuntimeType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    FLAVOR_FIELD_NUMBER: _ClassVar[int]
    FLYTE_SDK: RuntimeMetadata.RuntimeType
    OTHER: RuntimeMetadata.RuntimeType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    flavor: str
    type: RuntimeMetadata.RuntimeType
    version: str
    def __init__(self, type: _Optional[_Union[RuntimeMetadata.RuntimeType, str]] = ..., version: _Optional[str] = ..., flavor: _Optional[str] = ...) -> None: ...

class Sql(_message.Message):
    __slots__ = ["dialect", "statement"]
    class Dialect(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ANSI: Sql.Dialect
    DIALECT_FIELD_NUMBER: _ClassVar[int]
    HIVE: Sql.Dialect
    OTHER: Sql.Dialect
    STATEMENT_FIELD_NUMBER: _ClassVar[int]
    UNDEFINED: Sql.Dialect
    dialect: Sql.Dialect
    statement: str
    def __init__(self, statement: _Optional[str] = ..., dialect: _Optional[_Union[Sql.Dialect, str]] = ...) -> None: ...

class TaskMetadata(_message.Message):
    __slots__ = ["cache_serializable", "deprecated_error_message", "discoverable", "discovery_version", "generates_deck", "interruptible", "pod_template_name", "retries", "runtime", "tags", "timeout"]
    class TagsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CACHE_SERIALIZABLE_FIELD_NUMBER: _ClassVar[int]
    DEPRECATED_ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    DISCOVERABLE_FIELD_NUMBER: _ClassVar[int]
    DISCOVERY_VERSION_FIELD_NUMBER: _ClassVar[int]
    GENERATES_DECK_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    POD_TEMPLATE_NAME_FIELD_NUMBER: _ClassVar[int]
    RETRIES_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    cache_serializable: bool
    deprecated_error_message: str
    discoverable: bool
    discovery_version: str
    generates_deck: bool
    interruptible: bool
    pod_template_name: str
    retries: _literals_pb2.RetryStrategy
    runtime: RuntimeMetadata
    tags: _containers.ScalarMap[str, str]
    timeout: _duration_pb2.Duration
    def __init__(self, discoverable: bool = ..., runtime: _Optional[_Union[RuntimeMetadata, _Mapping]] = ..., timeout: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., retries: _Optional[_Union[_literals_pb2.RetryStrategy, _Mapping]] = ..., discovery_version: _Optional[str] = ..., deprecated_error_message: _Optional[str] = ..., interruptible: bool = ..., cache_serializable: bool = ..., generates_deck: bool = ..., tags: _Optional[_Mapping[str, str]] = ..., pod_template_name: _Optional[str] = ...) -> None: ...

class TaskTemplate(_message.Message):
    __slots__ = ["config", "container", "custom", "id", "interface", "k8s_pod", "metadata", "security_context", "sql", "task_type_version", "type"]
    class ConfigEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    CONTAINER_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    K8S_POD_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    SQL_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_VERSION_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    config: _containers.ScalarMap[str, str]
    container: Container
    custom: _struct_pb2.Struct
    id: _identifier_pb2.Identifier
    interface: _interface_pb2.TypedInterface
    k8s_pod: K8sPod
    metadata: TaskMetadata
    security_context: _security_pb2.SecurityContext
    sql: Sql
    task_type_version: int
    type: str
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., type: _Optional[str] = ..., metadata: _Optional[_Union[TaskMetadata, _Mapping]] = ..., interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ..., custom: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., container: _Optional[_Union[Container, _Mapping]] = ..., k8s_pod: _Optional[_Union[K8sPod, _Mapping]] = ..., sql: _Optional[_Union[Sql, _Mapping]] = ..., task_type_version: _Optional[int] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., config: _Optional[_Mapping[str, str]] = ...) -> None: ...
