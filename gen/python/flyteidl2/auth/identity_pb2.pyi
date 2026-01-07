from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UserInfoRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UserInfoResponse(_message.Message):
    __slots__ = ["subject", "name", "preferred_username", "given_name", "family_name", "email", "picture", "additional_claims"]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PREFERRED_USERNAME_FIELD_NUMBER: _ClassVar[int]
    GIVEN_NAME_FIELD_NUMBER: _ClassVar[int]
    FAMILY_NAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PICTURE_FIELD_NUMBER: _ClassVar[int]
    ADDITIONAL_CLAIMS_FIELD_NUMBER: _ClassVar[int]
    subject: str
    name: str
    preferred_username: str
    given_name: str
    family_name: str
    email: str
    picture: str
    additional_claims: _struct_pb2.Struct
    def __init__(self, subject: _Optional[str] = ..., name: _Optional[str] = ..., preferred_username: _Optional[str] = ..., given_name: _Optional[str] = ..., family_name: _Optional[str] = ..., email: _Optional[str] = ..., picture: _Optional[str] = ..., additional_claims: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
