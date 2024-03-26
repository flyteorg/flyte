from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class EmailMessage(_message.Message):
    __slots__ = ["recipients_email", "sender_email", "subject_line", "body"]
    RECIPIENTS_EMAIL_FIELD_NUMBER: _ClassVar[int]
    SENDER_EMAIL_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_LINE_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    recipients_email: _containers.RepeatedScalarFieldContainer[str]
    sender_email: str
    subject_line: str
    body: str
    def __init__(self, recipients_email: _Optional[_Iterable[str]] = ..., sender_email: _Optional[str] = ..., subject_line: _Optional[str] = ..., body: _Optional[str] = ...) -> None: ...
