from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.app import app_definition_pb2 as _app_definition_pb2
from flyteidl2.app import replica_definition_pb2 as _replica_definition_pb2
from flyteidl2.logs.dataplane import payload_pb2 as _payload_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TailLogsRequest(_message.Message):
    __slots__ = ["app_id", "replica_id"]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    REPLICA_ID_FIELD_NUMBER: _ClassVar[int]
    app_id: _app_definition_pb2.Identifier
    replica_id: _replica_definition_pb2.ReplicaIdentifier
    def __init__(self, app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ..., replica_id: _Optional[_Union[_replica_definition_pb2.ReplicaIdentifier, _Mapping]] = ...) -> None: ...

class ReplicaIdentifierList(_message.Message):
    __slots__ = ["replicas"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    replicas: _containers.RepeatedCompositeFieldContainer[_replica_definition_pb2.ReplicaIdentifier]
    def __init__(self, replicas: _Optional[_Iterable[_Union[_replica_definition_pb2.ReplicaIdentifier, _Mapping]]] = ...) -> None: ...

class LogLines(_message.Message):
    __slots__ = ["lines", "replica_id", "structured_lines"]
    LINES_FIELD_NUMBER: _ClassVar[int]
    REPLICA_ID_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_LINES_FIELD_NUMBER: _ClassVar[int]
    lines: _containers.RepeatedScalarFieldContainer[str]
    replica_id: _replica_definition_pb2.ReplicaIdentifier
    structured_lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
    def __init__(self, lines: _Optional[_Iterable[str]] = ..., replica_id: _Optional[_Union[_replica_definition_pb2.ReplicaIdentifier, _Mapping]] = ..., structured_lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ...) -> None: ...

class LogLinesBatch(_message.Message):
    __slots__ = ["logs"]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[LogLines]
    def __init__(self, logs: _Optional[_Iterable[_Union[LogLines, _Mapping]]] = ...) -> None: ...

class TailLogsResponse(_message.Message):
    __slots__ = ["replicas", "log_lines", "batches"]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    LOG_LINES_FIELD_NUMBER: _ClassVar[int]
    BATCHES_FIELD_NUMBER: _ClassVar[int]
    replicas: ReplicaIdentifierList
    log_lines: _payload_pb2.LogLines
    batches: LogLinesBatch
    def __init__(self, replicas: _Optional[_Union[ReplicaIdentifierList, _Mapping]] = ..., log_lines: _Optional[_Union[_payload_pb2.LogLines, _Mapping]] = ..., batches: _Optional[_Union[LogLinesBatch, _Mapping]] = ...) -> None: ...
