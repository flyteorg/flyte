from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import types_pb2 as _types_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class KeyMapMetadata(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ["source_identifier", "key_map", "created_at", "last_updated_at"]
    SOURCE_IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    KEY_MAP_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    source_identifier: _identifier_pb2.Identifier
    key_map: KeyMapMetadata
    created_at: _timestamp_pb2.Timestamp
    last_updated_at: _timestamp_pb2.Timestamp
    def __init__(self, source_identifier: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., key_map: _Optional[_Union[KeyMapMetadata, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., last_updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CachedOutput(_message.Message):
    __slots__ = ["output_literals", "output_uri", "metadata"]
    OUTPUT_LITERALS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    output_literals: _literals_pb2.LiteralMap
    output_uri: str
    metadata: Metadata
    def __init__(self, output_literals: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., output_uri: _Optional[str] = ..., metadata: _Optional[_Union[Metadata, _Mapping]] = ...) -> None: ...

class GetCacheRequest(_message.Message):
    __slots__ = ["key"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class GetCacheResponse(_message.Message):
    __slots__ = ["output"]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: CachedOutput
    def __init__(self, output: _Optional[_Union[CachedOutput, _Mapping]] = ...) -> None: ...

class PutCacheRequest(_message.Message):
    __slots__ = ["key", "output", "overwrite"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_FIELD_NUMBER: _ClassVar[int]
    key: str
    output: CachedOutput
    overwrite: bool
    def __init__(self, key: _Optional[str] = ..., output: _Optional[_Union[CachedOutput, _Mapping]] = ..., overwrite: bool = ...) -> None: ...

class PutCacheResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class DeleteCacheRequest(_message.Message):
    __slots__ = ["key"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class DeleteCacheResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Reservation(_message.Message):
    __slots__ = ["key", "owner_id", "heartbeat_interval", "expires_at"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    key: str
    owner_id: str
    heartbeat_interval: _duration_pb2.Duration
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, key: _Optional[str] = ..., owner_id: _Optional[str] = ..., heartbeat_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetOrExtendReservationRequest(_message.Message):
    __slots__ = ["key", "owner_id", "heartbeat_interval"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    key: str
    owner_id: str
    heartbeat_interval: _duration_pb2.Duration
    def __init__(self, key: _Optional[str] = ..., owner_id: _Optional[str] = ..., heartbeat_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class GetOrExtendReservationResponse(_message.Message):
    __slots__ = ["reservation"]
    RESERVATION_FIELD_NUMBER: _ClassVar[int]
    reservation: Reservation
    def __init__(self, reservation: _Optional[_Union[Reservation, _Mapping]] = ...) -> None: ...

class ReleaseReservationRequest(_message.Message):
    __slots__ = ["key", "owner_id"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    key: str
    owner_id: str
    def __init__(self, key: _Optional[str] = ..., owner_id: _Optional[str] = ...) -> None: ...

class ReleaseReservationResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
