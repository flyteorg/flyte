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

class Resources(_message.Message):
    __slots__ = ["requests", "limits"]
    class ResourceName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[Resources.ResourceName]
        CPU: _ClassVar[Resources.ResourceName]
        GPU: _ClassVar[Resources.ResourceName]
        MEMORY: _ClassVar[Resources.ResourceName]
        STORAGE: _ClassVar[Resources.ResourceName]
        EPHEMERAL_STORAGE: _ClassVar[Resources.ResourceName]
    UNKNOWN: Resources.ResourceName
    CPU: Resources.ResourceName
    GPU: Resources.ResourceName
    MEMORY: Resources.ResourceName
    STORAGE: Resources.ResourceName
    EPHEMERAL_STORAGE: Resources.ResourceName
    class ResourceEntry(_message.Message):
        __slots__ = ["name", "value"]
        NAME_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        name: Resources.ResourceName
        value: str
        def __init__(self, name: _Optional[_Union[Resources.ResourceName, str]] = ..., value: _Optional[str] = ...) -> None: ...
    REQUESTS_FIELD_NUMBER: _ClassVar[int]
    LIMITS_FIELD_NUMBER: _ClassVar[int]
    requests: _containers.RepeatedCompositeFieldContainer[Resources.ResourceEntry]
    limits: _containers.RepeatedCompositeFieldContainer[Resources.ResourceEntry]
    def __init__(self, requests: _Optional[_Iterable[_Union[Resources.ResourceEntry, _Mapping]]] = ..., limits: _Optional[_Iterable[_Union[Resources.ResourceEntry, _Mapping]]] = ...) -> None: ...

class RuntimeMetadata(_message.Message):
    __slots__ = ["type", "version", "flavor"]
    class RuntimeType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        OTHER: _ClassVar[RuntimeMetadata.RuntimeType]
        FLYTE_SDK: _ClassVar[RuntimeMetadata.RuntimeType]
    OTHER: RuntimeMetadata.RuntimeType
    FLYTE_SDK: RuntimeMetadata.RuntimeType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    FLAVOR_FIELD_NUMBER: _ClassVar[int]
    type: RuntimeMetadata.RuntimeType
    version: str
    flavor: str
    def __init__(self, type: _Optional[_Union[RuntimeMetadata.RuntimeType, str]] = ..., version: _Optional[str] = ..., flavor: _Optional[str] = ...) -> None: ...

class TaskMetadata(_message.Message):
    __slots__ = ["discoverable", "runtime", "timeout", "retries", "discovery_version", "deprecated_error_message", "interruptible", "cache_serializable", "generates_deck", "tags", "pod_template_name"]
    class TagsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    DISCOVERABLE_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    RETRIES_FIELD_NUMBER: _ClassVar[int]
    DISCOVERY_VERSION_FIELD_NUMBER: _ClassVar[int]
    DEPRECATED_ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    CACHE_SERIALIZABLE_FIELD_NUMBER: _ClassVar[int]
    GENERATES_DECK_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    POD_TEMPLATE_NAME_FIELD_NUMBER: _ClassVar[int]
    discoverable: bool
    runtime: RuntimeMetadata
    timeout: _duration_pb2.Duration
    retries: _literals_pb2.RetryStrategy
    discovery_version: str
    deprecated_error_message: str
    interruptible: bool
    cache_serializable: bool
    generates_deck: bool
    tags: _containers.ScalarMap[str, str]
    pod_template_name: str
    def __init__(self, discoverable: bool = ..., runtime: _Optional[_Union[RuntimeMetadata, _Mapping]] = ..., timeout: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., retries: _Optional[_Union[_literals_pb2.RetryStrategy, _Mapping]] = ..., discovery_version: _Optional[str] = ..., deprecated_error_message: _Optional[str] = ..., interruptible: bool = ..., cache_serializable: bool = ..., generates_deck: bool = ..., tags: _Optional[_Mapping[str, str]] = ..., pod_template_name: _Optional[str] = ...) -> None: ...

class TaskTemplate(_message.Message):
    __slots__ = ["id", "type", "metadata", "interface", "custom", "container", "k8s_pod", "sql", "task_type_version", "security_context", "config"]
    class ConfigEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_FIELD_NUMBER: _ClassVar[int]
    CONTAINER_FIELD_NUMBER: _ClassVar[int]
    K8S_POD_FIELD_NUMBER: _ClassVar[int]
    SQL_FIELD_NUMBER: _ClassVar[int]
    TASK_TYPE_VERSION_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    type: str
    metadata: TaskMetadata
    interface: _interface_pb2.TypedInterface
    custom: _struct_pb2.Struct
    container: Container
    k8s_pod: K8sPod
    sql: Sql
    task_type_version: int
    security_context: _security_pb2.SecurityContext
    config: _containers.ScalarMap[str, str]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., type: _Optional[str] = ..., metadata: _Optional[_Union[TaskMetadata, _Mapping]] = ..., interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ..., custom: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., container: _Optional[_Union[Container, _Mapping]] = ..., k8s_pod: _Optional[_Union[K8sPod, _Mapping]] = ..., sql: _Optional[_Union[Sql, _Mapping]] = ..., task_type_version: _Optional[int] = ..., security_context: _Optional[_Union[_security_pb2.SecurityContext, _Mapping]] = ..., config: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ContainerPort(_message.Message):
    __slots__ = ["container_port"]
    CONTAINER_PORT_FIELD_NUMBER: _ClassVar[int]
    container_port: int
    def __init__(self, container_port: _Optional[int] = ...) -> None: ...

