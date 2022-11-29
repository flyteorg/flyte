from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AddTagRequest(_message.Message):
    __slots__ = ["tag"]
    TAG_FIELD_NUMBER: _ClassVar[int]
    tag: Tag
    def __init__(self, tag: _Optional[_Union[Tag, _Mapping]] = ...) -> None: ...

class AddTagResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Artifact(_message.Message):
    __slots__ = ["created_at", "data", "dataset", "id", "metadata", "partitions", "tags"]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    created_at: _timestamp_pb2.Timestamp
    data: _containers.RepeatedCompositeFieldContainer[ArtifactData]
    dataset: DatasetID
    id: str
    metadata: Metadata
    partitions: _containers.RepeatedCompositeFieldContainer[Partition]
    tags: _containers.RepeatedCompositeFieldContainer[Tag]
    def __init__(self, id: _Optional[str] = ..., dataset: _Optional[_Union[DatasetID, _Mapping]] = ..., data: _Optional[_Iterable[_Union[ArtifactData, _Mapping]]] = ..., metadata: _Optional[_Union[Metadata, _Mapping]] = ..., partitions: _Optional[_Iterable[_Union[Partition, _Mapping]]] = ..., tags: _Optional[_Iterable[_Union[Tag, _Mapping]]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class ArtifactData(_message.Message):
    __slots__ = ["name", "value"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    value: _literals_pb2.Literal
    def __init__(self, name: _Optional[str] = ..., value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ...) -> None: ...

class ArtifactPropertyFilter(_message.Message):
    __slots__ = ["artifact_id"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    def __init__(self, artifact_id: _Optional[str] = ...) -> None: ...

class CreateArtifactRequest(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class CreateArtifactResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class CreateDatasetRequest(_message.Message):
    __slots__ = ["dataset"]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    dataset: Dataset
    def __init__(self, dataset: _Optional[_Union[Dataset, _Mapping]] = ...) -> None: ...

class CreateDatasetResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Dataset(_message.Message):
    __slots__ = ["id", "metadata", "partitionKeys"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    PARTITIONKEYS_FIELD_NUMBER: _ClassVar[int]
    id: DatasetID
    metadata: Metadata
    partitionKeys: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[_Union[DatasetID, _Mapping]] = ..., metadata: _Optional[_Union[Metadata, _Mapping]] = ..., partitionKeys: _Optional[_Iterable[str]] = ...) -> None: ...

class DatasetID(_message.Message):
    __slots__ = ["UUID", "domain", "name", "project", "version"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    UUID: str
    UUID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    domain: str
    name: str
    project: str
    version: str
    def __init__(self, project: _Optional[str] = ..., name: _Optional[str] = ..., domain: _Optional[str] = ..., version: _Optional[str] = ..., UUID: _Optional[str] = ...) -> None: ...

class DatasetPropertyFilter(_message.Message):
    __slots__ = ["domain", "name", "project", "version"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    domain: str
    name: str
    project: str
    version: str
    def __init__(self, project: _Optional[str] = ..., name: _Optional[str] = ..., domain: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class FilterExpression(_message.Message):
    __slots__ = ["filters"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    filters: _containers.RepeatedCompositeFieldContainer[SinglePropertyFilter]
    def __init__(self, filters: _Optional[_Iterable[_Union[SinglePropertyFilter, _Mapping]]] = ...) -> None: ...

class GetArtifactRequest(_message.Message):
    __slots__ = ["artifact_id", "dataset", "tag_name"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    TAG_NAME_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    dataset: DatasetID
    tag_name: str
    def __init__(self, dataset: _Optional[_Union[DatasetID, _Mapping]] = ..., artifact_id: _Optional[str] = ..., tag_name: _Optional[str] = ...) -> None: ...

class GetArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: Artifact
    def __init__(self, artifact: _Optional[_Union[Artifact, _Mapping]] = ...) -> None: ...

class GetDatasetRequest(_message.Message):
    __slots__ = ["dataset"]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    dataset: DatasetID
    def __init__(self, dataset: _Optional[_Union[DatasetID, _Mapping]] = ...) -> None: ...

class GetDatasetResponse(_message.Message):
    __slots__ = ["dataset"]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    dataset: Dataset
    def __init__(self, dataset: _Optional[_Union[Dataset, _Mapping]] = ...) -> None: ...

class GetOrExtendReservationRequest(_message.Message):
    __slots__ = ["heartbeat_interval", "owner_id", "reservation_id"]
    HEARTBEAT_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_ID_FIELD_NUMBER: _ClassVar[int]
    heartbeat_interval: _duration_pb2.Duration
    owner_id: str
    reservation_id: ReservationID
    def __init__(self, reservation_id: _Optional[_Union[ReservationID, _Mapping]] = ..., owner_id: _Optional[str] = ..., heartbeat_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class GetOrExtendReservationResponse(_message.Message):
    __slots__ = ["reservation"]
    RESERVATION_FIELD_NUMBER: _ClassVar[int]
    reservation: Reservation
    def __init__(self, reservation: _Optional[_Union[Reservation, _Mapping]] = ...) -> None: ...

class KeyValuePair(_message.Message):
    __slots__ = ["key", "value"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    key: str
    value: str
    def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...

class ListArtifactsRequest(_message.Message):
    __slots__ = ["dataset", "filter", "pagination"]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    PAGINATION_FIELD_NUMBER: _ClassVar[int]
    dataset: DatasetID
    filter: FilterExpression
    pagination: PaginationOptions
    def __init__(self, dataset: _Optional[_Union[DatasetID, _Mapping]] = ..., filter: _Optional[_Union[FilterExpression, _Mapping]] = ..., pagination: _Optional[_Union[PaginationOptions, _Mapping]] = ...) -> None: ...

class ListArtifactsResponse(_message.Message):
    __slots__ = ["artifacts", "next_token"]
    ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    NEXT_TOKEN_FIELD_NUMBER: _ClassVar[int]
    artifacts: _containers.RepeatedCompositeFieldContainer[Artifact]
    next_token: str
    def __init__(self, artifacts: _Optional[_Iterable[_Union[Artifact, _Mapping]]] = ..., next_token: _Optional[str] = ...) -> None: ...

class ListDatasetsRequest(_message.Message):
    __slots__ = ["filter", "pagination"]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    PAGINATION_FIELD_NUMBER: _ClassVar[int]
    filter: FilterExpression
    pagination: PaginationOptions
    def __init__(self, filter: _Optional[_Union[FilterExpression, _Mapping]] = ..., pagination: _Optional[_Union[PaginationOptions, _Mapping]] = ...) -> None: ...

class ListDatasetsResponse(_message.Message):
    __slots__ = ["datasets", "next_token"]
    DATASETS_FIELD_NUMBER: _ClassVar[int]
    NEXT_TOKEN_FIELD_NUMBER: _ClassVar[int]
    datasets: _containers.RepeatedCompositeFieldContainer[Dataset]
    next_token: str
    def __init__(self, datasets: _Optional[_Iterable[_Union[Dataset, _Mapping]]] = ..., next_token: _Optional[str] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ["key_map"]
    class KeyMapEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    KEY_MAP_FIELD_NUMBER: _ClassVar[int]
    key_map: _containers.ScalarMap[str, str]
    def __init__(self, key_map: _Optional[_Mapping[str, str]] = ...) -> None: ...

class PaginationOptions(_message.Message):
    __slots__ = ["limit", "sortKey", "sortOrder", "token"]
    class SortKey(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    class SortOrder(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ASCENDING: PaginationOptions.SortOrder
    CREATION_TIME: PaginationOptions.SortKey
    DESCENDING: PaginationOptions.SortOrder
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    SORTKEY_FIELD_NUMBER: _ClassVar[int]
    SORTORDER_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    limit: int
    sortKey: PaginationOptions.SortKey
    sortOrder: PaginationOptions.SortOrder
    token: str
    def __init__(self, limit: _Optional[int] = ..., token: _Optional[str] = ..., sortKey: _Optional[_Union[PaginationOptions.SortKey, str]] = ..., sortOrder: _Optional[_Union[PaginationOptions.SortOrder, str]] = ...) -> None: ...

class Partition(_message.Message):
    __slots__ = ["key", "value"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    key: str
    value: str
    def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...

class PartitionPropertyFilter(_message.Message):
    __slots__ = ["key_val"]
    KEY_VAL_FIELD_NUMBER: _ClassVar[int]
    key_val: KeyValuePair
    def __init__(self, key_val: _Optional[_Union[KeyValuePair, _Mapping]] = ...) -> None: ...

class ReleaseReservationRequest(_message.Message):
    __slots__ = ["owner_id", "reservation_id"]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_ID_FIELD_NUMBER: _ClassVar[int]
    owner_id: str
    reservation_id: ReservationID
    def __init__(self, reservation_id: _Optional[_Union[ReservationID, _Mapping]] = ..., owner_id: _Optional[str] = ...) -> None: ...

class ReleaseReservationResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Reservation(_message.Message):
    __slots__ = ["expires_at", "heartbeat_interval", "metadata", "owner_id", "reservation_id"]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    OWNER_ID_FIELD_NUMBER: _ClassVar[int]
    RESERVATION_ID_FIELD_NUMBER: _ClassVar[int]
    expires_at: _timestamp_pb2.Timestamp
    heartbeat_interval: _duration_pb2.Duration
    metadata: Metadata
    owner_id: str
    reservation_id: ReservationID
    def __init__(self, reservation_id: _Optional[_Union[ReservationID, _Mapping]] = ..., owner_id: _Optional[str] = ..., heartbeat_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., metadata: _Optional[_Union[Metadata, _Mapping]] = ...) -> None: ...

class ReservationID(_message.Message):
    __slots__ = ["dataset_id", "tag_name"]
    DATASET_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_NAME_FIELD_NUMBER: _ClassVar[int]
    dataset_id: DatasetID
    tag_name: str
    def __init__(self, dataset_id: _Optional[_Union[DatasetID, _Mapping]] = ..., tag_name: _Optional[str] = ...) -> None: ...

class SinglePropertyFilter(_message.Message):
    __slots__ = ["artifact_filter", "dataset_filter", "operator", "partition_filter", "tag_filter"]
    class ComparisonOperator(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ARTIFACT_FILTER_FIELD_NUMBER: _ClassVar[int]
    DATASET_FILTER_FIELD_NUMBER: _ClassVar[int]
    EQUALS: SinglePropertyFilter.ComparisonOperator
    OPERATOR_FIELD_NUMBER: _ClassVar[int]
    PARTITION_FILTER_FIELD_NUMBER: _ClassVar[int]
    TAG_FILTER_FIELD_NUMBER: _ClassVar[int]
    artifact_filter: ArtifactPropertyFilter
    dataset_filter: DatasetPropertyFilter
    operator: SinglePropertyFilter.ComparisonOperator
    partition_filter: PartitionPropertyFilter
    tag_filter: TagPropertyFilter
    def __init__(self, tag_filter: _Optional[_Union[TagPropertyFilter, _Mapping]] = ..., partition_filter: _Optional[_Union[PartitionPropertyFilter, _Mapping]] = ..., artifact_filter: _Optional[_Union[ArtifactPropertyFilter, _Mapping]] = ..., dataset_filter: _Optional[_Union[DatasetPropertyFilter, _Mapping]] = ..., operator: _Optional[_Union[SinglePropertyFilter.ComparisonOperator, str]] = ...) -> None: ...

class Tag(_message.Message):
    __slots__ = ["artifact_id", "dataset", "name"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    dataset: DatasetID
    name: str
    def __init__(self, name: _Optional[str] = ..., artifact_id: _Optional[str] = ..., dataset: _Optional[_Union[DatasetID, _Mapping]] = ...) -> None: ...

class TagPropertyFilter(_message.Message):
    __slots__ = ["tag_name"]
    TAG_NAME_FIELD_NUMBER: _ClassVar[int]
    tag_name: str
    def __init__(self, tag_name: _Optional[str] = ...) -> None: ...

class UpdateArtifactRequest(_message.Message):
    __slots__ = ["artifact_id", "data", "dataset", "tag_name"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    DATASET_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    TAG_NAME_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    data: _containers.RepeatedCompositeFieldContainer[ArtifactData]
    dataset: DatasetID
    tag_name: str
    def __init__(self, dataset: _Optional[_Union[DatasetID, _Mapping]] = ..., artifact_id: _Optional[str] = ..., tag_name: _Optional[str] = ..., data: _Optional[_Iterable[_Union[ArtifactData, _Mapping]]] = ...) -> None: ...

class UpdateArtifactResponse(_message.Message):
    __slots__ = ["artifact_id"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_id: str
    def __init__(self, artifact_id: _Optional[str] = ...) -> None: ...
