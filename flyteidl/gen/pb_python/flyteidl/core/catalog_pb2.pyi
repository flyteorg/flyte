from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CatalogCacheStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    CACHE_DISABLED: _ClassVar[CatalogCacheStatus]
    CACHE_MISS: _ClassVar[CatalogCacheStatus]
    CACHE_HIT: _ClassVar[CatalogCacheStatus]
    CACHE_POPULATED: _ClassVar[CatalogCacheStatus]
    CACHE_LOOKUP_FAILURE: _ClassVar[CatalogCacheStatus]
    CACHE_PUT_FAILURE: _ClassVar[CatalogCacheStatus]
    CACHE_SKIPPED: _ClassVar[CatalogCacheStatus]
    CACHE_EVICTED: _ClassVar[CatalogCacheStatus]
CACHE_DISABLED: CatalogCacheStatus
CACHE_MISS: CatalogCacheStatus
CACHE_HIT: CatalogCacheStatus
CACHE_POPULATED: CatalogCacheStatus
CACHE_LOOKUP_FAILURE: CatalogCacheStatus
CACHE_PUT_FAILURE: CatalogCacheStatus
CACHE_SKIPPED: CatalogCacheStatus
CACHE_EVICTED: CatalogCacheStatus

class CatalogArtifactTag(_message.Message):
    __slots__ = ["artifact_id", "name"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    name: str
    def __init__(self, artifact_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class CatalogMetadata(_message.Message):
    __slots__ = ["dataset_id", "artifact_tag", "source_task_execution"]
    DATASET_ID_FIELD_NUMBER: _ClassVar[int]
    ARTIFACT_TAG_FIELD_NUMBER: _ClassVar[int]
    SOURCE_TASK_EXECUTION_FIELD_NUMBER: _ClassVar[int]
    dataset_id: _identifier_pb2.Identifier
    artifact_tag: CatalogArtifactTag
    source_task_execution: _identifier_pb2.TaskExecutionIdentifier
    def __init__(self, dataset_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., artifact_tag: _Optional[_Union[CatalogArtifactTag, _Mapping]] = ..., source_task_execution: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class CatalogReservation(_message.Message):
    __slots__ = []
    class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        RESERVATION_DISABLED: _ClassVar[CatalogReservation.Status]
        RESERVATION_ACQUIRED: _ClassVar[CatalogReservation.Status]
        RESERVATION_EXISTS: _ClassVar[CatalogReservation.Status]
        RESERVATION_RELEASED: _ClassVar[CatalogReservation.Status]
        RESERVATION_FAILURE: _ClassVar[CatalogReservation.Status]
    RESERVATION_DISABLED: CatalogReservation.Status
    RESERVATION_ACQUIRED: CatalogReservation.Status
    RESERVATION_EXISTS: CatalogReservation.Status
    RESERVATION_RELEASED: CatalogReservation.Status
    RESERVATION_FAILURE: CatalogReservation.Status
    def __init__(self) -> None: ...
