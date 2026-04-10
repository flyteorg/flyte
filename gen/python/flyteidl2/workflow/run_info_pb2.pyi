from flyteidl2.task import run_pb2 as _run_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RunInfo(_message.Message):
    __slots__ = ["task_spec_digest", "inputs_uri", "outputs_uri", "run_spec"]
    TASK_SPEC_DIGEST_FIELD_NUMBER: _ClassVar[int]
    INPUTS_URI_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_URI_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    task_spec_digest: str
    inputs_uri: str
    outputs_uri: str
    run_spec: _run_pb2.RunSpec
    def __init__(self, task_spec_digest: _Optional[str] = ..., inputs_uri: _Optional[str] = ..., outputs_uri: _Optional[str] = ..., run_spec: _Optional[_Union[_run_pb2.RunSpec, _Mapping]] = ...) -> None: ...
