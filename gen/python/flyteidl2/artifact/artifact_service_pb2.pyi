from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.artifact import artifact_pb2 as _artifact_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateArtifactRequest(_message.Message):
    __slots__ = ["artifact_id", "spec"]
    ARTIFACT_ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    artifact_id: _artifact_pb2.ArtifactIdentifier
    spec: _artifact_pb2.ArtifactSpec
    def __init__(self, artifact_id: _Optional[_Union[_artifact_pb2.ArtifactIdentifier, _Mapping]] = ..., spec: _Optional[_Union[_artifact_pb2.ArtifactSpec, _Mapping]] = ...) -> None: ...

class CreateArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: _artifact_pb2.Artifact
    def __init__(self, artifact: _Optional[_Union[_artifact_pb2.Artifact, _Mapping]] = ...) -> None: ...

class GetArtifactRequest(_message.Message):
    __slots__ = ["name", "version"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: _artifact_pb2.ArtifactName
    version: str
    def __init__(self, name: _Optional[_Union[_artifact_pb2.ArtifactName, _Mapping]] = ..., version: _Optional[str] = ...) -> None: ...

class GetArtifactResponse(_message.Message):
    __slots__ = ["artifact"]
    ARTIFACT_FIELD_NUMBER: _ClassVar[int]
    artifact: _artifact_pb2.Artifact
    def __init__(self, artifact: _Optional[_Union[_artifact_pb2.Artifact, _Mapping]] = ...) -> None: ...

class ListArtifactsRequest(_message.Message):
    __slots__ = ["request", "project_id", "name"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    project_id: _identifier_pb2.ProjectIdentifier
    name: str
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class ListArtifactsResponse(_message.Message):
    __slots__ = ["artifacts", "token"]
    ARTIFACTS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    artifacts: _containers.RepeatedCompositeFieldContainer[_artifact_pb2.Artifact]
    token: str
    def __init__(self, artifacts: _Optional[_Iterable[_Union[_artifact_pb2.Artifact, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class ListArtifactNamesRequest(_message.Message):
    __slots__ = ["request", "project_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    project_id: _identifier_pb2.ProjectIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ...) -> None: ...

class ArtifactGroup(_message.Message):
    __slots__ = ["latest", "versions"]
    LATEST_FIELD_NUMBER: _ClassVar[int]
    VERSIONS_FIELD_NUMBER: _ClassVar[int]
    latest: _artifact_pb2.Artifact
    versions: int
    def __init__(self, latest: _Optional[_Union[_artifact_pb2.Artifact, _Mapping]] = ..., versions: _Optional[int] = ...) -> None: ...

class ListArtifactNamesResponse(_message.Message):
    __slots__ = ["groups", "token"]
    GROUPS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    groups: _containers.RepeatedCompositeFieldContainer[ArtifactGroup]
    token: str
    def __init__(self, groups: _Optional[_Iterable[_Union[ArtifactGroup, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...
