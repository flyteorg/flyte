from flyteidl.core import condition_pb2 as _condition_pb2
from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import interface_pb2 as _interface_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _tasks_pb2
from flyteidl.core import types_pb2 as _types_pb2
from flyteidl.core import security_pb2 as _security_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Alias(_message.Message):
    __slots__ = ["alias", "var"]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    VAR_FIELD_NUMBER: _ClassVar[int]
    alias: str
    var: str
    def __init__(self, var: _Optional[str] = ..., alias: _Optional[str] = ...) -> None: ...

class ApproveCondition(_message.Message):
    __slots__ = ["signal_id"]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    signal_id: str
    def __init__(self, signal_id: _Optional[str] = ...) -> None: ...

class BranchNode(_message.Message):
    __slots__ = ["if_else"]
    IF_ELSE_FIELD_NUMBER: _ClassVar[int]
    if_else: IfElseBlock
    def __init__(self, if_else: _Optional[_Union[IfElseBlock, _Mapping]] = ...) -> None: ...

class GateNode(_message.Message):
    __slots__ = ["approve", "signal", "sleep"]
    APPROVE_FIELD_NUMBER: _ClassVar[int]
    SIGNAL_FIELD_NUMBER: _ClassVar[int]
    SLEEP_FIELD_NUMBER: _ClassVar[int]
    approve: ApproveCondition
    signal: SignalCondition
    sleep: SleepCondition
    def __init__(self, approve: _Optional[_Union[ApproveCondition, _Mapping]] = ..., signal: _Optional[_Union[SignalCondition, _Mapping]] = ..., sleep: _Optional[_Union[SleepCondition, _Mapping]] = ...) -> None: ...

class IfBlock(_message.Message):
    __slots__ = ["condition", "then_node"]
    CONDITION_FIELD_NUMBER: _ClassVar[int]
    THEN_NODE_FIELD_NUMBER: _ClassVar[int]
    condition: _condition_pb2.BooleanExpression
    then_node: Node
    def __init__(self, condition: _Optional[_Union[_condition_pb2.BooleanExpression, _Mapping]] = ..., then_node: _Optional[_Union[Node, _Mapping]] = ...) -> None: ...

class IfElseBlock(_message.Message):
    __slots__ = ["case", "else_node", "error", "other"]
    CASE_FIELD_NUMBER: _ClassVar[int]
    ELSE_NODE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    OTHER_FIELD_NUMBER: _ClassVar[int]
    case: IfBlock
    else_node: Node
    error: _types_pb2.Error
    other: _containers.RepeatedCompositeFieldContainer[IfBlock]
    def __init__(self, case: _Optional[_Union[IfBlock, _Mapping]] = ..., other: _Optional[_Iterable[_Union[IfBlock, _Mapping]]] = ..., else_node: _Optional[_Union[Node, _Mapping]] = ..., error: _Optional[_Union[_types_pb2.Error, _Mapping]] = ...) -> None: ...

class Node(_message.Message):
    __slots__ = ["branch_node", "gate_node", "id", "inputs", "metadata", "output_aliases", "task_node", "upstream_node_ids", "workflow_node"]
    BRANCH_NODE_FIELD_NUMBER: _ClassVar[int]
    GATE_NODE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_ALIASES_FIELD_NUMBER: _ClassVar[int]
    TASK_NODE_FIELD_NUMBER: _ClassVar[int]
    UPSTREAM_NODE_IDS_FIELD_NUMBER: _ClassVar[int]
    WORKFLOW_NODE_FIELD_NUMBER: _ClassVar[int]
    branch_node: BranchNode
    gate_node: GateNode
    id: str
    inputs: _containers.RepeatedCompositeFieldContainer[_literals_pb2.Binding]
    metadata: NodeMetadata
    output_aliases: _containers.RepeatedCompositeFieldContainer[Alias]
    task_node: TaskNode
    upstream_node_ids: _containers.RepeatedScalarFieldContainer[str]
    workflow_node: WorkflowNode
    def __init__(self, id: _Optional[str] = ..., metadata: _Optional[_Union[NodeMetadata, _Mapping]] = ..., inputs: _Optional[_Iterable[_Union[_literals_pb2.Binding, _Mapping]]] = ..., upstream_node_ids: _Optional[_Iterable[str]] = ..., output_aliases: _Optional[_Iterable[_Union[Alias, _Mapping]]] = ..., task_node: _Optional[_Union[TaskNode, _Mapping]] = ..., workflow_node: _Optional[_Union[WorkflowNode, _Mapping]] = ..., branch_node: _Optional[_Union[BranchNode, _Mapping]] = ..., gate_node: _Optional[_Union[GateNode, _Mapping]] = ...) -> None: ...

