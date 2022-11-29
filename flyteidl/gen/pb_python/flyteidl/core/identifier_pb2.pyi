from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DATASET: ResourceType
DESCRIPTOR: _descriptor.FileDescriptor
LAUNCH_PLAN: ResourceType
TASK: ResourceType
UNSPECIFIED: ResourceType
WORKFLOW: ResourceType

class Identifier(_message.Message):
    __slots__ = ["domain", "name", "project", "resource_type", "version"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    domain: str
    name: str
    project: str
    resource_type: ResourceType
    version: str
    def __init__(self, resource_type: _Optional[_Union[ResourceType, str]] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class NodeExecutionIdentifier(_message.Message):
    __slots__ = ["execution_id", "node_id"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    execution_id: WorkflowExecutionIdentifier
    node_id: str
    def __init__(self, node_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class SignalIdentifier(_message.Message):
    __slots__ = ["execution_id", "signal_id"]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    execution_id: WorkflowExecutionIdentifier
    signal_id: str
    def __init__(self, signal_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionIdentifier(_message.Message):
    __slots__ = ["node_execution_id", "retry_attempt", "task_id"]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    node_execution_id: NodeExecutionIdentifier
    retry_attempt: int
    task_id: Identifier
    def __init__(self, task_id: _Optional[_Union[Identifier, _Mapping]] = ..., node_execution_id: _Optional[_Union[NodeExecutionIdentifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ...) -> None: ...

class WorkflowExecutionIdentifier(_message.Message):
    __slots__ = ["domain", "name", "project"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    domain: str
    name: str
    project: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ResourceType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
