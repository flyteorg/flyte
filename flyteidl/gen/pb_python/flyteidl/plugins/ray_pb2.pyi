from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import containers as _containers
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

class RayCluster(_message.Message):
    __slots__ = ["head_group_spec", "worker_group_spec", "enable_autoscaling"]
    HEAD_GROUP_SPEC_FIELD_NUMBER: _ClassVar[int]
    WORKER_GROUP_SPEC_FIELD_NUMBER: _ClassVar[int]
    ENABLE_AUTOSCALING_FIELD_NUMBER: _ClassVar[int]
    head_group_spec: HeadGroupSpec
    worker_group_spec: _containers.RepeatedCompositeFieldContainer[WorkerGroupSpec]
    enable_autoscaling: bool
    def __init__(self, head_group_spec: _Optional[_Union[HeadGroupSpec, _Mapping]] = ..., worker_group_spec: _Optional[_Iterable[_Union[WorkerGroupSpec, _Mapping]]] = ..., enable_autoscaling: bool = ...) -> None: ...

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
