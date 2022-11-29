from flyteidl.core import workflow_pb2 as _workflow_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowClosure(_message.Message):
    __slots__ = ["tasks", "workflow"]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    tasks: _containers.RepeatedCompositeFieldContainer[_tasks_pb2.TaskTemplate]
    workflow: _workflow_pb2.WorkflowTemplate
    def __init__(self, workflow: _Optional[_Union[_workflow_pb2.WorkflowTemplate, _Mapping]] = ..., tasks: _Optional[_Iterable[_Union[_tasks_pb2.TaskTemplate, _Mapping]]] = ...) -> None: ...
