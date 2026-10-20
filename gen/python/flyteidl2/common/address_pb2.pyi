from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ActionAddress(_message.Message):
    __slots__ = ["pod", "path", "url", "cluster"]
    POD_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    pod: PodAddress
    path: str
    url: str
    cluster: str
    def __init__(self, pod: _Optional[_Union[PodAddress, _Mapping]] = ..., path: _Optional[str] = ..., url: _Optional[str] = ..., cluster: _Optional[str] = ...) -> None: ...

class PodAddress(_message.Message):
    __slots__ = ["namespace", "name", "ip", "port"]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    namespace: str
    name: str
    ip: str
    port: int
    def __init__(self, namespace: _Optional[str] = ..., name: _Optional[str] = ..., ip: _Optional[str] = ..., port: _Optional[int] = ...) -> None: ...
