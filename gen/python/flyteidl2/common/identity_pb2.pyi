from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import policy_pb2 as _policy_pb2
from flyteidl2.common import role_pb2 as _role_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class User(_message.Message):
    __slots__ = ["id", "spec", "roles", "policies"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    ROLES_FIELD_NUMBER: _ClassVar[int]
    POLICIES_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.UserIdentifier
    spec: UserSpec
    roles: _containers.RepeatedCompositeFieldContainer[_role_pb2.Role]
    policies: _containers.RepeatedCompositeFieldContainer[_policy_pb2.Policy]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.UserIdentifier, _Mapping]] = ..., spec: _Optional[_Union[UserSpec, _Mapping]] = ..., roles: _Optional[_Iterable[_Union[_role_pb2.Role, _Mapping]]] = ..., policies: _Optional[_Iterable[_Union[_policy_pb2.Policy, _Mapping]]] = ...) -> None: ...

class UserSpec(_message.Message):
    __slots__ = ["first_name", "last_name", "email", "organization", "user_handle", "groups", "photo_url"]
    FIRST_NAME_FIELD_NUMBER: _ClassVar[int]
    LAST_NAME_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    USER_HANDLE_FIELD_NUMBER: _ClassVar[int]
    GROUPS_FIELD_NUMBER: _ClassVar[int]
    PHOTO_URL_FIELD_NUMBER: _ClassVar[int]
    first_name: str
    last_name: str
    email: str
    organization: str
    user_handle: str
    groups: _containers.RepeatedScalarFieldContainer[str]
    photo_url: str
    def __init__(self, first_name: _Optional[str] = ..., last_name: _Optional[str] = ..., email: _Optional[str] = ..., organization: _Optional[str] = ..., user_handle: _Optional[str] = ..., groups: _Optional[_Iterable[str]] = ..., photo_url: _Optional[str] = ...) -> None: ...

class Application(_message.Message):
    __slots__ = ["id", "spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ApplicationIdentifier
    spec: AppSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ApplicationIdentifier, _Mapping]] = ..., spec: _Optional[_Union[AppSpec, _Mapping]] = ...) -> None: ...

class AppSpec(_message.Message):
    __slots__ = ["name", "organization"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    name: str
    organization: str
    def __init__(self, name: _Optional[str] = ..., organization: _Optional[str] = ...) -> None: ...

class EnrichedIdentity(_message.Message):
    __slots__ = ["user", "application"]
    USER_FIELD_NUMBER: _ClassVar[int]
    APPLICATION_FIELD_NUMBER: _ClassVar[int]
    user: User
    application: Application
    def __init__(self, user: _Optional[_Union[User, _Mapping]] = ..., application: _Optional[_Union[Application, _Mapping]] = ...) -> None: ...

class Identity(_message.Message):
    __slots__ = ["user_id", "application_id"]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    APPLICATION_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: _identifier_pb2.UserIdentifier
    application_id: _identifier_pb2.ApplicationIdentifier
    def __init__(self, user_id: _Optional[_Union[_identifier_pb2.UserIdentifier, _Mapping]] = ..., application_id: _Optional[_Union[_identifier_pb2.ApplicationIdentifier, _Mapping]] = ...) -> None: ...
