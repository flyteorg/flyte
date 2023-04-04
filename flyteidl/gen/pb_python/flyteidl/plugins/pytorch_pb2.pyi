from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DistributedPyTorchTrainingTask(_message.Message):
    __slots__ = ["workers"]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    workers: int
    def __init__(self, workers: _Optional[int] = ...) -> None: ...
