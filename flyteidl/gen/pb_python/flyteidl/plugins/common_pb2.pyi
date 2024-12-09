from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RestartPolicy(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    RESTART_POLICY_NEVER: _ClassVar[RestartPolicy]
    RESTART_POLICY_ON_FAILURE: _ClassVar[RestartPolicy]
    RESTART_POLICY_ALWAYS: _ClassVar[RestartPolicy]
RESTART_POLICY_NEVER: RestartPolicy
RESTART_POLICY_ON_FAILURE: RestartPolicy
RESTART_POLICY_ALWAYS: RestartPolicy

class CommonReplicaSpec(_message.Message):
    __slots__ = ["replicas", "image", "resources", "restart_policy"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    RESTART_POLICY_FIELD_NUMBER: _ClassVar[int]
    replicas: int
    image: str
    resources: _tasks_pb2.Resources
    restart_policy: RestartPolicy
    def __init__(self, replicas: _Optional[int] = ..., image: _Optional[str] = ..., resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ..., restart_policy: _Optional[_Union[RestartPolicy, str]] = ...) -> None: ...
