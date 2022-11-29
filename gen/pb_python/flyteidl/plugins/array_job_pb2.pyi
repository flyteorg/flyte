from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ArrayJob(_message.Message):
    __slots__ = ["min_success_ratio", "min_successes", "parallelism", "size"]
    MIN_SUCCESSES_FIELD_NUMBER: _ClassVar[int]
    MIN_SUCCESS_RATIO_FIELD_NUMBER: _ClassVar[int]
    PARALLELISM_FIELD_NUMBER: _ClassVar[int]
    SIZE_FIELD_NUMBER: _ClassVar[int]
    min_success_ratio: float
    min_successes: int
    parallelism: int
    size: int
    def __init__(self, parallelism: _Optional[int] = ..., size: _Optional[int] = ..., min_successes: _Optional[int] = ..., min_success_ratio: _Optional[float] = ...) -> None: ...
