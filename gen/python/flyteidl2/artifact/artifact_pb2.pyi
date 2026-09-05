from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import identity_pb2 as _identity_pb2
from flyteidl2.core import artifact_id_pb2 as _artifact_id_pb2
from flyteidl2.core import literals_pb2 as _literals_pb2
from flyteidl2.core import types_pb2 as _types_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ArtifactName(_message.Message):
    __slots__ = ["org", "project", "domain", "name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ArtifactIdentifier(_message.Message):
    __slots__ = ["name", "version"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: ArtifactName
    version: str
    def __init__(self, name: _Optional[_Union[ArtifactName, _Mapping]] = ..., version: _Optional[str] = ...) -> None: ...

class TaskActionSource(_message.Message):
    __slots__ = ["action", "attempt"]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    action: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, action: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class ArtifactSource(_message.Message):
    __slots__ = ["task_action", "external_ref"]
    TASK_ACTION_FIELD_NUMBER: _ClassVar[int]
    EXTERNAL_REF_FIELD_NUMBER: _ClassVar[int]
    task_action: TaskActionSource
    external_ref: str
    def __init__(self, task_action: _Optional[_Union[TaskActionSource, _Mapping]] = ..., external_ref: _Optional[str] = ...) -> None: ...

class ArtifactSpec(_message.Message):
    __slots__ = ["value", "type", "info", "source", "parent_artifacts"]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    INFO_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    PARENT_ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    value: _literals_pb2.Literal
    type: _types_pb2.LiteralType
    info: _artifact_id_pb2.ArtifactInfo
    source: ArtifactSource
    parent_artifacts: _containers.RepeatedCompositeFieldContainer[_artifact_id_pb2.ArtifactVersionId]
    def __init__(self, value: _Optional[_Union[_literals_pb2.Literal, _Mapping]] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., info: _Optional[_Union[_artifact_id_pb2.ArtifactInfo, _Mapping]] = ..., source: _Optional[_Union[ArtifactSource, _Mapping]] = ..., parent_artifacts: _Optional[_Iterable[_Union[_artifact_id_pb2.ArtifactVersionId, _Mapping]]] = ...) -> None: ...

class Artifact(_message.Message):
    __slots__ = ["artifact_id", "spec", "created_at", "created_by"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_FIELD_NUMBER: _ClassVar[int]
    artifact_id: ArtifactIdentifier
    spec: ArtifactSpec
    created_at: _timestamp_pb2.Timestamp
    created_by: _identity_pb2.EnrichedIdentity
    def __init__(self, artifact_id: _Optional[_Union[ArtifactIdentifier, _Mapping]] = ..., spec: _Optional[_Union[ArtifactSpec, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., created_by: _Optional[_Union[_identity_pb2.EnrichedIdentity, _Mapping]] = ...) -> None: ...
