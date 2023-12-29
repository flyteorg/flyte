from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetVersionResponse(_message.Message):
    __slots__ = ["control_plane_version"]
    CONTROL_PLANE_VERSION_FIELD_NUMBER: _ClassVar[int]
    control_plane_version: Version
    def __init__(self, control_plane_version: _Optional[_Union[Version, _Mapping]] = ...) -> None: ...

class Version(_message.Message):
    __slots__ = ["Build", "Version", "BuildTime"]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    BUILDTIME_FIELD_NUMBER: _ClassVar[int]
    Build: str
    Version: str
    BuildTime: str
    def __init__(self, Build: _Optional[str] = ..., Version: _Optional[str] = ..., BuildTime: _Optional[str] = ...) -> None: ...

class GetVersionRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
