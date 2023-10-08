from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RoleSpec(_message.Message):
    __slots__ = ["pod", "pod_template_name"]
    POD_FIELD_NUMBER: _ClassVar[int]
    POD_TEMPLATE_NAME_FIELD_NUMBER: _ClassVar[int]
    pod: _tasks_pb2.K8sPod
    pod_template_name: str
    def __init__(self, pod: _Optional[_Union[_tasks_pb2.K8sPod, _Mapping]] = ..., pod_template_name: _Optional[str] = ...) -> None: ...
