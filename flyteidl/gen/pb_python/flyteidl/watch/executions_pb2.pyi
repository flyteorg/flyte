from google.protobuf import timestamp_pb2 as _timestamp_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WatchExecutionStatusUpdatesRequest(_message.Message):
    __slots__ = ["cluster"]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    cluster: str
    def __init__(self, cluster: _Optional[str] = ...) -> None: ...

class WatchExecutionStatusUpdatesResponse(_message.Message):
    __slots__ = ["id", "phase", "output_uri", "error"]
    ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_URI_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.WorkflowExecutionIdentifier
    phase: _execution_pb2.WorkflowExecution.Phase
    output_uri: str
    error: _execution_pb2.ExecutionError
    def __init__(self, id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., phase: _Optional[_Union[_execution_pb2.WorkflowExecution.Phase, str]] = ..., output_uri: _Optional[str] = ..., error: _Optional[_Union[_execution_pb2.ExecutionError, _Mapping]] = ...) -> None: ...
