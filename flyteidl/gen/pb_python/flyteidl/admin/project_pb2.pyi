from flyteidl.admin import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetDomainRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Domain(_message.Message):
    __slots__ = ["id", "name"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class GetDomainsResponse(_message.Message):
    __slots__ = ["domains"]
    DOMAINS_FIELD_NUMBER: _ClassVar[int]
    domains: _containers.RepeatedCompositeFieldContainer[Domain]
    def __init__(self, domains: _Optional[_Iterable[_Union[Domain, _Mapping]]] = ...) -> None: ...

class Project(_message.Message):
    __slots__ = ["id", "name", "domains", "description", "labels", "state", "org"]
    class ProjectState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        ACTIVE: _ClassVar[Project.ProjectState]
        ARCHIVED: _ClassVar[Project.ProjectState]
        SYSTEM_GENERATED: _ClassVar[Project.ProjectState]
    ACTIVE: Project.ProjectState
    ARCHIVED: Project.ProjectState
    SYSTEM_GENERATED: Project.ProjectState
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DOMAINS_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    domains: _containers.RepeatedCompositeFieldContainer[Domain]
    description: str
    labels: _common_pb2.Labels
    state: Project.ProjectState
    org: str
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., domains: _Optional[_Iterable[_Union[Domain, _Mapping]]] = ..., description: _Optional[str] = ..., labels: _Optional[_Union[_common_pb2.Labels, _Mapping]] = ..., state: _Optional[_Union[Project.ProjectState, str]] = ..., org: _Optional[str] = ...) -> None: ...

class Projects(_message.Message):
    __slots__ = ["projects", "token"]
    PROJECTS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    projects: _containers.RepeatedCompositeFieldContainer[Project]
    token: str
    def __init__(self, projects: _Optional[_Iterable[_Union[Project, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class ProjectListRequest(_message.Message):
    __slots__ = ["limit", "token", "filters", "sort_by", "org"]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    limit: int
    token: str
    filters: str
    sort_by: _common_pb2.Sort
    org: str
    def __init__(self, limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[_common_pb2.Sort, _Mapping]] = ..., org: _Optional[str] = ...) -> None: ...

class ProjectRegisterRequest(_message.Message):
    __slots__ = ["project"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    project: Project
    def __init__(self, project: _Optional[_Union[Project, _Mapping]] = ...) -> None: ...

class ProjectRegisterResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ProjectUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class ProjectGetRequest(_message.Message):
    __slots__ = ["id", "org"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    id: str
    org: str
    def __init__(self, id: _Optional[str] = ..., org: _Optional[str] = ...) -> None: ...
