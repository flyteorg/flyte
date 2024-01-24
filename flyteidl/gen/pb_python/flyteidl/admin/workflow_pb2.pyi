from flyteidl.core import compiler_pb2 as _compiler_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import workflow_pb2 as _workflow_pb2
from flyteidl.admin import description_entity_pb2 as _description_entity_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WorkflowCreateRequest(_message.Message):
    __slots__ = ["id", "spec"]
    ID_FIELD_NUMBER: _ClassVar[int]
    SPEC_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    spec: WorkflowSpec
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., spec: _Optional[_Union[WorkflowSpec, _Mapping]] = ...) -> None: ...

class WorkflowCreateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Workflow(_message.Message):
    __slots__ = ["id", "closure", "short_description"]
    ID_FIELD_NUMBER: _ClassVar[int]
    CLOSURE_FIELD_NUMBER: _ClassVar[int]
    SHORT_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    closure: WorkflowClosure
    short_description: str
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., closure: _Optional[_Union[WorkflowClosure, _Mapping]] = ..., short_description: _Optional[str] = ...) -> None: ...

class WorkflowList(_message.Message):
    __slots__ = ["workflows", "token"]
    WORKFLOWS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    workflows: _containers.RepeatedCompositeFieldContainer[Workflow]
    token: str
    def __init__(self, workflows: _Optional[_Iterable[_Union[Workflow, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WorkflowSpec(_message.Message):
    __slots__ = ["template", "sub_workflows", "description"]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    SUB_WORKFLOWS_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    template: _workflow_pb2.WorkflowTemplate
    sub_workflows: _containers.RepeatedCompositeFieldContainer[_workflow_pb2.WorkflowTemplate]
    description: _description_entity_pb2.DescriptionEntity
    def __init__(self, template: _Optional[_Union[_workflow_pb2.WorkflowTemplate, _Mapping]] = ..., sub_workflows: _Optional[_Iterable[_Union[_workflow_pb2.WorkflowTemplate, _Mapping]]] = ..., description: _Optional[_Union[_description_entity_pb2.DescriptionEntity, _Mapping]] = ...) -> None: ...

class WorkflowClosure(_message.Message):
    __slots__ = ["compiled_workflow", "created_at"]
    COMPILED_WORKFLOW_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    compiled_workflow: _compiler_pb2.CompiledWorkflowClosure
    created_at: _timestamp_pb2.Timestamp
    def __init__(self, compiled_workflow: _Optional[_Union[_compiler_pb2.CompiledWorkflowClosure, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class WorkflowErrorExistsDifferentStructure(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class WorkflowErrorExistsIdenticalStructure(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class CreateWorkflowFailureReason(_message.Message):
    __slots__ = ["exists_different_structure", "exists_identical_structure"]
    EXISTS_DIFFERENT_STRUCTURE_FIELD_NUMBER: _ClassVar[int]
    EXISTS_IDENTICAL_STRUCTURE_FIELD_NUMBER: _ClassVar[int]
    exists_different_structure: WorkflowErrorExistsDifferentStructure
    exists_identical_structure: WorkflowErrorExistsIdenticalStructure
    def __init__(self, exists_different_structure: _Optional[_Union[WorkflowErrorExistsDifferentStructure, _Mapping]] = ..., exists_identical_structure: _Optional[_Union[WorkflowErrorExistsIdenticalStructure, _Mapping]] = ...) -> None: ...
