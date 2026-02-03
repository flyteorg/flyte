from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.org import domain_definition_pb2 as _domain_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ListDomainsRequest(_message.Message):
    __slots__ = ["org", "include_settings"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_SETTINGS_FIELD_NUMBER: _ClassVar[int]
    org: str
    include_settings: bool
    def __init__(self, org: _Optional[str] = ..., include_settings: bool = ...) -> None: ...

class ListDomainsResponse(_message.Message):
    __slots__ = ["domains"]
    DOMAINS_FIELD_NUMBER: _ClassVar[int]
    domains: _containers.RepeatedCompositeFieldContainer[_domain_definition_pb2.Domain]
    def __init__(self, domains: _Optional[_Iterable[_Union[_domain_definition_pb2.Domain, _Mapping]]] = ...) -> None: ...

class CreateDomainRequest(_message.Message):
    __slots__ = ["org", "id", "friendly_name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    FRIENDLY_NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    id: str
    friendly_name: str
    def __init__(self, org: _Optional[str] = ..., id: _Optional[str] = ..., friendly_name: _Optional[str] = ...) -> None: ...

class CreateDomainResponse(_message.Message):
    __slots__ = ["domain"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    domain: _domain_definition_pb2.Domain
    def __init__(self, domain: _Optional[_Union[_domain_definition_pb2.Domain, _Mapping]] = ...) -> None: ...

class GetDomainRequest(_message.Message):
    __slots__ = ["org", "id", "include_settings"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_SETTINGS_FIELD_NUMBER: _ClassVar[int]
    org: str
    id: str
    include_settings: bool
    def __init__(self, org: _Optional[str] = ..., id: _Optional[str] = ..., include_settings: bool = ...) -> None: ...

class GetDomainResponse(_message.Message):
    __slots__ = ["domain"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    domain: _domain_definition_pb2.Domain
    def __init__(self, domain: _Optional[_Union[_domain_definition_pb2.Domain, _Mapping]] = ...) -> None: ...

class UpdateDomainRequest(_message.Message):
    __slots__ = ["org", "id", "friendly_name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    FRIENDLY_NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    id: str
    friendly_name: str
    def __init__(self, org: _Optional[str] = ..., id: _Optional[str] = ..., friendly_name: _Optional[str] = ...) -> None: ...

class UpdateDomainResponse(_message.Message):
    __slots__ = ["domain"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    domain: _domain_definition_pb2.Domain
    def __init__(self, domain: _Optional[_Union[_domain_definition_pb2.Domain, _Mapping]] = ...) -> None: ...

class DeleteDomainRequest(_message.Message):
    __slots__ = ["org", "id"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    org: str
    id: str
    def __init__(self, org: _Optional[str] = ..., id: _Optional[str] = ...) -> None: ...

class DeleteDomainResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
