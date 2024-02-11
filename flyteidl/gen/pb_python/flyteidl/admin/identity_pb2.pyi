from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class User(_message.Message):
    __slots__ = ["first_name", "last_name", "email", "subject"]
    FIRST_NAME_FIELD_NUMBER: _ClassVar[int]
    LAST_NAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    first_name: str
    last_name: str
    email: str
    subject: str
    def __init__(self, first_name: _Optional[str] = ..., last_name: _Optional[str] = ..., email: _Optional[str] = ..., subject: _Optional[str] = ...) -> None: ...

class Application(_message.Message):
    __slots__ = ["subject"]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    subject: str
    def __init__(self, subject: _Optional[str] = ...) -> None: ...

class Identity(_message.Message):
    __slots__ = ["user", "application"]
    USER_FIELD_NUMBER: _ClassVar[int]
    APPLICATION_FIELD_NUMBER: _ClassVar[int]
    user: User
    application: Application
    def __init__(self, user: _Optional[_Union[User, _Mapping]] = ..., application: _Optional[_Union[Application, _Mapping]] = ...) -> None: ...
