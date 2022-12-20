from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from flyteidl.core import types_pb2 as _types_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Binary(_message.Message):
    __slots__ = ["tag", "value"]
    TAG_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    tag: str
    value: bytes
    def __init__(self, value: _Optional[bytes] = ..., tag: _Optional[str] = ...) -> None: ...

class Binding(_message.Message):
    __slots__ = ["binding", "var"]
    BINDING_FIELD_NUMBER: _ClassVar[int]
    VAR_FIELD_NUMBER: _ClassVar[int]
    binding: BindingData
    var: str
    def __init__(self, var: _Optional[str] = ..., binding: _Optional[_Union[BindingData, _Mapping]] = ...) -> None: ...

class BindingData(_message.Message):
    __slots__ = ["collection", "map", "promise", "scalar", "union"]
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    MAP_FIELD_NUMBER: _ClassVar[int]
    PROMISE_FIELD_NUMBER: _ClassVar[int]
    SCALAR_FIELD_NUMBER: _ClassVar[int]
    UNION_FIELD_NUMBER: _ClassVar[int]
    collection: BindingDataCollection
    map: BindingDataMap
    promise: _types_pb2.OutputReference
    scalar: Scalar
    union: UnionInfo
    def __init__(self, scalar: _Optional[_Union[Scalar, _Mapping]] = ..., collection: _Optional[_Union[BindingDataCollection, _Mapping]] = ..., promise: _Optional[_Union[_types_pb2.OutputReference, _Mapping]] = ..., map: _Optional[_Union[BindingDataMap, _Mapping]] = ..., union: _Optional[_Union[UnionInfo, _Mapping]] = ...) -> None: ...

class BindingDataCollection(_message.Message):
    __slots__ = ["bindings"]
    BINDINGS_FIELD_NUMBER: _ClassVar[int]
    bindings: _containers.RepeatedCompositeFieldContainer[BindingData]
    def __init__(self, bindings: _Optional[_Iterable[_Union[BindingData, _Mapping]]] = ...) -> None: ...

class BindingDataMap(_message.Message):
    __slots__ = ["bindings"]
    class BindingsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: BindingData
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[BindingData, _Mapping]] = ...) -> None: ...
    BINDINGS_FIELD_NUMBER: _ClassVar[int]
    bindings: _containers.MessageMap[str, BindingData]
    def __init__(self, bindings: _Optional[_Mapping[str, BindingData]] = ...) -> None: ...

