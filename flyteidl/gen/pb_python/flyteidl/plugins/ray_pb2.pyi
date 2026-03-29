from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RayJob(_message.Message):
    __slots__ = ["ray_cluster", "runtime_env", "shutdown_after_job_finishes", "ttl_seconds_after_finished", "runtime_env_yaml"]
    RAY_CLUSTER_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_ENV_FIELD_NUMBER: _ClassVar[int]
    SHUTDOWN_AFTER_JOB_FINISHES_FIELD_NUMBER: _ClassVar[int]
    TTL_SECONDS_AFTER_FINISHED_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_ENV_YAML_FIELD_NUMBER: _ClassVar[int]
    ray_cluster: RayCluster
    runtime_env: str
    shutdown_after_job_finishes: bool
    ttl_seconds_after_finished: int
    runtime_env_yaml: str
    def __init__(self, ray_cluster: _Optional[_Union[RayCluster, _Mapping]] = ..., runtime_env: _Optional[str] = ..., shutdown_after_job_finishes: bool = ..., ttl_seconds_after_finished: _Optional[int] = ..., runtime_env_yaml: _Optional[str] = ...) -> None: ...

class Resources(_message.Message):
    __slots__ = ["requests", "limits"]
    class ResourceName(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        CPU: _ClassVar[Resources.ResourceName]
        MEMORY: _ClassVar[Resources.ResourceName]
    CPU: Resources.ResourceName
    MEMORY: Resources.ResourceName
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

class EnvValueFrom(_message.Message):
    __slots__ = ["source", "name", "key"]
    class Source(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        CONFIGMAP: _ClassVar[EnvValueFrom.Source]
        SECRET: _ClassVar[EnvValueFrom.Source]
        RESOURCEFIELD: _ClassVar[EnvValueFrom.Source]
        FIELD: _ClassVar[EnvValueFrom.Source]
    CONFIGMAP: EnvValueFrom.Source
    SECRET: EnvValueFrom.Source
    RESOURCEFIELD: EnvValueFrom.Source
    FIELD: EnvValueFrom.Source
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    source: EnvValueFrom.Source
    name: str
    key: str
    def __init__(self, source: _Optional[_Union[EnvValueFrom.Source, str]] = ..., name: _Optional[str] = ..., key: _Optional[str] = ...) -> None: ...

class EnvVar(_message.Message):
    __slots__ = ["name", "value", "valuesFrom"]
    class ValuesFromEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: EnvValueFrom
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[EnvValueFrom, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    VALUESFROM_FIELD_NUMBER: _ClassVar[int]
    name: str
    value: str
    valuesFrom: _containers.MessageMap[str, EnvValueFrom]
    def __init__(self, name: _Optional[str] = ..., value: _Optional[str] = ..., valuesFrom: _Optional[_Mapping[str, EnvValueFrom]] = ...) -> None: ...

class AutoscalerOptions(_message.Message):
    __slots__ = ["upscaling_mode", "idle_timeout_seconds", "env", "image", "resources"]
    UPSCALING_MODE_FIELD_NUMBER: _ClassVar[int]
    IDLE_TIMEOUT_SECONDS_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    upscaling_mode: str
    idle_timeout_seconds: int
    env: _containers.RepeatedCompositeFieldContainer[EnvVar]
    image: str
    resources: Resources
    def __init__(self, upscaling_mode: _Optional[str] = ..., idle_timeout_seconds: _Optional[int] = ..., env: _Optional[_Iterable[_Union[EnvVar, _Mapping]]] = ..., image: _Optional[str] = ..., resources: _Optional[_Union[Resources, _Mapping]] = ...) -> None: ...

class RayCluster(_message.Message):
    __slots__ = ["head_group_spec", "worker_group_spec", "enable_autoscaling", "autoscaler_options"]
    HEAD_GROUP_SPEC_FIELD_NUMBER: _ClassVar[int]
    WORKER_GROUP_SPEC_FIELD_NUMBER: _ClassVar[int]
    ENABLE_AUTOSCALING_FIELD_NUMBER: _ClassVar[int]
    AUTOSCALER_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    head_group_spec: HeadGroupSpec
    worker_group_spec: _containers.RepeatedCompositeFieldContainer[WorkerGroupSpec]
    enable_autoscaling: bool
    autoscaler_options: AutoscalerOptions
    def __init__(self, head_group_spec: _Optional[_Union[HeadGroupSpec, _Mapping]] = ..., worker_group_spec: _Optional[_Iterable[_Union[WorkerGroupSpec, _Mapping]]] = ..., enable_autoscaling: bool = ..., autoscaler_options: _Optional[_Union[AutoscalerOptions, _Mapping]] = ...) -> None: ...

class HeadGroupSpec(_message.Message):
    __slots__ = ["ray_start_params", "k8s_pod"]
    class RayStartParamsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RAY_START_PARAMS_FIELD_NUMBER: _ClassVar[int]
    K8S_POD_FIELD_NUMBER: _ClassVar[int]
    ray_start_params: _containers.ScalarMap[str, str]
    k8s_pod: _tasks_pb2.K8sPod
    def __init__(self, ray_start_params: _Optional[_Mapping[str, str]] = ..., k8s_pod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ...) -> None: ...

class WorkerGroupSpec(_message.Message):
    __slots__ = ["group_name", "replicas", "min_replicas", "max_replicas", "ray_start_params", "k8s_pod"]
    class RayStartParamsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    GROUP_NAME_FIELD_NUMBER: _ClassVar[int]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    MAX_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    RAY_START_PARAMS_FIELD_NUMBER: _ClassVar[int]
    K8S_POD_FIELD_NUMBER: _ClassVar[int]
    group_name: str
    replicas: int
    min_replicas: int
    max_replicas: int
    ray_start_params: _containers.ScalarMap[str, str]
    k8s_pod: _tasks_pb2.K8sPod
    def __init__(self, group_name: _Optional[str] = ..., replicas: _Optional[int] = ..., min_replicas: _Optional[int] = ..., max_replicas: _Optional[int] = ..., ray_start_params: _Optional[_Mapping[str, str]] = ..., k8s_pod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ...) -> None: ...
