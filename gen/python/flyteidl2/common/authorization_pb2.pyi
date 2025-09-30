from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Action(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ACTION_NONE: _ClassVar[Action]
    ACTION_CREATE: _ClassVar[Action]
    ACTION_READ: _ClassVar[Action]
    ACTION_UPDATE: _ClassVar[Action]
    ACTION_DELETE: _ClassVar[Action]
    ACTION_VIEW_FLYTE_INVENTORY: _ClassVar[Action]
    ACTION_VIEW_FLYTE_EXECUTIONS: _ClassVar[Action]
    ACTION_REGISTER_FLYTE_INVENTORY: _ClassVar[Action]
    ACTION_CREATE_FLYTE_EXECUTIONS: _ClassVar[Action]
    ACTION_ADMINISTER_PROJECT: _ClassVar[Action]
    ACTION_MANAGE_PERMISSIONS: _ClassVar[Action]
    ACTION_ADMINISTER_ACCOUNT: _ClassVar[Action]
    ACTION_MANAGE_CLUSTER: _ClassVar[Action]
    ACTION_EDIT_EXECUTION_RELATED_ATTRIBUTES: _ClassVar[Action]
    ACTION_EDIT_CLUSTER_RELATED_ATTRIBUTES: _ClassVar[Action]
    ACTION_EDIT_UNUSED_ATTRIBUTES: _ClassVar[Action]
    ACTION_SUPPORT_SYSTEM_LOGS: _ClassVar[Action]
ACTION_NONE: Action
ACTION_CREATE: Action
ACTION_READ: Action
ACTION_UPDATE: Action
ACTION_DELETE: Action
ACTION_VIEW_FLYTE_INVENTORY: Action
ACTION_VIEW_FLYTE_EXECUTIONS: Action
ACTION_REGISTER_FLYTE_INVENTORY: Action
ACTION_CREATE_FLYTE_EXECUTIONS: Action
ACTION_ADMINISTER_PROJECT: Action
ACTION_MANAGE_PERMISSIONS: Action
ACTION_ADMINISTER_ACCOUNT: Action
ACTION_MANAGE_CLUSTER: Action
ACTION_EDIT_EXECUTION_RELATED_ATTRIBUTES: Action
ACTION_EDIT_CLUSTER_RELATED_ATTRIBUTES: Action
ACTION_EDIT_UNUSED_ATTRIBUTES: Action
ACTION_SUPPORT_SYSTEM_LOGS: Action

class Organization(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class Domain(_message.Message):
    __slots__ = ["name", "organization"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    name: str
    organization: Organization
    def __init__(self, name: _Optional[str] = ..., organization: _Optional[_Union[Organization, _Mapping]] = ...) -> None: ...

class Project(_message.Message):
    __slots__ = ["name", "domain"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    name: str
    domain: Domain
    def __init__(self, name: _Optional[str] = ..., domain: _Optional[_Union[Domain, _Mapping]] = ...) -> None: ...

class Workflow(_message.Message):
    __slots__ = ["name", "project"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    name: str
    project: Project
    def __init__(self, name: _Optional[str] = ..., project: _Optional[_Union[Project, _Mapping]] = ...) -> None: ...

class LaunchPlan(_message.Message):
    __slots__ = ["name", "project"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    name: str
    project: Project
    def __init__(self, name: _Optional[str] = ..., project: _Optional[_Union[Project, _Mapping]] = ...) -> None: ...

class Resource(_message.Message):
    __slots__ = ["organization", "domain", "project", "workflow", "launch_plan", "cluster"]
    ORGANIZATION_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    LAUNCH_PLAN_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    organization: Organization
    domain: Domain
    project: Project
    workflow: Workflow
    launch_plan: LaunchPlan
    cluster: _identifier_pb2.ClusterIdentifier
    def __init__(self, organization: _Optional[_Union[Organization, _Mapping]] = ..., domain: _Optional[_Union[Domain, _Mapping]] = ..., project: _Optional[_Union[Project, _Mapping]] = ..., workflow: _Optional[_Union[Workflow, _Mapping]] = ..., launch_plan: _Optional[_Union[LaunchPlan, _Mapping]] = ..., cluster: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ...) -> None: ...

class Permission(_message.Message):
    __slots__ = ["resource", "actions"]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    resource: Resource
    actions: _containers.RepeatedScalarFieldContainer[Action]
    def __init__(self, resource: _Optional[_Union[Resource, _Mapping]] = ..., actions: _Optional[_Iterable[_Union[Action, str]]] = ...) -> None: ...