class Blob(_message.Message):
    __slots__ = ["metadata", "uri"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    metadata: BlobMetadata
    uri: str
    def __init__(self, metadata: _Optional[_Union[BlobMetadata, _Mapping]] = ..., uri: _Optional[str] = ...) -> None: ...

class BlobMetadata(_message.Message):
    __slots__ = ["type"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    type: _types_pb2.BlobType
    def __init__(self, type: _Optional[_Union[_types_pb2.BlobType, _Mapping]] = ...) -> None: ...

class KeyValuePair(_message.Message):
    __slots__ = ["key", "value"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    key: str
    value: str
    def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...

class Literal(_message.Message):
    __slots__ = ["collection", "hash", "map", "scalar"]
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    MAP_FIELD_NUMBER: _ClassVar[int]
    SCALAR_FIELD_NUMBER: _ClassVar[int]
    collection: LiteralCollection
    hash: str
    map: LiteralMap
    scalar: Scalar
    def __init__(self, scalar: _Optional[_Union[Scalar, _Mapping]] = ..., collection: _Optional[_Union[LiteralCollection, _Mapping]] = ..., map: _Optional[_Union[LiteralMap, _Mapping]] = ..., hash: _Optional[str] = ...) -> None: ...

class LiteralCollection(_message.Message):
    __slots__ = ["literals"]
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.RepeatedCompositeFieldContainer[Literal]
    def __init__(self, literals: _Optional[_Iterable[_Union[Literal, _Mapping]]] = ...) -> None: ...

class LiteralMap(_message.Message):
    __slots__ = ["literals"]
    class LiteralsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: Literal
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[Literal, _Mapping]] = ...) -> None: ...
    LITERALS_FIELD_NUMBER: _ClassVar[int]
    literals: _containers.MessageMap[str, Literal]
    def __init__(self, literals: _Optional[_Mapping[str, Literal]] = ...) -> None: ...

class Primitive(_message.Message):
    __slots__ = ["boolean", "datetime", "duration", "float_value", "integer", "string_value"]
    BOOLEAN_FIELD_NUMBER: _ClassVar[int]
    DATETIME_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    FLOAT_VALUE_FIELD_NUMBER: _ClassVar[int]
    INTEGER_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    boolean: bool
    datetime: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    float_value: float
    integer: int
    string_value: str
    def __init__(self, integer: _Optional[int] = ..., float_value: _Optional[float] = ..., string_value: _Optional[str] = ..., boolean: bool = ..., datetime: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class RetryStrategy(_message.Message):
    __slots__ = ["retries"]
    RETRIES_FIELD_NUMBER: _ClassVar[int]
    retries: int
    def __init__(self, retries: _Optional[int] = ...) -> None: ...

class Scalar(_message.Message):
    __slots__ = ["binary", "blob", "error", "generic", "none_type", "primitive", "schema", "structured_dataset", "union"]
    BINARY_FIELD_NUMBER: _ClassVar[int]
    BLOB_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    GENERIC_FIELD_NUMBER: _ClassVar[int]
    NONE_TYPE_FIELD_NUMBER: _ClassVar[int]
    PRIMITIVE_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_DATASET_FIELD_NUMBER: _ClassVar[int]
    UNION_FIELD_NUMBER: _ClassVar[int]
    binary: Binary
    blob: Blob
    error: _types_pb2.Error
    generic: _struct_pb2.Struct
    none_type: Void
    primitive: Primitive
    schema: Schema
    structured_dataset: StructuredDataset
    union: Union
    def __init__(self, primitive: _Optional[_Union[Primitive, _Mapping]] = ..., blob: _Optional[_Union[Blob, _Mapping]] = ..., binary: _Optional[_Union[Binary, _Mapping]] = ..., schema: _Optional[_Union[Schema, _Mapping]] = ..., none_type: _Optional[_Union[Void, _Mapping]] = ..., error: _Optional[_Union[_types_pb2.Error, _Mapping]] = ..., generic: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., structured_dataset: _Optional[_Union[StructuredDataset, _Mapping]] = ..., union: _Optional[_Union[Union, _Mapping]] = ...) -> None: ...

class Schema(_message.Message):
    __slots__ = ["type", "uri"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    type: _types_pb2.SchemaType
    uri: str
    def __init__(self, uri: _Optional[str] = ..., type: _Optional[_Union[_types_pb2.SchemaType, _Mapping]] = ...) -> None: ...

class StructuredDataset(_message.Message):
    __slots__ = ["metadata", "uri"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    metadata: StructuredDatasetMetadata
    uri: str
    def __init__(self, uri: _Optional[str] = ..., metadata: _Optional[_Union[StructuredDatasetMetadata, _Mapping]] = ...) -> None: ...

class StructuredDatasetMetadata(_message.Message):
    __slots__ = ["structured_dataset_type"]
    STRUCTURED_DATASET_TYPE_FIELD_NUMBER: _ClassVar[int]
    structured_dataset_type: _types_pb2.StructuredDatasetType
    def __init__(self, structured_dataset_type: _Optional[_Union[_types_pb2.StructuredDatasetType, _Mapping]] = ...) -> None: ...

class Union(_message.Message):
    __slots__ = ["type", "value"]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    type: _types_pb2.LiteralType
    value: Literal
    def __init__(self, value: _Optional[_Union[Literal, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ...) -> None: ...

class UnionInfo(_message.Message):
    __slots__ = ["targetType"]
    TARGETTYPE_FIELD_NUMBER: _ClassVar[int]
    targetType: _types_pb2.LiteralType
    def __init__(self, targetType: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ...) -> None: ...

class Void(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
