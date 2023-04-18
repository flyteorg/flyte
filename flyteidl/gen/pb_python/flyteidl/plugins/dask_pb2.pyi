from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DaskJob(_message.Message):
    __slots__ = ["scheduler", "workers"]
    SCHEDULER_FIELD_NUMBER: _ClassVar[int]
    WORKERS_FIELD_NUMBER: _ClassVar[int]
    scheduler: DaskScheduler
    workers: DaskWorkerGroup
    def __init__(self, scheduler: _Optional[_Union[DaskScheduler, _Mapping]] = ..., workers: _Optional[_Union[DaskWorkerGroup, _Mapping]] = ...) -> None: ...

class DaskScheduler(_message.Message):
    __slots__ = ["image", "resources"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    image: str
    resources: _tasks_pb2.Resources
    def __init__(self, image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ...) -> None: ...

class DaskWorkerGroup(_message.Message):
    __slots__ = ["number_of_workers", "image", "resources"]
    NUMBER_OF_WORKERS_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    number_of_workers: int
    image: str
    resources: _tasks_pb2.Resources
    def __init__(self, number_of_workers: _Optional[int] = ..., image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ...) -> None: ...
