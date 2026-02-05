from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.imagebuilder import definition_pb2 as _definition_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetImageRequest(_message.Message):
    __slots__ = ["id", "organization", "project_id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    id: _definition_pb2.ImageIdentifier
    organization: str
    project_id: _identifier_pb2.ProjectIdentifier
    def __init__(self, id: _Optional[_Union[_definition_pb2.ImageIdentifier, _Mapping]] = ..., organization: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ...) -> None: ...

class GetImageResponse(_message.Message):
    __slots__ = ["image"]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    image: _definition_pb2.Image
    def __init__(self, image: _Optional[_Union[_definition_pb2.Image, _Mapping]] = ...) -> None: ...
