from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class OrgLimitExceeded(_message.Message):
    __slots__ = ["limit", "ceiling", "current", "requested", "tier_default", "upgrade_url", "tier"]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    CEILING_FIELD_NUMBER: _ClassVar[int]
    CURRENT_FIELD_NUMBER: _ClassVar[int]
    REQUESTED_FIELD_NUMBER: _ClassVar[int]
    TIER_DEFAULT_FIELD_NUMBER: _ClassVar[int]
    UPGRADE_URL_FIELD_NUMBER: _ClassVar[int]
    TIER_FIELD_NUMBER: _ClassVar[int]
    limit: str
    ceiling: int
    current: int
    requested: int
    tier_default: int
    upgrade_url: str
    tier: str
    def __init__(self, limit: _Optional[str] = ..., ceiling: _Optional[int] = ..., current: _Optional[int] = ..., requested: _Optional[int] = ..., tier_default: _Optional[int] = ..., upgrade_url: _Optional[str] = ..., tier: _Optional[str] = ...) -> None: ...

class OrgEntitlementRequired(_message.Message):
    __slots__ = ["entitlement", "values", "upgrade_url"]
    ENTITLEMENT_FIELD_NUMBER: _ClassVar[int]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    UPGRADE_URL_FIELD_NUMBER: _ClassVar[int]
    entitlement: str
    values: _containers.RepeatedScalarFieldContainer[str]
    upgrade_url: str
    def __init__(self, entitlement: _Optional[str] = ..., values: _Optional[_Iterable[str]] = ..., upgrade_url: _Optional[str] = ...) -> None: ...
