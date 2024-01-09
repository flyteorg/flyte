from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Span(_message.Message):
    __slots__ = ["start_time", "end_time", "workflow_id", "node_id", "task_id", "operation_id", "spans"]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_ID_FIELD_NUMBER: _ClassVar[int]
    NODE_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    OPERATION_ID_FIELD_NUMBER: _ClassVar[int]
    SPANS_FIELD_NUMBER: _ClassVar[int]
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    workflow_id: _identifier_pb2.WorkflowExecutionIdentifier
    node_id: _identifier_pb2.NodeExecutionIdentifier
    task_id: _identifier_pb2.TaskExecutionIdentifier
    operation_id: str
    spans: _containers.RepeatedCompositeFieldContainer[Span]
    def __init__(self, start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., workflow_id: _Optional[_Union[_identifier_pb2.WorkflowExecutionIdentifier, _Mapping]] = ..., node_id: _Optional[_Union[_identifier_pb2.NodeExecutionIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., operation_id: _Optional[str] = ..., spans: _Optional[_Iterable[_Union[Span, _Mapping]]] = ...) -> None: ...

class ExecutionMetricResult(_message.Message):
    __slots__ = ["metric", "data"]
    METRIC_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    metric: str
    data: _struct_pb2.Struct
    def __init__(self, metric: _Optional[str] = ..., data: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
