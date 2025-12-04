from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DistributedMPITrainingTask(_message.Message):
    __slots__ = ["num_workers", "num_launcher_replicas", "slots"]
    NUM_WORKERS_FIELD_NUMBER: _ClassVar[int]
    NUM_LAUNCHER_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    SLOTS_FIELD_NUMBER: _ClassVar[int]
    num_workers: int
    num_launcher_replicas: int
    slots: int
    def __init__(self, num_workers: _Optional[int] = ..., num_launcher_replicas: _Optional[int] = ..., slots: _Optional[int] = ...) -> None: ...
