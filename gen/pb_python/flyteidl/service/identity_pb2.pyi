from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class UserInfoRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class UserInfoResponse(_message.Message):
    __slots__ = ["email", "family_name", "given_name", "name", "picture", "preferred_username", "subject"]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    FAMILY_NAME_FIELD_NUMBER: _ClassVar[int]
    GIVEN_NAME_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PICTURE_FIELD_NUMBER: _ClassVar[int]
    PREFERRED_USERNAME_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    email: str
    family_name: str
    given_name: str
    name: str
    picture: str
    preferred_username: str
    subject: str
    def __init__(self, subject: _Optional[str] = ..., name: _Optional[str] = ..., preferred_username: _Optional[str] = ..., given_name: _Optional[str] = ..., family_name: _Optional[str] = ..., email: _Optional[str] = ..., picture: _Optional[str] = ...) -> None: ...
