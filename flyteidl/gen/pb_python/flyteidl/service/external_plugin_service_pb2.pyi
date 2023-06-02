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

class TaskCreateRequest(_message.Message):
    __slots__ = ["inputs", "template", "output_prefix"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    inputs: _literals_pb2.LiteralMap
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    def __init__(self, inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ...) -> None: ...

class TaskCreateResponse(_message.Message):
    __slots__ = ["job_id"]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    def __init__(self, job_id: _Optional[str] = ...) -> None: ...

class TaskGetRequest(_message.Message):
    __slots__ = ["task_type", "job_id"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    job_id: str
    def __init__(self, task_type: _Optional[str] = ..., job_id: _Optional[str] = ...) -> None: ...

class TaskGetResponse(_message.Message):
    __slots__ = ["state", "outputs"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    state: State
    outputs: _literals_pb2.LiteralMap
    def __init__(self, state: _Optional[_Union[State, str]] = ..., outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ...) -> None: ...

class TaskDeleteRequest(_message.Message):
    __slots__ = ["task_type", "job_id"]
    TASK_TYPE_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    task_type: str
    job_id: str
    def __init__(self, task_type: _Optional[str] = ..., job_id: _Optional[str] = ...) -> None: ...

class TaskDeleteResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
