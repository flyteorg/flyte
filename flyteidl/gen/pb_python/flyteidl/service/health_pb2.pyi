from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HealthRequest(_message.Message):
    __slots__ = ["service"]
    SERVICE_FIELD_NUMBER: _ClassVar[int]
    service: str
    def __init__(self, service: _Optional[str] = ...) -> None: ...

class HealthResponse(_message.Message):
    __slots__ = ["status"]
    class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        UNKNOWN: _ClassVar[HealthResponse.Status]
        SERVING: _ClassVar[HealthResponse.Status]
        NOT_SERVING: _ClassVar[HealthResponse.Status]
        SERVICE_UNKNOWN: _ClassVar[HealthResponse.Status]
    UNKNOWN: HealthResponse.Status
    SERVING: HealthResponse.Status
    NOT_SERVING: HealthResponse.Status
    SERVICE_UNKNOWN: HealthResponse.Status
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: HealthResponse.Status
    def __init__(self, status: _Optional[_Union[HealthResponse.Status, str]] = ...) -> None: ...
