from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class State(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    RETRYABLE_FAILURE: _ClassVar[State]
    PERMANENT_FAILURE: _ClassVar[State]
    PENDING: _ClassVar[State]
    RUNNING: _ClassVar[State]
    SUCCEEDED: _ClassVar[State]
RETRYABLE_FAILURE: State
PERMANENT_FAILURE: State
PENDING: State
RUNNING: State
SUCCEEDED: State

class CreateTaskRequest(_message.Message):
    __slots__ = ["inputs", "template", "output_prefix"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    inputs: _literals_pb2.LiteralMap
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    def __init__(self, inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ...) -> None: ...

class CreateTaskResponse(_message.Message):
    __slots__ = ["resource_meta"]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    resource_meta: bytes
    def __init__(self, resource_meta: _Optional[bytes] = ...) -> None: ...

class GetTaskRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ...) -> None: ...

class GetTaskResponse(_message.Message):
    __slots__ = ["resource"]
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    resource: Resource
    def __init__(self, resource: _Optional[_Union[Resource, _Mapping]] = ...) -> None: ...

class Resource(_message.Message):
    __slots__ = ["state", "outputs"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    state: State
    outputs: _literals_pb2.LiteralMap
    def __init__(self, state: _Optional[_Union[State, str]] = ..., outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class DeleteTaskRequest(_message.Message):
    __slots__ = ["task_type", "resource_meta"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_META_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    resource_meta: bytes
    def __init__(self, task_type: _Optional[str] = ..., resource_meta: _Optional[bytes] = ...) -> None: ...

class DeleteTaskResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
