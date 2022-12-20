from google.api import annotations_pb2 as _annotations_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

ARTIFACT_TYPE_DECK: ArtifactType
ARTIFACT_TYPE_UNDEFINED: ArtifactType
DESCRIPTOR: _descriptor.FileDescriptor

class CreateDownloadLinkRequest(_message.Message):
    __slots__ = ["artifact_type", "expires_in", "node_execution_id"]
    ARTIFACT_TYPE_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_type: ArtifactType
    expires_in: _duration_pb2.Duration
    node_execution_id: _identifier_pb2.NodeExecutionIdentifier
    def __init__(self, artifact_type: _Optional[_Union[ArtifactType, str]] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., node_execution_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ...) -> None: ...

class CreateDownloadLinkResponse(_message.Message):
    __slots__ = ["expires_at", "signed_url"]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    expires_at: _timestamp_pb2.Timestamp
    signed_url: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, signed_url: _Optional[_Iterable[str]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreateDownloadLocationRequest(_message.Message):
    __slots__ = ["expires_in", "native_url"]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    NATIVE_URL_FIELD_NUMBER: _ClassVar[int]
    expires_in: _duration_pb2.Duration
    native_url: str
    def __init__(self, native_url: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class CreateDownloadLocationResponse(_message.Message):
    __slots__ = ["expires_at", "signed_url"]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    expires_at: _timestamp_pb2.Timestamp
    signed_url: str
    def __init__(self, signed_url: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreateUploadLocationRequest(_message.Message):
    __slots__ = ["content_md5", "domain", "expires_in", "filename", "project"]
    CONTENT_MD5_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    content_md5: bytes
    domain: str
    expires_in: _duration_pb2.Duration
    filename: str
    project: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., filename: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., content_md5: _Optional[bytes] = ...) -> None: ...

class CreateUploadLocationResponse(_message.Message):
    __slots__ = ["expires_at", "native_url", "signed_url"]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    NATIVE_URL_FIELD_NUMBER: _ClassVar[int]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    expires_at: _timestamp_pb2.Timestamp
    native_url: str
    signed_url: str
    def __init__(self, signed_url: _Optional[str] = ..., native_url: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ArtifactType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
