from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class HiveQuery(_message.Message):
    __slots__ = ["query", "timeout_sec", "retryCount"]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_SEC_FIELD_NUMBER: _ClassVar[int]
    RETRYCOUNT_FIELD_NUMBER: _ClassVar[int]
    query: str
    timeout_sec: int
    retryCount: int
    def __init__(self, query: _Optional[str] = ..., timeout_sec: _Optional[int] = ..., retryCount: _Optional[int] = ...) -> None: ...

class HiveQueryCollection(_message.Message):
    __slots__ = ["queries"]
    QUERIES_FIELD_NUMBER: _ClassVar[int]
    queries: _containers.RepeatedCompositeFieldContainer[HiveQuery]
    def __init__(self, queries: _Optional[_Iterable[_Union[HiveQuery, _Mapping]]] = ...) -> None: ...

class QuboleHiveJob(_message.Message):
    __slots__ = ["cluster_label", "query_collection", "tags", "query"]
    CLUSTER_LABEL_FIELD_NUMBER: _ClassVar[int]
    QUERY_COLLECTION_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    cluster_label: str
    query_collection: HiveQueryCollection
    tags: _containers.RepeatedScalarFieldContainer[str]
    query: HiveQuery
    def __init__(self, cluster_label: _Optional[str] = ..., query_collection: _Optional[_Union[HiveQueryCollection, _Mapping]] = ..., tags: _Optional[_Iterable[str]] = ..., query: _Optional[_Union[HiveQuery, _Mapping]] = ...) -> None: ...
