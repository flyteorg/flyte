from google.api import annotations_pb2 as _annotations_pb2
from flyteidl.core import errors_pb2 as _errors_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EvictCacheRequest(_message.Message):
    __slots__ = ["task_execution_id", "workflow_execution_id"]
    TASK_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    task_execution_id: _identifier_pb2.TaskExecutionIdentifier
    workflow_execution_id: _identifier_pb2.WorkflowExecutionIdentifier
    def __init__(self, workflow_execution_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ...) -> None: ...

class EvictCacheResponse(_message.Message):
    __slots__ = ["errors"]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    errors: _errors_pb2.CacheEvictionErrorList
    def __init__(self, errors: _Optional[_Union[_errors_pb2.CacheEvictionErrorList, _Mapping]] = ...) -> None: ...
