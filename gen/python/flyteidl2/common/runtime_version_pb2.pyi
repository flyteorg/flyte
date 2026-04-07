from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RuntimeMetadata(_message.Message):
    __slots__ = ["type", "version", "flavor"]
    class RuntimeType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        OTHER: _ClassVar[RuntimeMetadata.RuntimeType]
        FLYTE_SDK: _ClassVar[RuntimeMetadata.RuntimeType]
        UNION_SDK: _ClassVar[RuntimeMetadata.RuntimeType]
    OTHER: RuntimeMetadata.RuntimeType
    FLYTE_SDK: RuntimeMetadata.RuntimeType
    UNION_SDK: RuntimeMetadata.RuntimeType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    FLAVOR_FIELD_NUMBER: _ClassVar[int]
    type: RuntimeMetadata.RuntimeType
    version: str
    flavor: str
    def __init__(self, type: _Optional[_Union[RuntimeMetadata.RuntimeType, str]] = ..., version: _Optional[str] = ..., flavor: _Optional[str] = ...) -> None: ...
