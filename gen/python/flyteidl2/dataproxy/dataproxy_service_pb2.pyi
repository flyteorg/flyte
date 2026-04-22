from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.app import app_definition_pb2 as _app_definition_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import run_pb2 as _run_pb2
from flyteidl2.logs.dataplane import payload_pb2 as _payload_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ArtifactType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ARTIFACT_TYPE_UNSPECIFIED: _ClassVar[ArtifactType]
    ARTIFACT_TYPE_REPORT: _ClassVar[ArtifactType]
    ARTIFACT_TYPE_CODE_BUNDLE: _ClassVar[ArtifactType]
ARTIFACT_TYPE_UNSPECIFIED: ArtifactType
ARTIFACT_TYPE_REPORT: ArtifactType
ARTIFACT_TYPE_CODE_BUNDLE: ArtifactType

class CreateUploadLocationRequest(_message.Message):
    __slots__ = ["project", "domain", "filename", "expires_in", "content_md5", "filename_root", "add_content_md5_metadata", "org", "content_length"]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    CONTENT_MD5_FIELD_NUMBER: _ClassVar[int]
    FILENAME_ROOT_FIELD_NUMBER: _ClassVar[int]
    ADD_CONTENT_MD5_METADATA_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CONTENT_LENGTH_FIELD_NUMBER: _ClassVar[int]
    project: str
    domain: str
    filename: str
    expires_in: _duration_pb2.Duration
    content_md5: bytes
    filename_root: str
    add_content_md5_metadata: bool
    org: str
    content_length: int
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., filename: _Optional[str] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., content_md5: _Optional[bytes] = ..., filename_root: _Optional[str] = ..., add_content_md5_metadata: bool = ..., org: _Optional[str] = ..., content_length: _Optional[int] = ...) -> None: ...

class CreateUploadLocationResponse(_message.Message):
    __slots__ = ["signed_url", "native_url", "expires_at", "headers"]
    class HeadersEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    NATIVE_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    HEADERS_FIELD_NUMBER: _ClassVar[int]
    signed_url: str
    native_url: str
    expires_at: _timestamp_pb2.Timestamp
    headers: _containers.ScalarMap[str, str]
    def __init__(self, signed_url: _Optional[str] = ..., native_url: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., headers: _Optional[_Mapping[str, str]] = ...) -> None: ...

class UploadInputsRequest(_message.Message):
    __slots__ = ["run_id", "project_id", "task_id", "task_spec", "trigger_name", "inputs"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    task_spec: _task_definition_pb2.TaskSpec
    trigger_name: _identifier_pb2.TriggerName
    inputs: _common_pb2.Inputs
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., trigger_name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ..., inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ...) -> None: ...

class UploadInputsResponse(_message.Message):
    __slots__ = ["offloaded_input_data"]
    OFFLOADED_INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    offloaded_input_data: _run_pb2.OffloadedInputData
    def __init__(self, offloaded_input_data: _Optional[_Union[_run_pb2.OffloadedInputData, _Mapping]] = ...) -> None: ...

class PreSignedURLs(_message.Message):
    __slots__ = ["signed_url", "expires_at"]
    SIGNED_URL_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    signed_url: _containers.RepeatedScalarFieldContainer[str]
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, signed_url: _Optional[_Iterable[str]] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreateDownloadLinkRequest(_message.Message):
    __slots__ = ["artifact_type", "expires_in", "action_attempt_id", "app_id", "task_id"]
    ARTIFACT_TYPE_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    ACTION_ATTEMPT_ID_FIELD_NUMBER: _ClassVar[int]
    APP_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    artifact_type: ArtifactType
    expires_in: _duration_pb2.Duration
    action_attempt_id: _identifier_pb2.ActionAttemptIdentifier
    app_id: _app_definition_pb2.Identifier
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, artifact_type: _Optional[_Union[ArtifactType, str]] = ..., expires_in: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., action_attempt_id: _Optional[_Union[_identifier_pb2.ActionAttemptIdentifier, _Mapping]] = ..., app_id: _Optional[_Union[_app_definition_pb2.Identifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class CreateDownloadLinkResponse(_message.Message):
    __slots__ = ["pre_signed_urls"]
    PRE_SIGNED_URLS_FIELD_NUMBER: _ClassVar[int]
    pre_signed_urls: PreSignedURLs
    def __init__(self, pre_signed_urls: _Optional[_Union[PreSignedURLs, _Mapping]] = ...) -> None: ...

class GetActionDataRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDataResponse(_message.Message):
    __slots__ = ["inputs", "outputs"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    inputs: _common_pb2.Inputs
    outputs: _common_pb2.Outputs
    def __init__(self, inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.Outputs, _Mapping]] = ...) -> None: ...

class TailLogsRequest(_message.Message):
    __slots__ = ["action_id", "attempt", "pod_name"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    POD_NAME_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    pod_name: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ..., pod_name: _Optional[str] = ...) -> None: ...

class TailLogsResponse(_message.Message):
    __slots__ = ["logs"]
    class Logs(_message.Message):
        __slots__ = ["lines"]
        LINES_FIELD_NUMBER: _ClassVar[int]
        lines: _containers.RepeatedCompositeFieldContainer[_payload_pb2.LogLine]
        def __init__(self, lines: _Optional[_Iterable[_Union[_payload_pb2.LogLine, _Mapping]]] = ...) -> None: ...
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[TailLogsResponse.Logs]
    def __init__(self, logs: _Optional[_Iterable[_Union[TailLogsResponse.Logs, _Mapping]]] = ...) -> None: ...
