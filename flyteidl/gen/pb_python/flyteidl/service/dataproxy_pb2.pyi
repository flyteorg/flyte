from google.api import annotations_pb2 as _annotations_pb2
from protoc_gen_openapiv2.options import annotations_pb2 as _annotations_pb2_1
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ArtifactType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ARTIFACT_TYPE_UNDEFINED: _ClassVar[ArtifactType]
    ARTIFACT_TYPE_DECK: _ClassVar[ArtifactType]
ARTIFACT_TYPE_UNDEFINED: ArtifactType
ARTIFACT_TYPE_DECK: ArtifactType

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

class CreateUploadLocationRequest(_message.Message):
    __slots__ = ["project", "domain", "filename", "expires_in", "content_md5", "filename_root", "add_content_md5_metadata", "org"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    CONTENT_MD5_FIELD_NUMBER: _ClassVar[int]
    FILENAME_ROOT_FIELD_NUMBER: _ClassVar[int]
    ADD_CONTENT_MD5_METADATA_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    filename: str
    expires_in: _duration_pb2.Duration
    content_md5: bytes
    filename_root: str
    add_content_md5_metadata: bool
    org: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., filename: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., content_md5: _Optional[bytes] = ..., filename_root: _Optional[str] = ..., add_content_md5_metadata: bool = ..., org: _Optional[str] = ...) -> None: ...

class CreateDownloadLocationRequest(_message.Message):
    __slots__ = ["native_url", "expires_in"]
    NATIVE_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    native_url: str
    expires_in: _duration_pb2.Duration
    def __init__(self, native_url: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class CreateDownloadLocationResponse(_message.Message):
    __slots__ = ["signed_url", "expires_at"]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    signed_url: str
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, signed_url: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

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
    __slots__ = ["signed_url", "expires_at", "pre_signed_urls"]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    PRE_SIGNED_URLS_FIELD_NUMBER: _ClassVar[int]
    signed_url: _containers.RepeatedScalarFieldContainer[str]
    expires_at: _timestamp_pb2.Timestamp
    pre_signed_urls: PreSignedURLs
    def __init__(self, signed_url: _Optional[_Iterable[str]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., pre_signed_urls: _Optional[_Union[PreSignedURLs, _Mapping]] = ...) -> None: ...

class PreSignedURLs(_message.Message):
    __slots__ = ["signed_url", "expires_at"]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    signed_url: _containers.RepeatedScalarFieldContainer[str]
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, signed_url: _Optional[_Iterable[str]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetDataRequest(_message.Message):
    __slots__ = ["flyte_url"]
    FLYTE_URL_FIELD_NUMBER: _ClassVar[int]
    flyte_url: str
    def __init__(self, flyte_url: _Optional[str] = ...) -> None: ...

class GetDataResponse(_message.Message):
    __slots__ = ["literal_map", "pre_signed_urls", "literal"]
    LITERAL_MAP_FIELD_NUMBER: _ClassVar[int]
    PRE_SIGNED_URLS_FIELD_NUMBER: _ClassVar[int]
    LITERAL_FIELD_NUMBER: _ClassVar[int]
    literal_map: _literals_pb2.LiteralMap
    pre_signed_urls: PreSignedURLs
    literal: _literals_pb2.Literal
    def __init__(self, literal_map: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., pre_signed_urls: _Optional[_Union[PreSignedURLs, _Mapping]] = ..., literal: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...
