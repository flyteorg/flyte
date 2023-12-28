from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.core import workflow_pb2 as _workflow_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DynamicJobSpec(_message.Message):
    __slots__ = ["nodes", "min_successes", "outputs", "tasks", "subworkflows"]
    NODES_FIELD_NUMBER: _ClassVar[int]
    MIN_SUCCESSES_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    SUBWORKFLOWS_FIELD_NUMBER: _ClassVar[int]
    nodes: _containers.RepeatedCompositeFieldContainer[_workflow_pb2.Node]
    min_successes: int
    outputs: _containers.RepeatedCompositeFieldContainer[_literals_pb2.Binding]
    tasks: _containers.RepeatedCompositeFieldContainer[_tasks_pb2.TaskTemplate]
    subworkflows: _containers.RepeatedCompositeFieldContainer[_workflow_pb2.WorkflowTemplate]
    def __init__(self, nodes: _Optional[_Iterable[_Union[_workflow_pb2.Node, _Mapping]]] = ..., min_successes: _Optional[int] = ..., outputs: _Optional[_Iterable[_Union[_literals_pb2.Binding, _Mapping]]] = ..., tasks: _Optional[_Iterable[_Union[_tasks_pb2.TaskTemplate, _Mapping]]] = ..., subworkflows: _Optional[_Iterable[_Union[_workflow_pb2.WorkflowTemplate, _Mapping]]] = ...) -> None: ...
