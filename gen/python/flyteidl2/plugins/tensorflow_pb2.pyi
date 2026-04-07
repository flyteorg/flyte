from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DistributedTensorflowTrainingTask(_message.Message):
    __slots__ = ["workers", "ps_replicas", "chief_replicas", "evaluator_replicas"]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    PS_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    CHIEF_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    EVALUATOR_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    workers: int
    ps_replicas: int
    chief_replicas: int
    evaluator_replicas: int
    def __init__(self, workers: _Optional[int] = ..., ps_replicas: _Optional[int] = ..., chief_replicas: _Optional[int] = ..., evaluator_replicas: _Optional[int] = ...) -> None: ...
