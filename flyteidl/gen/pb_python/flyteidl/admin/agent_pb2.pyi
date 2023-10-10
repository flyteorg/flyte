from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
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

class TaskExecutionMetadata(_message.Message):
    __slots__ = ["task_execution_id", "namespace", "labels", "annotations", "k8s_service_account", "environment_variables"]
    class LabelsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class AnnotationsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    class EnvironmentVariablesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    TASK_EXECUTION_ID_FIELD_NUMBER: _ClassVar[int]
    NAMESPACE_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    ANNOTATIONS_FIELD_NUMBER: _ClassVar[int]
    K8S_SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    ENVIRONMENT_VARIABLES_FIELD_NUMBER: _ClassVar[int]
    task_execution_id: _identifier_pb2.TaskExecutionIdentifier
    namespace: str
    labels: _containers.ScalarMap[str, str]
    annotations: _containers.ScalarMap[str, str]
    k8s_service_account: str
    environment_variables: _containers.ScalarMap[str, str]
    def __init__(self, task_execution_id: _Optional[_Union[_identifier_pb2.TaskExecutionIdentifier, _Mapping]] = ..., namespace: _Optional[str] = ..., labels: _Optional[_Mapping[str, str]] = ..., annotations: _Optional[_Mapping[str, str]] = ..., k8s_service_account: _Optional[str] = ..., environment_variables: _Optional[_Mapping[str, str]] = ...) -> None: ...

class CreateTaskRequest(_message.Message):
    __slots__ = ["inputs", "template", "output_prefix", "task_execution_metadata"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PREFIX_FIELD_NUMBER: _ClassVar[int]
    TASK_EXECUTION_METADATA_FIELD_NUMBER: _ClassVar[int]
    inputs: _literals_pb2.LiteralMap
    template: _tasks_pb2.TaskTemplate
    output_prefix: str
    task_execution_metadata: TaskExecutionMetadata
    def __init__(self, inputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., template: _Optional[_Union[_tasks_pb2.TaskTemplate, _Mapping]] = ..., output_prefix: _Optional[str] = ..., task_execution_metadata: _Optional[_Union[TaskExecutionMetadata, _Mapping]] = ...) -> None: ...

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
    __slots__ = ["state", "outputs", "message"]
    STATE_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    state: State
    outputs: _literals_pb2.LiteralMap
    message: str
    def __init__(self, state: _Optional[_Union[State, str]] = ..., outputs: _Optional[_Union[_literals_pb2.LiteralMap, _Mapping]] = ..., message: _Optional[str] = ...) -> None: ...

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
