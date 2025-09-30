from flyteidl2.core import literals_pb2 as _literals_pb2
from google.protobuf import wrappers_pb2 as _wrappers_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Labels(_message.Message):
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

class Annotations(_message.Message):
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

class Envs(_message.Message):
    __slots__ = ["values"]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedCompositeFieldContainer[_literals_pb2.KeyValuePair]
    def __init__(self, values: _Optional[_Iterable[_Union[_literals_pb2.KeyValuePair, _Mapping]]] = ...) -> None: ...

class RunSpec(_message.Message):
    __slots__ = ["labels", "annotations", "envs", "interruptible", "overwrite_cache", "cluster"]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    ENVS_FIELD_NUMBER: _ClassVar[int]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    OVERWRITE_CACHE_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    labels: Labels
    annotations: Annotations
    envs: Envs
    interruptible: _wrappers_pb2.BoolValue
    overwrite_cache: bool
    cluster: str
    def __init__(self, labels: _Optional[_Union[Labels, _Mapping]] = ..., annotations: _Optional[_Union[Annotations, _Mapping]] = ..., envs: _Optional[_Union[Envs, _Mapping]] = ..., interruptible: _Optional[_Union[_wrappers_pb2.BoolValue, _Mapping]] = ..., overwrite_cache: bool = ..., cluster: _Optional[str] = ...) -> None: ...
