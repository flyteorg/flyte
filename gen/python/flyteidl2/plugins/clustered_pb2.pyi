from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RdzvBackend(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    STATIC: _ClassVar[RdzvBackend]
    C10D: _ClassVar[RdzvBackend]

class Interconnect(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    TCP: _ClassVar[Interconnect]
    EFA: _ClassVar[Interconnect]
    INFINIBAND: _ClassVar[Interconnect]
    ROCE: _ClassVar[Interconnect]
STATIC: RdzvBackend
C10D: RdzvBackend
TCP: Interconnect
EFA: Interconnect
INFINIBAND: Interconnect
ROCE: Interconnect

class ClusteredTaskSpec(_message.Message):
    __slots__ = ["replicas", "accelerators_per_replica", "runtime", "interconnect", "failure_policy", "ttl_seconds_after_finished"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    ACCELERATORS_PER_REPLICA_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    INTERCONNECT_FIELD_NUMBER: _ClassVar[int]
    FAILURE_POLICY_FIELD_NUMBER: _ClassVar[int]
    TTL_SECONDS_AFTER_FINISHED_FIELD_NUMBER: _ClassVar[int]
    replicas: int
    accelerators_per_replica: int
    runtime: Runtime
    interconnect: Interconnect
    failure_policy: ClusterFailurePolicy
    ttl_seconds_after_finished: _wrappers_pb2.Int32Value
    def __init__(self, replicas: _Optional[int] = ..., accelerators_per_replica: _Optional[int] = ..., runtime: _Optional[_Union[Runtime, _Mapping]] = ..., interconnect: _Optional[_Union[Interconnect, str]] = ..., failure_policy: _Optional[_Union[ClusterFailurePolicy, _Mapping]] = ..., ttl_seconds_after_finished: _Optional[_Union[_wrappers_pb2.Int32Value, _Mapping]] = ...) -> None: ...

class Runtime(_message.Message):
    __slots__ = ["torchrun"]
    TORCHRUN_FIELD_NUMBER: _ClassVar[int]
    torchrun: TorchRunSpec
    def __init__(self, torchrun: _Optional[_Union[TorchRunSpec, _Mapping]] = ...) -> None: ...

class TorchRunSpec(_message.Message):
    __slots__ = ["rdzv_backend", "max_restarts", "master_port"]
    RDZV_BACKEND_FIELD_NUMBER: _ClassVar[int]
    MAX_RESTARTS_FIELD_NUMBER: _ClassVar[int]
    MASTER_PORT_FIELD_NUMBER: _ClassVar[int]
    rdzv_backend: RdzvBackend
    max_restarts: int
    master_port: int
    def __init__(self, rdzv_backend: _Optional[_Union[RdzvBackend, str]] = ..., max_restarts: _Optional[int] = ..., master_port: _Optional[int] = ...) -> None: ...

class ClusterFailurePolicy(_message.Message):
    __slots__ = ["max_restarts", "restart_on_host_maintenance"]
    MAX_RESTARTS_FIELD_NUMBER: _ClassVar[int]
    RESTART_ON_HOST_MAINTENANCE_FIELD_NUMBER: _ClassVar[int]
    max_restarts: int
    restart_on_host_maintenance: bool
    def __init__(self, max_restarts: _Optional[int] = ..., restart_on_host_maintenance: bool = ...) -> None: ...
