from buf.validate import validate_pb2 as _validate_pb2
from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from protoc_gen_openapiv2.options import annotations_pb2 as _annotations_pb2_1
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateUploadLocationRequest(_message.Message):
    __slots__ = ["project", "domain", "filename", "expires_in", "content_md5", "filename_root", "add_content_md5_metadata", "org", "content_length"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    CONTENT_MD5_FIELD_NUMBER: _ClassVar[int]
    FILENAME_ROOT_FIELD_NUMBER: _ClassVar[int]
    ADD_CONTENT_MD5_METADATA_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CONTENT_LENGTH_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    filename: str
    expires_in: _duration_pb2.Duration
    content_md5: bytes
    filename_root: str
    add_content_md5_metadata: bool
    org: str
    content_length: int
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., filename: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., content_md5: _Optional[bytes] = ..., filename_root: _Optional[str] = ..., add_content_md5_metadata: bool = ..., org: _Optional[str] = ..., content_length: _Optional[int] = ...) -> None: ...

class CreateUploadLocationResponse(_message.Message):
    __slots__ = ["signed_url", "native_url", "expires_at", "headers"]
    class HeadersEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    NATIVE_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    HEADERS_FIELD_NUMBER: _ClassVar[int]
    signed_url: str
    native_url: str
    expires_at: _timestamp_pb2.Timestamp
    headers: _containers.ScalarMap[str, str]
    def __init__(self, signed_url: _Optional[str] = ..., native_url: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., headers: _Optional[_Mapping[str, str]] = ...) -> None: ...
