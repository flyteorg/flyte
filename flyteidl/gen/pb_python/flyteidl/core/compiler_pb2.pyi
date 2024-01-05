from flyteidl.core import workflow_pb2 as _workflow_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ConnectionSet(_message.Message):
    __slots__ = ["downstream", "upstream"]
    class IdList(_message.Message):
        __slots__ = ["ids"]
        IDS_FIELD_NUMBER: _ClassVar[int]
        ids: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, ids: _Optional[_Iterable[str]] = ...) -> None: ...
    class DownstreamEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: ConnectionSet.IdList
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[ConnectionSet.IdList, _Mapping]] = ...) -> None: ...
    class UpstreamEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: ConnectionSet.IdList
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[ConnectionSet.IdList, _Mapping]] = ...) -> None: ...
    DOWNSTREAM_FIELD_NUMBER: _ClassVar[int]
    UPSTREAM_FIELD_NUMBER: _ClassVar[int]
    downstream: _containers.MessageMap[str, ConnectionSet.IdList]
    upstream: _containers.MessageMap[str, ConnectionSet.IdList]
    def __init__(self, downstream: _Optional[_Mapping[str, ConnectionSet.IdList]] = ..., upstream: _Optional[_Mapping[str, ConnectionSet.IdList]] = ...) -> None: ...

class CompiledWorkflow(_message.Message):
    __slots__ = ["template", "connections"]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    template: _workflow_pb2.WorkflowTemplate
    connections: ConnectionSet
    def __init__(self, template: _Optional[_Union[_workflow_pb2.WorkflowTemplate, _Mapping]] = ..., connections: _Optional[_Union[ConnectionSet, _Mapping]] = ...) -> None: ...

class CompiledTask(_message.Message):
    __slots__ = ["template"]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    template: _tasks_pb2.TaskTemplate
    def __init__(self, template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ...) -> None: ...

class CompiledWorkflowClosure(_message.Message):
    __slots__ = ["primary", "sub_workflows", "tasks"]
    PRIMARY_FIELD_NUMBER: _ClassVar[int]
    SUB_WORKFLOWS_FIELD_NUMBER: _ClassVar[int]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    primary: CompiledWorkflow
    sub_workflows: _containers.RepeatedCompositeFieldContainer[CompiledWorkflow]
    tasks: _containers.RepeatedCompositeFieldContainer[CompiledTask]
    def __init__(self, primary: _Optional[_Union[CompiledWorkflow, _Mapping]] = ..., sub_workflows: _Optional[_Iterable[_Union[CompiledWorkflow, _Mapping]]] = ..., tasks: _Optional[_Iterable[_Union[CompiledTask, _Mapping]]] = ...) -> None: ...