class NodeMetadata(_message.Message):
    __slots__ = ["interruptible", "name", "retries", "timeout"]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    RETRIES_FIELD_NUMBER: _ClassVar[int]
    TIMEOUT_FIELD_NUMBER: _ClassVar[int]
    interruptible: bool
    name: str
    retries: _literals_pb2.RetryStrategy
    timeout: _duration_pb2.Duration
    def __init__(self, name: _Optional[str] = ..., timeout: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., retries: _Optional[_Union[_literals_pb2.RetryStrategy, _Mapping]] = ..., interruptible: bool = ...) -> None: ...

class SignalCondition(_message.Message):
    __slots__ = ["output_variable_name", "signal_id", "type"]
    OUTPUT_VARIABLE_NAME_FIELD_NUMBER: _ClassVar[int]
    SIGNAL_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    output_variable_name: str
    signal_id: str
    type: _types_pb2.LiteralType
    def __init__(self, signal_id: _Optional[str] = ..., type: _Optional[_Union[_types_pb2.LiteralType, _Mapping]] = ..., output_variable_name: _Optional[str] = ...) -> None: ...

class SleepCondition(_message.Message):
    __slots__ = ["duration"]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    duration: _duration_pb2.Duration
    def __init__(self, duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class TaskNode(_message.Message):
    __slots__ = ["overrides", "reference_id"]
    OVERRIDES_FIELD_NUMBER: _ClassVar[int]
    REFERENCE_ID_FIELD_NUMBER: _ClassVar[int]
    overrides: TaskNodeOverrides
    reference_id: _identifier_pb2.Identifier
    def __init__(self, reference_id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., overrides: _Optional[_Union[TaskNodeOverrides, _Mapping]] = ...) -> None: ...

class TaskNodeOverrides(_message.Message):
    __slots__ = ["resources"]
    RESOURCES_FIELD_NUMBER: _ClassVar[int]
    resources: _tasks_pb2.Resources
    def __init__(self, resources: _Optional[_Union[_tasks_pb2.Resources, _Mapping]] = ...) -> None: ...

class WorkflowMetadata(_message.Message):
    __slots__ = ["on_failure", "quality_of_service", "tags"]
    class OnFailurePolicy(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    class TagsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    FAIL_AFTER_EXECUTABLE_NODES_COMPLETE: WorkflowMetadata.OnFailurePolicy
    FAIL_IMMEDIATELY: WorkflowMetadata.OnFailurePolicy
    ON_FAILURE_FIELD_NUMBER: _ClassVar[int]
    QUALITY_OF_SERVICE_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    on_failure: WorkflowMetadata.OnFailurePolicy
    quality_of_service: _execution_pb2.QualityOfService
    tags: _containers.ScalarMap[str, str]
    def __init__(self, quality_of_service: _Optional[_Union[_execution_pb2.QualityOfService, _Mapping]] = ..., on_failure: _Optional[_Union[WorkflowMetadata.OnFailurePolicy, str]] = ..., tags: _Optional[_Mapping[str, str]] = ...) -> None: ...

class WorkflowMetadataDefaults(_message.Message):
    __slots__ = ["interruptible"]
    INTERRUPTIBLE_FIELD_NUMBER: _ClassVar[int]
    interruptible: bool
    def __init__(self, interruptible: bool = ...) -> None: ...

class WorkflowNode(_message.Message):
    __slots__ = ["launchplan_ref", "sub_workflow_ref"]
    LAUNCHPLAN_REF_FIELD_NUMBER: _ClassVar[int]
    SUB_WORKFLOW_REF_FIELD_NUMBER: _ClassVar[int]
    launchplan_ref: _identifier_pb2.Identifier
    sub_workflow_ref: _identifier_pb2.Identifier
    def __init__(self, launchplan_ref: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., sub_workflow_ref: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class WorkflowTemplate(_message.Message):
    __slots__ = ["failure_node", "id", "interface", "metadata", "metadata_defaults", "nodes", "outputs"]
    FAILURE_NODE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    METADATA_DEFAULTS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    NODES_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    failure_node: Node
    id: _identifier_pb2.Identifier
    interface: _interface_pb2.TypedInterface
    metadata: WorkflowMetadata
    metadata_defaults: WorkflowMetadataDefaults
    nodes: _containers.RepeatedCompositeFieldContainer[Node]
    outputs: _containers.RepeatedCompositeFieldContainer[_literals_pb2.Binding]
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ..., metadata: _Optional[_Union[WorkflowMetadata, _Mapping]] = ..., interface: _Optional[_Union[_interface_pb2.TypedInterface, _Mapping]] = ..., nodes: _Optional[_Iterable[_Union[Node, _Mapping]]] = ..., outputs: _Optional[_Iterable[_Union[_literals_pb2.Binding, _Mapping]]] = ..., failure_node: _Optional[_Union[Node, _Mapping]] = ..., metadata_defaults: _Optional[_Union[WorkflowMetadataDefaults, _Mapping]] = ...) -> None: ...
