from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ResourceType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    UNSPECIFIED: _ClassVar[ResourceType]
    TASK: _ClassVar[ResourceType]
    WORKFLOW: _ClassVar[ResourceType]
    LAUNCH_PLAN: _ClassVar[ResourceType]
    DATASET: _ClassVar[ResourceType]
UNSPECIFIED: ResourceType
TASK: ResourceType
WORKFLOW: ResourceType
LAUNCH_PLAN: ResourceType
DATASET: ResourceType

class Identifier(_message.Message):
    __slots__ = ["resource_type", "project", "domain", "name", "version"]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    resource_type: ResourceType
    project: str
    domain: str
    name: str
    version: str
    def __init__(self, resource_type: _Optional[_Union[ResourceType, str]] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...

class WorkflowExecutionIdentifier(_message.Message):
    __slots__ = ["project", "domain", "name"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    name: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class NodeExecutionIdentifier(_message.Message):
    __slots__ = ["node_id", "execution_id"]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    node_id: str
    execution_id: WorkflowExecutionIdentifier
    def __init__(self, node_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...

class TaskExecutionIdentifier(_message.Message):
    __slots__ = ["task_id", "node_execution_id", "retry_attempt"]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    NODE_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    RETRY_ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    task_id: Identifier
    node_execution_id: NodeExecutionIdentifier
    retry_attempt: int
    def __init__(self, task_id: _Optional[_Union[Identifier, _Mapping]] = ..., node_execution_id: _Optional[_Union[NodeExecutionIdentifier, _Mapping]] = ..., retry_attempt: _Optional[int] = ...) -> None: ...

class SignalIdentifier(_message.Message):
    __slots__ = ["signal_id", "execution_id"]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    signal_id: str
    execution_id: WorkflowExecutionIdentifier
    def __init__(self, signal_id: _Optional[str] = ..., execution_id: _Optional[_Union[WorkflowExecutionIdentifier, _Mapping]] = ...) -> None: ...
