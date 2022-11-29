from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DistributedTensorflowTrainingTask(_message.Message):
    __slots__ = ["chief_replicas", "ps_replicas", "workers"]
    CHIEF_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    PS_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    chief_replicas: int
    ps_replicas: int
    workers: int
    def __init__(self, workers: _Optional[int] = ..., ps_replicas: _Optional[int] = ..., chief_replicas: _Optional[int] = ...) -> None: ...
