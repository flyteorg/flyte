from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EnvironmentType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    FAST_TASK: _ClassVar[EnvironmentType]
FAST_TASK: EnvironmentType

class ExecutionEnvironmentAssignment(_message.Message):
    __slots__ = ["id", "node_ids", "environment", "environment_spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NODE_IDS_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_SPEC_FIELD_NUMBER: _ClassVar[int]
    id: str
    node_ids: _containers.RepeatedScalarFieldContainer[str]
    environment: ExecutionEnvironment
    environment_spec: ExecutionEnvironmentSpec
    def __init__(self, id: _Optional[str] = ..., node_ids: _Optional[_Iterable[str]] = ..., environment: _Optional[_Union[ExecutionEnvironment, _Mapping]] = ..., environment_spec: _Optional[_Union[ExecutionEnvironmentSpec, _Mapping]] = ...) -> None: ...

class ExecutionEnvironment(_message.Message):
    __slots__ = ["type", "fast_task"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FAST_TASK_FIELD_NUMBER: _ClassVar[int]
    type: EnvironmentType
    fast_task: FastTaskEnvironment
    def __init__(self, type: _Optional[_Union[EnvironmentType, str]] = ..., fast_task: _Optional[_Union[FastTaskEnvironment, _Mapping]] = ...) -> None: ...

class FastTaskEnvironment(_message.Message):
    __slots__ = ["queue_id", "namespace", "pod_name"]
    QUEUE_ID_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    POD_NAME_FIELD_NUMBER: _ClassVar[int]
    queue_id: str
    namespace: str
    pod_name: str
    def __init__(self, queue_id: _Optional[str] = ..., namespace: _Optional[str] = ..., pod_name: _Optional[str] = ...) -> None: ...

class ExecutionEnvironmentSpec(_message.Message):
    __slots__ = ["type", "fast_task"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FAST_TASK_FIELD_NUMBER: _ClassVar[int]
    type: EnvironmentType
    fast_task: FastTaskEnvironmentSpec
    def __init__(self, type: _Optional[_Union[EnvironmentType, str]] = ..., fast_task: _Optional[_Union[FastTaskEnvironmentSpec, _Mapping]] = ...) -> None: ...

class FastTaskEnvironmentSpec(_message.Message):
    __slots__ = ["image", "replica_count"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    REPLICA_COUNT_FIELD_NUMBER: _ClassVar[int]
    image: str
    replica_count: int
    def __init__(self, image: _Optional[str] = ..., replica_count: _Optional[int] = ...) -> None: ...
