from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.app import app_definition_pb2 as _app_definition_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRequest(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class CreateResponse(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class GetRequest(_message.Message):
    __slots__ = ["app_id", "ingress"]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    app_id: _app_definition_pb2.Identifier
    ingress: _app_definition_pb2.Ingress
    def __init__(self, app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ..., ingress: _Optional[_Union[_app_definition_pb2.Ingress, _Mapping]] = ...) -> None: ...

class GetResponse(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class UpdateRequest(_message.Message):
    __slots__ = ["app", "reason"]
    APP_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    reason: str
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class UpdateResponse(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class DeleteRequest(_message.Message):
    __slots__ = ["app_id"]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    app_id: _app_definition_pb2.Identifier
    def __init__(self, app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ...) -> None: ...

class DeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["request", "org", "cluster_id", "project"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    cluster_id: _identifier_pb2.ClusterIdentifier
    project: _identifier_pb2.ProjectIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., cluster_id: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ..., project: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["apps", "token"]
    APPS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    apps: _containers.RepeatedCompositeFieldContainer[_app_definition_pb2.App]
    token: str
    def __init__(self, apps: _Optional[_Iterable[_Union[_app_definition_pb2.App, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WatchRequest(_message.Message):
    __slots__ = ["org", "cluster_id", "project", "app_id"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    org: str
    cluster_id: _identifier_pb2.ClusterIdentifier
    project: _identifier_pb2.ProjectIdentifier
    app_id: _app_definition_pb2.Identifier
    def __init__(self, org: _Optional[str] = ..., cluster_id: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ..., project: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ...) -> None: ...

class CreateEvent(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class UpdateEvent(_message.Message):
    __slots__ = ["updated_app", "old_app"]
    UPDATED_APP_FIELD_NUMBER: _ClassVar[int]
    OLD_APP_FIELD_NUMBER: _ClassVar[int]
    updated_app: _app_definition_pb2.App
    old_app: _app_definition_pb2.App
    def __init__(self, updated_app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ..., old_app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class DeleteEvent(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class WatchResponse(_message.Message):
    __slots__ = ["create_event", "update_event", "delete_event"]
    CREATE_EVENT_FIELD_NUMBER: _ClassVar[int]
    UPDATE_EVENT_FIELD_NUMBER: _ClassVar[int]
    DELETE_EVENT_FIELD_NUMBER: _ClassVar[int]
    create_event: CreateEvent
    update_event: UpdateEvent
    delete_event: DeleteEvent
    def __init__(self, create_event: _Optional[_Union[CreateEvent, _Mapping]] = ..., update_event: _Optional[_Union[UpdateEvent, _Mapping]] = ..., delete_event: _Optional[_Union[DeleteEvent, _Mapping]] = ...) -> None: ...

class UpdateStatusRequest(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class UpdateStatusResponse(_message.Message):
    __slots__ = ["app"]
    APP_FIELD_NUMBER: _ClassVar[int]
    app: _app_definition_pb2.App
    def __init__(self, app: _Optional[_Union[_app_definition_pb2.App, _Mapping]] = ...) -> None: ...

class LeaseRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ClusterIdentifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ...) -> None: ...

class LeaseResponse(_message.Message):
    __slots__ = ["apps"]
    APPS_FIELD_NUMBER: _ClassVar[int]
    apps: _containers.RepeatedCompositeFieldContainer[_app_definition_pb2.App]
    def __init__(self, apps: _Optional[_Iterable[_Union[_app_definition_pb2.App, _Mapping]]] = ...) -> None: ...
