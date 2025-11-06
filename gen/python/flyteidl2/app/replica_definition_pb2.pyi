from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.app import app_definition_pb2 as _app_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ReplicaIdentifier(_message.Message):
    __slots__ = ["app_id", "name"]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    app_id: _app_definition_pb2.Identifier
    name: str
    def __init__(self, app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class ReplicaMeta(_message.Message):
    __slots__ = ["id", "revision"]
    ID_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    id: ReplicaIdentifier
    revision: int
    def __init__(self, id: _Optional[_Union[ReplicaIdentifier, _Mapping]] = ..., revision: _Optional[int] = ...) -> None: ...

class ReplicaList(_message.Message):
    __slots__ = ["items"]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[Replica]
    def __init__(self, items: _Optional[_Iterable[_Union[Replica, _Mapping]]] = ...) -> None: ...

class Replica(_message.Message):
    __slots__ = ["metadata", "status"]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    metadata: ReplicaMeta
    status: ReplicaStatus
    def __init__(self, metadata: _Optional[_Union[ReplicaMeta, _Mapping]] = ..., status: _Optional[_Union[ReplicaStatus, _Mapping]] = ...) -> None: ...

class ReplicaStatus(_message.Message):
    __slots__ = ["deployment_status", "reason"]
    DEPLOYMENT_STATUS_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    deployment_status: str
    reason: str
    def __init__(self, deployment_status: _Optional[str] = ..., reason: _Optional[str] = ...) -> None: ...
