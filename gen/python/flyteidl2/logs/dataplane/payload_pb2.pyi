from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LogLineOriginator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNKNOWN: _ClassVar[LogLineOriginator]
    USER: _ClassVar[LogLineOriginator]
    SYSTEM: _ClassVar[LogLineOriginator]

class LogsSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    LIVE_OR_PERSISTED: _ClassVar[LogsSource]
    LIVE_ONLY: _ClassVar[LogsSource]
    PERSISTED_ONLY: _ClassVar[LogsSource]
UNKNOWN: LogLineOriginator
USER: LogLineOriginator
SYSTEM: LogLineOriginator
LIVE_OR_PERSISTED: LogsSource
LIVE_ONLY: LogsSource
PERSISTED_ONLY: LogsSource

class PodResource(_message.Message):
    __slots__ = ["namespace", "name", "container"]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONTAINER_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    name: str
    container: str
    def __init__(self, namespace: _Optional[str] = ..., name: _Optional[str] = ..., container: _Optional[str] = ...) -> None: ...

class LoggingContext(_message.Message):
    __slots__ = ["cluster_name", "kubernetes_namespace", "kubernetes_pod_name", "kubernetes_container_name", "execution_attempt_start_time", "execution_attempt_end_time", "kubernetes_pod_labels"]
    class KubernetesPodLabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CLUSTER_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_POD_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_CONTAINER_NAME_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ATTEMPT_START_TIME_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ATTEMPT_END_TIME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_POD_LABELS_FIELD_NUMBER: _ClassVar[int]
    cluster_name: str
    kubernetes_namespace: str
    kubernetes_pod_name: str
    kubernetes_container_name: str
    execution_attempt_start_time: _timestamp_pb2.Timestamp
    execution_attempt_end_time: _timestamp_pb2.Timestamp
    kubernetes_pod_labels: _containers.ScalarMap[str, str]
    def __init__(self, cluster_name: _Optional[str] = ..., kubernetes_namespace: _Optional[str] = ..., kubernetes_pod_name: _Optional[str] = ..., kubernetes_container_name: _Optional[str] = ..., execution_attempt_start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., execution_attempt_end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., kubernetes_pod_labels: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ContainerIdentifier(_message.Message):
    __slots__ = ["cluster_name", "kubernetes_namespace", "kubernetes_pod_name", "kubernetes_container_name"]
    CLUSTER_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_POD_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_CONTAINER_NAME_FIELD_NUMBER: _ClassVar[int]
    cluster_name: str
    kubernetes_namespace: str
    kubernetes_pod_name: str
    kubernetes_container_name: str
    def __init__(self, cluster_name: _Optional[str] = ..., kubernetes_namespace: _Optional[str] = ..., kubernetes_pod_name: _Optional[str] = ..., kubernetes_container_name: _Optional[str] = ...) -> None: ...

class ContainerSelector(_message.Message):
    __slots__ = ["cluster_name", "kubernetes_namespace", "kubernetes_pod_name_prefix", "kubernetes_container_name", "kubernetes_pod_label_selector"]
    CLUSTER_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_POD_NAME_PREFIX_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_CONTAINER_NAME_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_POD_LABEL_SELECTOR_FIELD_NUMBER: _ClassVar[int]
    cluster_name: str
    kubernetes_namespace: str
    kubernetes_pod_name_prefix: str
    kubernetes_container_name: str
    kubernetes_pod_label_selector: str
    def __init__(self, cluster_name: _Optional[str] = ..., kubernetes_namespace: _Optional[str] = ..., kubernetes_pod_name_prefix: _Optional[str] = ..., kubernetes_container_name: _Optional[str] = ..., kubernetes_pod_label_selector: _Optional[str] = ...) -> None: ...

class LiveLogsOptions(_message.Message):
    __slots__ = ["log_pod_status", "log_timestamps"]
    LOG_POD_STATUS_FIELD_NUMBER: _ClassVar[int]
    LOG_TIMESTAMPS_FIELD_NUMBER: _ClassVar[int]
    log_pod_status: bool
    log_timestamps: bool
    def __init__(self, log_pod_status: bool = ..., log_timestamps: bool = ...) -> None: ...

class LogLine(_message.Message):
    __slots__ = ["timestamp", "message", "originator"]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ORIGINATOR_FIELD_NUMBER: _ClassVar[int]
    timestamp: _timestamp_pb2.Timestamp
    message: str
    originator: LogLineOriginator
    def __init__(self, timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., message: _Optional[str] = ..., originator: _Optional[_Union[LogLineOriginator, str]] = ...) -> None: ...

class LogLines(_message.Message):
    __slots__ = ["lines", "container_index", "container", "structured_lines"]
    LINES_FIELD_NUMBER: _ClassVar[int]
    CONTAINER_INDEX_FIELD_NUMBER: _ClassVar[int]
    CONTAINER_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_LINES_FIELD_NUMBER: _ClassVar[int]
    lines: _containers.RepeatedScalarFieldContainer[str]
    container_index: int
    container: ContainerIdentifier
    structured_lines: _containers.RepeatedCompositeFieldContainer[LogLine]
    def __init__(self, lines: _Optional[_Iterable[str]] = ..., container_index: _Optional[int] = ..., container: _Optional[_Union[ContainerIdentifier, _Mapping]] = ..., structured_lines: _Optional[_Iterable[_Union[LogLine, _Mapping]]] = ...) -> None: ...

class LogContainersList(_message.Message):
    __slots__ = ["containers"]
    CONTAINERS_FIELD_NUMBER: _ClassVar[int]
    containers: _containers.RepeatedCompositeFieldContainer[ContainerIdentifier]
    def __init__(self, containers: _Optional[_Iterable[_Union[ContainerIdentifier, _Mapping]]] = ...) -> None: ...

class LogLinesBatch(_message.Message):
    __slots__ = ["logs"]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[LogLines]
    def __init__(self, logs: _Optional[_Iterable[_Union[LogLines, _Mapping]]] = ...) -> None: ...
