from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EnvironmentAssignment(_message.Message):
    __slots__ = ["id", "node_ids", "environment"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NODE_IDS_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    id: str
    node_ids: _containers.RepeatedScalarFieldContainer[str]
    environment: Environment
    def __init__(self, id: _Optional[str] = ..., node_ids: _Optional[_Iterable[str]] = ..., environment: _Optional[_Union[Environment, _Mapping]] = ...) -> None: ...

class Environment(_message.Message):
    __slots__ = ["type", "fasttask_environment"]
    class EnvironmentType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        FASTTASK: _ClassVar[Environment.EnvironmentType]
    FASTTASK: Environment.EnvironmentType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FASTTASK_ENVIRONMENT_FIELD_NUMBER: _ClassVar[int]
    type: Environment.EnvironmentType
    fasttask_environment: FastTaskEnvironment
    def __init__(self, type: _Optional[_Union[Environment.EnvironmentType, str]] = ..., fasttask_environment: _Optional[_Union[FastTaskEnvironment, _Mapping]] = ...) -> None: ...

class FastTaskEnvironment(_message.Message):
    __slots__ = ["queue_id", "namespace", "pod_id"]
    QUEUE_ID_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    POD_ID_FIELD_NUMBER: _ClassVar[int]
    queue_id: str
    namespace: str
    pod_id: str
    def __init__(self, queue_id: _Optional[str] = ..., namespace: _Optional[str] = ..., pod_id: _Optional[str] = ...) -> None: ...
