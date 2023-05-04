from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.plugins.kubeflow import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DistributedMPITrainingTask(_message.Message):
    __slots__ = ["worker_replicas", "launcher_replicas", "run_policy", "slots"]
    WORKER_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    LAUNCHER_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    RUN_POLICY_FIELD_NUMBER: _ClassVar[int]
    SLOTS_FIELD_NUMBER: _ClassVar[int]
    worker_replicas: DistributedMPITrainingReplicaSpec
    launcher_replicas: DistributedMPITrainingReplicaSpec
    run_policy: _common_pb2.RunPolicy
    slots: int
    def __init__(self, worker_replicas: _Optional[_Union[DistributedMPITrainingReplicaSpec, _Mapping]] = ..., launcher_replicas: _Optional[_Union[DistributedMPITrainingReplicaSpec, _Mapping]] = ..., run_policy: _Optional[_Union[_common_pb2.RunPolicy, _Mapping]] = ..., slots: _Optional[int] = ...) -> None: ...

class DistributedMPITrainingReplicaSpec(_message.Message):
    __slots__ = ["replicas", "image", "resources", "restart_policy", "command"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    RESTART_POLICY_FIELD_NUMBER: _ClassVar[int]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    replicas: int
    image: str
    resources: _tasks_pb2.Resources
    restart_policy: _common_pb2.RestartPolicy
    command: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, replicas: _Optional[int] = ..., image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ..., restart_policy: _Optional[_Union[_common_pb2.RestartPolicy, str]] = ..., command: _Optional[_Iterable[str]] = ...) -> None: ...
