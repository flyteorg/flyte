from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.plugins.kubeflow import common_pb2 as _common_pb2
from flyteidl.plugins import common_pb2 as _common_pb2_1
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ElasticConfig(_message.Message):
    __slots__ = ["rdzv_backend", "min_replicas", "max_replicas", "nproc_per_node", "max_restarts"]
    RDZV_BACKEND_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    MAX_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    NPROC_PER_NODE_FIELD_NUMBER: _ClassVar[int]
    MAX_RESTARTS_FIELD_NUMBER: _ClassVar[int]
    rdzv_backend: str
    min_replicas: int
    max_replicas: int
    nproc_per_node: int
    max_restarts: int
    def __init__(self, rdzv_backend: _Optional[str] = ..., min_replicas: _Optional[int] = ..., max_replicas: _Optional[int] = ..., nproc_per_node: _Optional[int] = ..., max_restarts: _Optional[int] = ...) -> None: ...

class DistributedPyTorchTrainingTask(_message.Message):
    __slots__ = ["worker_replicas", "master_replicas", "run_policy", "elastic_config"]
    WORKER_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    MASTER_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    RUN_POLICY_FIELD_NUMBER: _ClassVar[int]
    ELASTIC_CONFIG_FIELD_NUMBER: _ClassVar[int]
    worker_replicas: DistributedPyTorchTrainingReplicaSpec
    master_replicas: DistributedPyTorchTrainingReplicaSpec
    run_policy: _common_pb2.RunPolicy
    elastic_config: ElasticConfig
    def __init__(self, worker_replicas: _Optional[_Union[DistributedPyTorchTrainingReplicaSpec, _Mapping]] = ..., master_replicas: _Optional[_Union[DistributedPyTorchTrainingReplicaSpec, _Mapping]] = ..., run_policy: _Optional[_Union[_common_pb2.RunPolicy, _Mapping]] = ..., elastic_config: _Optional[_Union[ElasticConfig, _Mapping]] = ...) -> None: ...

class DistributedPyTorchTrainingReplicaSpec(_message.Message):
    __slots__ = ["replicas", "image", "resources", "restart_policy", "common"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    RESTART_POLICY_FIELD_NUMBER: _ClassVar[int]
    COMMON_FIELD_NUMBER: _ClassVar[int]
    replicas: int
    image: str
    resources: _tasks_pb2.Resources
    restart_policy: _common_pb2_1.RestartPolicy
    common: _common_pb2_1.CommonReplicaSpec
    def __init__(self, replicas: _Optional[int] = ..., image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ..., restart_policy: _Optional[_Union[_common_pb2_1.RestartPolicy, str]] = ..., common: _Optional[_Union[_common_pb2_1.CommonReplicaSpec, _Mapping]] = ...) -> None: ...
