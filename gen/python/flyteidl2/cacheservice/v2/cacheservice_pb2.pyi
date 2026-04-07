from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.cacheservice import cacheservice_pb2 as _cacheservice_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Identifier(_message.Message):
    __slots__ = ["org", "project", "domain"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ...) -> None: ...

class GetCacheRequest(_message.Message):
    __slots__ = ["base_request", "identifier"]
    BASE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    base_request: _cacheservice_pb2.GetCacheRequest
    identifier: Identifier
    def __init__(self, base_request: _Optional[_Union[_cacheservice_pb2.GetCacheRequest, _Mapping]] = ..., identifier: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class PutCacheRequest(_message.Message):
    __slots__ = ["base_request", "identifier"]
    BASE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    base_request: _cacheservice_pb2.PutCacheRequest
    identifier: Identifier
    def __init__(self, base_request: _Optional[_Union[_cacheservice_pb2.PutCacheRequest, _Mapping]] = ..., identifier: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class DeleteCacheRequest(_message.Message):
    __slots__ = ["base_request", "identifier"]
    BASE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    base_request: _cacheservice_pb2.DeleteCacheRequest
    identifier: Identifier
    def __init__(self, base_request: _Optional[_Union[_cacheservice_pb2.DeleteCacheRequest, _Mapping]] = ..., identifier: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class GetOrExtendReservationRequest(_message.Message):
    __slots__ = ["base_request", "identifier"]
    BASE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    base_request: _cacheservice_pb2.GetOrExtendReservationRequest
    identifier: Identifier
    def __init__(self, base_request: _Optional[_Union[_cacheservice_pb2.GetOrExtendReservationRequest, _Mapping]] = ..., identifier: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...

class ReleaseReservationRequest(_message.Message):
    __slots__ = ["base_request", "identifier"]
    BASE_REQUEST_FIELD_NUMBER: _ClassVar[int]
    IDENTIFIER_FIELD_NUMBER: _ClassVar[int]
    base_request: _cacheservice_pb2.ReleaseReservationRequest
    identifier: Identifier
    def __init__(self, base_request: _Optional[_Union[_cacheservice_pb2.ReleaseReservationRequest, _Mapping]] = ..., identifier: _Optional[_Union[Identifier, _Mapping]] = ...) -> None: ...