class Container(_message.Message):
    __slots__ = ["image", "command", "args", "resources", "env", "config", "ports", "data_config", "architecture"]
    class Architecture(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[Container.Architecture]
        AMD64: _ClassVar[Container.Architecture]
        ARM64: _ClassVar[Container.Architecture]
        ARM_V6: _ClassVar[Container.Architecture]
        ARM_V7: _ClassVar[Container.Architecture]
    UNKNOWN: Container.Architecture
    AMD64: Container.Architecture
    ARM64: Container.Architecture
    ARM_V6: Container.Architecture
    ARM_V7: Container.Architecture
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    ARGS_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    PORTS_FIELD_NUMBER: _ClassVar[int]
    DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    ARCHITECTURE_FIELD_NUMBER: _ClassVar[int]
    image: str
    command: _containers.RepeatedScalarFieldContainer[str]
    args: _containers.RepeatedScalarFieldContainer[str]
    resources: Resources
    env: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    config: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    ports: _containers.RepeatedCompositeFieldContainer[ContainerPort]
    data_config: DataLoadingConfig
    architecture: Container.Architecture
    def __init__(self, image: _Optional[str] = ..., command: _Optional[_Iterable[str]] = ..., args: _Optional[_Iterable[str]] = ..., resources: _Optional[_Union[Resources, _Mapping]] = ..., env: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ..., config: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ..., ports: _Optional[_Iterable[_Union[ContainerPort, _Mapping]]] = ..., data_config: _Optional[_Union[DataLoadingConfig, _Mapping]] = ..., architecture: _Optional[_Union[Container.Architecture, str]] = ...) -> None: ...

class IOStrategy(_message.Message):
    __slots__ = ["download_mode", "upload_mode"]
    class DownloadMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        DOWNLOAD_EAGER: _ClassVar[IOStrategy.DownloadMode]
        DOWNLOAD_STREAM: _ClassVar[IOStrategy.DownloadMode]
        DO_NOT_DOWNLOAD: _ClassVar[IOStrategy.DownloadMode]
    DOWNLOAD_EAGER: IOStrategy.DownloadMode
    DOWNLOAD_STREAM: IOStrategy.DownloadMode
    DO_NOT_DOWNLOAD: IOStrategy.DownloadMode
    class UploadMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UPLOAD_ON_EXIT: _ClassVar[IOStrategy.UploadMode]
        UPLOAD_EAGER: _ClassVar[IOStrategy.UploadMode]
        DO_NOT_UPLOAD: _ClassVar[IOStrategy.UploadMode]
    UPLOAD_ON_EXIT: IOStrategy.UploadMode
    UPLOAD_EAGER: IOStrategy.UploadMode
    DO_NOT_UPLOAD: IOStrategy.UploadMode
    DOWNLOAD_MODE_FIELD_NUMBER: _ClassVar[int]
    UPLOAD_MODE_FIELD_NUMBER: _ClassVar[int]
    download_mode: IOStrategy.DownloadMode
    upload_mode: IOStrategy.UploadMode
    def __init__(self, download_mode: _Optional[_Union[IOStrategy.DownloadMode, str]] = ..., upload_mode: _Optional[_Union[IOStrategy.UploadMode, str]] = ...) -> None: ...

class DataLoadingConfig(_message.Message):
    __slots__ = ["enabled", "input_path", "output_path", "format", "io_strategy"]
    class LiteralMapFormat(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        JSON: _ClassVar[DataLoadingConfig.LiteralMapFormat]
        YAML: _ClassVar[DataLoadingConfig.LiteralMapFormat]
        PROTO: _ClassVar[DataLoadingConfig.LiteralMapFormat]
    JSON: DataLoadingConfig.LiteralMapFormat
    YAML: DataLoadingConfig.LiteralMapFormat
    PROTO: DataLoadingConfig.LiteralMapFormat
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    INPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    IO_STRATEGY_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    input_path: str
    output_path: str
    format: DataLoadingConfig.LiteralMapFormat
    io_strategy: IOStrategy
    def __init__(self, enabled: bool = ..., input_path: _Optional[str] = ..., output_path: _Optional[str] = ..., format: _Optional[_Union[DataLoadingConfig.LiteralMapFormat, str]] = ..., io_strategy: _Optional[_Union[IOStrategy, _Mapping]] = ...) -> None: ...

class K8sPod(_message.Message):
    __slots__ = ["metadata", "pod_spec", "data_config"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    POD_SPEC_FIELD_NUMBER: _ClassVar[int]
    DATA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    metadata: K8sObjectMetadata
    pod_spec: _struct_pb2.Struct
    data_config: DataLoadingConfig
    def __init__(self, metadata: _Optional[_Union[K8sObjectMetadata, _Mapping]] = ..., pod_spec: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., data_config: _Optional[_Union[DataLoadingConfig, _Mapping]] = ...) -> None: ...

class K8sObjectMetadata(_message.Message):
    __slots__ = ["labels", "annotations"]
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
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    labels: _containers.ScalarMap[str, str]
    annotations: _containers.ScalarMap[str, str]
    def __init__(self, labels: _Optional[_Mapping[str, str]] = ..., annotations: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Sql(_message.Message):
    __slots__ = ["statement", "dialect"]
    class Dialect(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNDEFINED: _ClassVar[Sql.Dialect]
        ANSI: _ClassVar[Sql.Dialect]
        HIVE: _ClassVar[Sql.Dialect]
        OTHER: _ClassVar[Sql.Dialect]
    UNDEFINED: Sql.Dialect
    ANSI: Sql.Dialect
    HIVE: Sql.Dialect
    OTHER: Sql.Dialect
    STATEMENT_FIELD_NUMBER: _ClassVar[int]
    DIALECT_FIELD_NUMBER: _ClassVar[int]
    statement: str
    dialect: Sql.Dialect
    def __init__(self, statement: _Optional[str] = ..., dialect: _Optional[_Union[Sql.Dialect, str]] = ...) -> None: ...
