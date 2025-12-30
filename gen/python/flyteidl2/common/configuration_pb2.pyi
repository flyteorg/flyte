from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class AttributesSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    SOURCE_UNSPECIFIED: _ClassVar[AttributesSource]
    GLOBAL: _ClassVar[AttributesSource]
    DOMAIN: _ClassVar[AttributesSource]
    PROJECT: _ClassVar[AttributesSource]
    PROJECT_DOMAIN: _ClassVar[AttributesSource]
    ORG: _ClassVar[AttributesSource]
SOURCE_UNSPECIFIED: AttributesSource
GLOBAL: AttributesSource
DOMAIN: AttributesSource
PROJECT: AttributesSource
PROJECT_DOMAIN: AttributesSource
ORG: AttributesSource
