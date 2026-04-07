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
    __slots__ = ["workers", "elastic_config"]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    ELASTIC_CONFIG_FIELD_NUMBER: _ClassVar[int]
    workers: int
    elastic_config: ElasticConfig
    def __init__(self, workers: _Optional[int] = ..., elastic_config: _Optional[_Union[ElasticConfig, _Mapping]] = ...) -> None: ...
