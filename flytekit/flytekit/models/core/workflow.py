import datetime
import typing

from flyteidl.core import tasks_pb2
from flyteidl.core import workflow_pb2 as _core_workflow

from flytekit.models import common as _common
from flytekit.models import interface as _interface
from flytekit.models import types as type_models
from flytekit.models.core import condition as _condition
from flytekit.models.core import identifier as _identifier
from flytekit.models.literals import Binding as _Binding
from flytekit.models.literals import RetryStrategy as _RetryStrategy
from flytekit.models.task import Resources


class IfBlock(_common.FlyteIdlEntity):
    def __init__(self, condition, then_node):
        """
        Defines a condition and the execution unit that should be executed if the condition is satisfied.

        :param flytekit.models.core.condition.BooleanExpression condition:
        :param Node then_node:
        """

        self._condition = condition
        self._then_node = then_node

    @property
    def condition(self):
        """
        :rtype: flytekit.models.core.condition.BooleanExpression
        """
        return self._condition

    @property
    def then_node(self):
        """
        :rtype: Node
        """
        return self._then_node

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.IfBlock
        """
        return _core_workflow.IfBlock(condition=self.condition.to_flyte_idl(), then_node=self.then_node.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        return cls(
            condition=_condition.BooleanExpression.from_flyte_idl(pb2_object.condition),
            then_node=Node.from_flyte_idl(pb2_object.then_node),
        )


class IfElseBlock(_common.FlyteIdlEntity):
    def __init__(self, case, other=None, else_node=None, error=None):
        """
        Defines a series of if/else blocks. The first branch whose condition evaluates to true is the one to execute.
        If no conditions were satisfied, the else_node or the error will execute.

        :param IfBlock case:
        :param list[IfBlock] other:
        :param Node else_node:
        :param type_models.Error error:
        """
        self._case = case
        self._other = other
        self._else_node = else_node
        self._error = error

    @property
    def case(self):
        """
        First condition to evaluate.

        :rtype: IfBlock
        """

        return self._case

    @property
    def other(self):
        """
        Additional branches to evaluate.

        :rtype: list[IfBlock]
        """

        return self._other

    @property
    def else_node(self):
        """
        The node to execute in case none of the branches were taken.

        :rtype: Node
        """

        return self._else_node

    @property
    def error(self):
        """
        An error to throw in case none of the branches were taken.

        :rtype: flytekit.models.types.Error
        """

        return self._error

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.IfElseBlock
        """
        return _core_workflow.IfElseBlock(
            case=self.case.to_flyte_idl(),
            other=[a.to_flyte_idl() for a in self.other] if self.other else None,
            else_node=self.else_node.to_flyte_idl() if self.else_node else None,
            error=self.error.to_flyte_idl() if self.error else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        return cls(
            case=IfBlock.from_flyte_idl(pb2_object.case),
            other=[IfBlock.from_flyte_idl(a) for a in pb2_object.other],
            else_node=Node.from_flyte_idl(pb2_object.else_node) if pb2_object.HasField("else_node") else None,
            error=type_models.Error.from_flyte_idl(pb2_object.error) if pb2_object.HasField("error") else None,
        )


class BranchNode(_common.FlyteIdlEntity):
    def __init__(self, if_else: IfElseBlock):
        """
        BranchNode is a special node that alter the flow of the workflow graph. It allows the control flow to branch at
        runtime based on a series of conditions that get evaluated on various parameters (e.g. inputs, primtives).

        :param IfElseBlock if_else:
        """

        self._if_else = if_else

    @property
    def if_else(self) -> IfElseBlock:
        """
        :rtype: IfElseBlock
        """

        return self._if_else

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.BranchNode
        """
        return _core_workflow.BranchNode(if_else=self.if_else.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, pb2_objct):
        return cls(if_else=IfElseBlock.from_flyte_idl(pb2_objct.if_else))


class NodeMetadata(_common.FlyteIdlEntity):
    def __init__(
        self,
        name,
        timeout=None,
        retries=None,
        interruptible: typing.Optional[bool] = None,
        cacheable: typing.Optional[bool] = None,
        cache_version: typing.Optional[str] = None,
        cache_serializable: typing.Optional[bool] = None,
    ):
        """
        Defines extra information about the Node.

        :param Text name: Friendly name for the Node.
        :param datetime.timedelta timeout: [Optional] Overall timeout for a task.
        :param flytekit.models.literals.RetryStrategy retries: [Optional] Number of retries per task.
        :param bool interruptible: Can be safely interrupted during execution.
        :param cacheable: Indicates that this nodes outputs should be cached.
        :param cache_version: The version of the cached data.
        :param cacheable: Indicates that cache operations on this node should be serialized.
        """
        self._name = name
        self._timeout = timeout if timeout is not None else datetime.timedelta()
        self._retries = retries if retries is not None else _RetryStrategy(0)
        self._interruptible = interruptible
        self._cacheable = cacheable
        self._cache_version = cache_version
        self._cache_serializable = cache_serializable

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    @property
    def timeout(self):
        """
        :rtype: datetime.timedelta
        """
        return self._timeout

    @property
    def retries(self):
        """
        :rtype: flytekit.models.literals.RetryStrategy
        """
        return self._retries

    @property
    def interruptible(self) -> typing.Optional[bool]:
        return self._interruptible

    @property
    def cacheable(self) -> typing.Optional[bool]:
        return self._cacheable

    @property
    def cache_version(self) -> typing.Optional[str]:
        return self._cache_version

    @property
    def cache_serializable(self) -> typing.Optional[bool]:
        return self._cache_serializable

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.NodeMetadata
        """
        node_metadata = _core_workflow.NodeMetadata(
            name=self.name,
            retries=self.retries.to_flyte_idl(),
            interruptible=self.interruptible,
            cacheable=self.cacheable,
            cache_version=self.cache_version,
            cache_serializable=self.cache_serializable,
        )
        if self.timeout:
            node_metadata.timeout.FromTimedelta(self.timeout)
        return node_metadata

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        return cls(
            pb2_object.name,
            pb2_object.timeout.ToTimedelta(),
            _RetryStrategy.from_flyte_idl(pb2_object.retries),
            pb2_object.interruptible if pb2_object.HasField("interruptible") else None,
            pb2_object.cacheable if pb2_object.HasField("cacheable") else None,
            pb2_object.cache_version if pb2_object.HasField("cache_version") else None,
            pb2_object.cache_serializable if pb2_object.HasField("cache_serializable") else None,
        )


class SignalCondition(_common.FlyteIdlEntity):
    def __init__(self, signal_id: str, type: type_models.LiteralType, output_variable_name: str):
        """
        Represents a dependency on an signal from a user.

        :param signal_id: The node id of the signal, also the signal name.
        :param type:
        """
        self._signal_id = signal_id
        self._type = type
        self._output_variable_name = output_variable_name

    @property
    def signal_id(self) -> str:
        return self._signal_id

    @property
    def type(self) -> type_models.LiteralType:
        return self._type

    @property
    def output_variable_name(self) -> str:
        return self._output_variable_name

    def to_flyte_idl(self) -> _core_workflow.SignalCondition:
        return _core_workflow.SignalCondition(
            signal_id=self.signal_id, type=self.type.to_flyte_idl(), output_variable_name=self.output_variable_name
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_workflow.SignalCondition):
        return cls(
            signal_id=pb2_object.signal_id,
            type=type_models.LiteralType.from_flyte_idl(pb2_object.type),
            output_variable_name=pb2_object.output_variable_name,
        )


class ApproveCondition(_common.FlyteIdlEntity):
    def __init__(self, signal_id: str):
        """
        Represents a dependency on an signal from a user.

        :param signal_id: The node id of the signal, also the signal name.
        """
        self._signal_id = signal_id

    @property
    def signal_id(self) -> str:
        return self._signal_id

    def to_flyte_idl(self) -> _core_workflow.ApproveCondition:
        return _core_workflow.ApproveCondition(signal_id=self.signal_id)

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_workflow.ApproveCondition):
        return cls(signal_id=pb2_object.signal_id)


class SleepCondition(_common.FlyteIdlEntity):
    def __init__(self, duration: datetime.timedelta):
        """
        A sleep condition.
        """
        self._duration = duration

    @property
    def duration(self) -> datetime.timedelta:
        return self._duration

    def to_flyte_idl(self) -> _core_workflow.SleepCondition:
        sc = _core_workflow.SleepCondition()
        sc.duration.FromTimedelta(self.duration)
        return sc

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_workflow.SignalCondition) -> "SleepCondition":
        return cls(duration=pb2_object.duration.ToTimedelta())


class GateNode(_common.FlyteIdlEntity):
    def __init__(
        self,
        signal: typing.Optional[SignalCondition] = None,
        sleep: typing.Optional[SleepCondition] = None,
        approve: typing.Optional[ApproveCondition] = None,
    ):
        self._signal = signal
        self._sleep = sleep
        self._approve = approve

    @property
    def signal(self) -> typing.Optional[SignalCondition]:
        return self._signal

    @property
    def sleep(self) -> typing.Optional[SignalCondition]:
        return self._sleep

    @property
    def approve(self) -> typing.Optional[ApproveCondition]:
        return self._approve

    @property
    def condition(self) -> typing.Union[SignalCondition, SleepCondition, ApproveCondition]:
        return self.signal or self.sleep or self.approve

    def to_flyte_idl(self) -> _core_workflow.GateNode:
        return _core_workflow.GateNode(
            signal=self.signal.to_flyte_idl() if self.signal else None,
            sleep=self.sleep.to_flyte_idl() if self.sleep else None,
            approve=self.approve.to_flyte_idl() if self.approve else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_workflow.GateNode) -> "GateNode":
        return cls(
            signal=SignalCondition.from_flyte_idl(pb2_object.signal) if pb2_object.HasField("signal") else None,
            sleep=SleepCondition.from_flyte_idl(pb2_object.sleep) if pb2_object.HasField("sleep") else None,
            approve=ApproveCondition.from_flyte_idl(pb2_object.approve) if pb2_object.HasField("approve") else None,
        )


class ArrayNode(_common.FlyteIdlEntity):
    def __init__(self, node: "Node", parallelism=None, min_successes=None, min_success_ratio=None) -> None:
        """
        TODO: docstring
        """
        self._node = node
        self._parallelism = parallelism
        # TODO either min_successes or min_success_ratio should be set
        self._min_successes = min_successes
        self._min_success_ratio = min_success_ratio

    @property
    def node(self) -> "Node":
        return self._node

    def to_flyte_idl(self) -> _core_workflow.ArrayNode:
        return _core_workflow.ArrayNode(
            node=self._node.to_flyte_idl() if self._node is not None else None,
            parallelism=self._parallelism,
            min_successes=self._min_successes,
            min_success_ratio=self._min_success_ratio,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object) -> "ArrayNode":
        return cls(
            Node.from_flyte_idl(pb2_object.node),
            pb2_object.parallelism,
            pb2_object.min_successes,
            pb2_object.min_success_ratio,
        )


class Node(_common.FlyteIdlEntity):
    def __init__(
        self,
        id,
        metadata,
        inputs,
        upstream_node_ids,
        output_aliases,
        task_node=None,
        workflow_node=None,
        branch_node=None,
        gate_node: typing.Optional[GateNode] = None,
        array_node: typing.Optional[ArrayNode] = None,
    ):
        """
        A Workflow graph Node. One unit of execution in the graph. Each node can be linked to a Task,
        a Workflow or a branch node.  One of the nodes must be specified.

        :param Text id: A workflow-level unique identifier that identifies this node in the workflow. "inputs" and
            "outputs" are reserved node ids that cannot be used by other nodes.
        :param NodeMetadata metadata: Extra metadata about the node.
        :param list[flytekit.models.literals.Binding] inputs: Specifies how to bind the underlying
            interface's inputs.  All required inputs specified in the underlying interface must be fulfilled.
        :param list[Text] upstream_node_ids: Specifies execution dependency for this node ensuring it will
            only get scheduled to run after all its upstream nodes have completed. This node will have
            an implicit dependency on any node that appears in inputs field.
        :param list[Alias] output_aliases: A node can define aliases for a subset of its outputs. This
            is particularly useful if different nodes need to conform to the same interface (e.g. all branches in
            a branch node). Downstream nodes must refer to this node's outputs using the alias if one is specified.
        :param TaskNode task_node: [Optional] Information about the Task to execute in this node.
        :param WorkflowNode workflow_node: [Optional] Information about the Workflow to execute in this mode.
        :param BranchNode branch_node: [Optional] Information about the branch node to evaluate in this node.
        """

        self._id = id
        self._metadata = metadata
        self._inputs = inputs
        self._upstream_node_ids = upstream_node_ids
        # TODO: For proper graph handling, we need to keep track of the node objects themselves, not just the node IDs
        self._output_aliases = output_aliases
        self._task_node = task_node
        self._workflow_node = workflow_node
        self._branch_node = branch_node
        self._gate_node = gate_node
        self._array_node = array_node

    @property
    def id(self):
        """
        A workflow-level unique identifier that identifies this node in the workflow. "inputs" and
        "outputs" are reserved node ids that cannot be used by other nodes.

        :rtype: Text
        """
        return self._id

    @property
    def metadata(self):
        """
        Extra metadata about the node.

        :rtype: NodeMetadata
        """
        return self._metadata

    @property
    def inputs(self):
        """
        Specifies how to bind the underlying interface's inputs.  All required inputs specified
        in the underlying interface must be fulfilled.

        :rtype: list[flytekit.models.literals.Binding]
        """
        return self._inputs

    @property
    def upstream_node_ids(self):
        """
        [Optional] Specifies execution dependency for this node ensuring it will
        only get scheduled to run after all its upstream nodes have completed. This node will have
        an implicit dependency on any node that appears in inputs field.

        :rtype: list[Text]
        """
        return self._upstream_node_ids

    @property
    def output_aliases(self):
        """
        [Optional] A node can define aliases for a subset of its outputs. This
        is particularly useful if different nodes need to conform to the same interface (e.g. all branches in
        a branch node). Downstream nodes must refer to this node's outputs using the alias if one is specified.

        :rtype: list[Alias]
        """
        return self._output_aliases

    @property
    def task_node(self):
        """
        [Optional] Information about the Task to execute in this node.

        :rtype: TaskNode
        """
        return self._task_node

    @property
    def workflow_node(self):
        """
        [Optional] Information about the Workflow to execute in this mode.

        :rtype: WorkflowNode
        """
        return self._workflow_node

    @property
    def branch_node(self):
        """
        [Optional] Information about the branch node to evaluate in this node.

        :rtype: BranchNode
        """
        return self._branch_node

    @property
    def gate_node(self) -> typing.Optional[GateNode]:
        return self._gate_node

    @property
    def array_node(self) -> typing.Optional[ArrayNode]:
        return self._array_node

    @property
    def target(self):
        """
        :rtype: T
        """
        return self.task_node or self.workflow_node or self.branch_node

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.Node
        """
        return _core_workflow.Node(
            id=self.id,
            metadata=self.metadata.to_flyte_idl() if self.metadata is not None else None,
            inputs=[i.to_flyte_idl() for i in self.inputs],
            upstream_node_ids=self.upstream_node_ids,
            output_aliases=[a.to_flyte_idl() for a in self.output_aliases],
            task_node=self.task_node.to_flyte_idl() if self.task_node is not None else None,
            workflow_node=self.workflow_node.to_flyte_idl() if self.workflow_node is not None else None,
            branch_node=self.branch_node.to_flyte_idl() if self.branch_node is not None else None,
            gate_node=self.gate_node.to_flyte_idl() if self.gate_node else None,
            array_node=self.array_node.to_flyte_idl() if self.array_node else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.Node pb2_object:
        :rtype: Node
        """
        return cls(
            id=pb2_object.id,
            metadata=NodeMetadata.from_flyte_idl(pb2_object.metadata),
            inputs=[_Binding.from_flyte_idl(b) for b in pb2_object.inputs],
            upstream_node_ids=pb2_object.upstream_node_ids,
            output_aliases=[Alias.from_flyte_idl(a) for a in pb2_object.output_aliases],
            task_node=TaskNode.from_flyte_idl(pb2_object.task_node) if pb2_object.HasField("task_node") else None,
            workflow_node=WorkflowNode.from_flyte_idl(pb2_object.workflow_node)
            if pb2_object.HasField("workflow_node")
            else None,
            branch_node=BranchNode.from_flyte_idl(pb2_object.branch_node)
            if pb2_object.HasField("branch_node")
            else None,
            gate_node=GateNode.from_flyte_idl(pb2_object.gate_node) if pb2_object.HasField("gate_node") else None,
            array_node=ArrayNode.from_flyte_idl(pb2_object.array_node) if pb2_object.HasField("array_node") else None,
        )


class TaskNodeOverrides(_common.FlyteIdlEntity):
    def __init__(
        self, resources: typing.Optional[Resources], extended_resources: typing.Optional[tasks_pb2.ExtendedResources]
    ):
        self._resources = resources
        self._extended_resources = extended_resources

    @property
    def resources(self) -> Resources:
        return self._resources

    @property
    def extended_resources(self) -> tasks_pb2.ExtendedResources:
        return self._extended_resources

    def to_flyte_idl(self):
        return _core_workflow.TaskNodeOverrides(
            resources=self.resources.to_flyte_idl() if self.resources is not None else None,
            extended_resources=self.extended_resources,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        resources = Resources.from_flyte_idl(pb2_object.resources)
        extended_resources = pb2_object.extended_resources if pb2_object.HasField("extended_resources") else None
        if bool(resources.requests) or bool(resources.limits):
            return cls(resources=resources, extended_resources=extended_resources)
        return cls(resources=None, extended_resources=extended_resources)


class TaskNode(_common.FlyteIdlEntity):
    def __init__(self, reference_id, overrides: typing.Optional[TaskNodeOverrides] = None):
        """
        Refers to the task that the Node is to execute.
        This is currently a oneof in protobuf, but there's only one option currently.
        This code should be updated when more options are available.

        :param flytekit.models.core.identifier.Identifier reference_id: A globally unique identifier for the task.
        :param flyteidl.core.workflow_pb2.TaskNodeOverrides:
        """
        self._reference_id = reference_id
        self._overrides = overrides

    @property
    def reference_id(self):
        """
        A globally unique identifier for the task. This should map to the identifier in Flyte Admin.

        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._reference_id

    @property
    def overrides(self) -> TaskNodeOverrides:
        return self._overrides

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.TaskNode
        """
        return _core_workflow.TaskNode(
            reference_id=self.reference_id.to_flyte_idl(),
            overrides=self.overrides.to_flyte_idl() if self.overrides is not None else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.TaskNode pb2_object:
        :rtype: TaskNode
        """
        overrides = TaskNodeOverrides.from_flyte_idl(pb2_object.overrides)
        if overrides.resources is None:
            overrides = None
        return cls(
            reference_id=_identifier.Identifier.from_flyte_idl(pb2_object.reference_id),
            overrides=overrides,
        )


class WorkflowNode(_common.FlyteIdlEntity):
    def __init__(self, launchplan_ref=None, sub_workflow_ref=None):
        """
        Refers to a the workflow the node is to execute. One of the references must be supplied.

        :param flytekit.models.core.identifier.Identifier launchplan_ref: [Optional] A globally unique identifier for
            the launch plan. Should map to Admin.
        :param flytekit.models.core.identifier.Identifier sub_workflow_ref: [Optional] Reference to a subworkflow,
            that should be defined with the compiler context.
        """
        self._launchplan_ref = launchplan_ref
        self._sub_workflow_ref = sub_workflow_ref

    @property
    def launchplan_ref(self):
        """
        [Optional] A globally unique identifier for the launch plan.  Should map to Admin.

        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._launchplan_ref

    @property
    def sub_workflow_ref(self):
        """
        [Optional] Reference to a subworkflow, that should be defined with the compiler context.

        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._sub_workflow_ref

    @property
    def reference(self):
        """
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self.launchplan_ref or self.sub_workflow_ref

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.WorkflowNode
        """
        return _core_workflow.WorkflowNode(
            launchplan_ref=self.launchplan_ref.to_flyte_idl() if self.launchplan_ref else None,
            sub_workflow_ref=self.sub_workflow_ref.to_flyte_idl() if self.sub_workflow_ref else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.WorkflowNode pb2_object:

        :rtype: WorkflowNode
        """
        if pb2_object.HasField("launchplan_ref"):
            return cls(launchplan_ref=_identifier.Identifier.from_flyte_idl(pb2_object.launchplan_ref))
        else:
            return cls(sub_workflow_ref=_identifier.Identifier.from_flyte_idl(pb2_object.sub_workflow_ref))


class WorkflowMetadata(_common.FlyteIdlEntity):
    class OnFailurePolicy(object):
        """
        Defines the execution behavior of the workflow when a failure is detected.

        Attributes:
            FAIL_IMMEDIATELY                        Instructs the system to fail as soon as a node fails in the
                                                    workflow. It'll automatically abort all currently running nodes and
                                                    clean up resources before finally marking the workflow executions as failed.

            FAIL_AFTER_EXECUTABLE_NODES_COMPLETE    Instructs the system to make as much progress as it can. The system
                                                    will not alter the dependencies of the execution graph so any node
                                                    that depend on the failed node will not be run. Other nodes that will
                                                    be executed to completion before cleaning up resources and marking
                                                    the workflow execution as failed.
        """

        FAIL_IMMEDIATELY = _core_workflow.WorkflowMetadata.FAIL_IMMEDIATELY
        FAIL_AFTER_EXECUTABLE_NODES_COMPLETE = _core_workflow.WorkflowMetadata.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE

    def __init__(self, on_failure=None):
        """
        Metadata for the workflow.

        :param on_failure flytekit.models.core.workflow.WorkflowMetadata.OnFailurePolicy: [Optional] The execution policy when the workflow detects a failure.
        """
        self._on_failure = on_failure

    @property
    def on_failure(self):
        """
        :rtype: flytekit.models.core.workflow.WorkflowMetadata.OnFailurePolicy
        """
        return self._on_failure

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.WorkflowMetadata
        """
        workflow_metadata = _core_workflow.WorkflowMetadata()
        if self.on_failure:
            workflow_metadata.on_failure = self.on_failure
        return workflow_metadata

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.WorkflowMetadata pb2_object:

        :rtype: WorkflowMetadata
        """
        return cls(
            on_failure=pb2_object.on_failure
            if pb2_object.on_failure
            else WorkflowMetadata.OnFailurePolicy.FAIL_IMMEDIATELY
        )


class WorkflowMetadataDefaults(_common.FlyteIdlEntity):
    def __init__(self, interruptible=None):
        """
        Metadata Defaults for the workflow.
        """
        self._interruptible = interruptible

    @property
    def interruptible(self):
        return self._interruptible

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.WorkflowMetadataDefaults
        """
        return _core_workflow.WorkflowMetadataDefaults(interruptible=self._interruptible)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.WorkflowMetadataDefaults pb2_object:

        :rtype: WorkflowMetadata
        """
        return cls(interruptible=pb2_object.interruptible)


class WorkflowTemplate(_common.FlyteIdlEntity):
    def __init__(
        self,
        id,
        metadata,
        metadata_defaults,
        interface,
        nodes,
        outputs,
        failure_node=None,
    ):
        """
        A workflow template encapsulates all the task, branch, and subworkflow nodes to run a statically analyzable,
        directed acyclic graph. It contains also metadata that tells the system how to execute the workflow (i.e.
        the AWS IAM role to run with).

        :param flytekit.models.core.identifier.Identifier id: This is an autogenerated id by the system. The id is
            globally unique across Flyte.
        :param WorkflowMetadata metadata: This contains information on how to run the workflow.
        :param WorkflowMetadataDefaults metadata_defaults: This contains the default information on how to run the workflow.
        :param flytekit.models.interface.TypedInterface interface: Defines a strongly typed interface for the
            Workflow (inputs, outputs).  This can include some optional parameters.
        :param list[Node] nodes: A list of nodes. In addition, "globals" is a special reserved node id that
            can be used to consume workflow inputs
        :param list[flytekit.models.literals.Binding] outputs: A list of output bindings that specify how to construct
            workflow outputs. Bindings can pull node outputs or specify literals. All workflow outputs specified in
            the interface field must be bound
            in order for the workflow to be validated. A workflow has an implicit dependency on all of its nodes
            to execute successfully in order to bind final outputs.
        :param Node failure_node: [Optional] A catch-all node. This node is executed whenever the execution
            engine determines the workflow has failed. The interface of this node must match the Workflow interface
            with an additional input named "error" of type pb.lyft.flyte.core.Error.
        """
        self._id = id
        self._metadata = metadata
        self._metadata_defaults = metadata_defaults
        self._interface = interface
        self._nodes = nodes
        self._outputs = outputs
        self._failure_node = failure_node

    @property
    def id(self):
        """
        This is an autogenerated id by the system. The id is globally unique across Flyte.

        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def metadata(self):
        """
        This contains information on how to run the workflow.

        :rtype: WorkflowMetadata
        """
        return self._metadata

    @property
    def metadata_defaults(self):
        """
        This contains information on how to run the workflow.

        :rtype: WorkflowMetadataDefaults
        """
        return self._metadata_defaults

    @property
    def interface(self):
        """
        Defines a strongly typed interface for the Workflow (inputs, outputs). This can include some optional
        parameters.

        :rtype: flytekit.models.interface.TypedInterface
        """
        return self._interface

    @property
    def nodes(self):
        """
        A list of nodes. In addition, "globals" is a special reserved node id that can be used to consume
        workflow inputs.

        :rtype: list[Node]
        """
        return self._nodes

    @property
    def outputs(self):
        """
        A list of output bindings that specify how to construct workflow outputs. Bindings can
        pull node outputs or specify literals. All workflow outputs specified in the interface field must be bound
        in order for the workflow to be validated. A workflow has an implicit dependency on all of its nodes
        to execute successfully in order to bind final outputs.

        :rtype: list[flytekit.models.literals.Binding]
        """
        return self._outputs

    @property
    def failure_node(self):
        """
        Node failure_node: A catch-all node. This node is executed whenever the execution engine determines the
        workflow has failed. The interface of this node must match the Workflow interface with an additional input
        named "error" of type pb.lyft.flyte.core.Error.

        :rtype: Node
        """
        return self._failure_node

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.WorkflowTemplate
        """
        return _core_workflow.WorkflowTemplate(
            id=self.id.to_flyte_idl(),
            metadata=self.metadata.to_flyte_idl(),
            metadata_defaults=self.metadata_defaults.to_flyte_idl(),
            interface=self.interface.to_flyte_idl(),
            nodes=[n.to_flyte_idl() for n in self.nodes],
            outputs=[o.to_flyte_idl() for o in self.outputs],
            failure_node=self.failure_node.to_flyte_idl() if self.failure_node is not None else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.WorkflowTemplate pb2_object:

        :rtype: WorkflowTemplate
        """
        return cls(
            id=_identifier.Identifier.from_flyte_idl(pb2_object.id),
            metadata=WorkflowMetadata.from_flyte_idl(pb2_object.metadata),
            metadata_defaults=WorkflowMetadataDefaults.from_flyte_idl(pb2_object.metadata_defaults),
            interface=_interface.TypedInterface.from_flyte_idl(pb2_object.interface),
            nodes=[Node.from_flyte_idl(n) for n in pb2_object.nodes],
            outputs=[_Binding.from_flyte_idl(b) for b in pb2_object.outputs],
            failure_node=Node.from_flyte_idl(pb2_object.failure_node) if pb2_object.HasField("failure_node") else None,
        )


class Alias(_common.FlyteIdlEntity):
    def __init__(self, var, alias):
        """
        Links a variable to an alias.

        :param Text var: Must match one of the output variable names on a node.
        :param Text alias: A workflow-level unique alias that downstream nodes can refer to in their input.
        """
        self._var = var
        self._alias = alias

    @property
    def var(self):
        """
        Must match one of the output variable names on a node.

        :rtype: Text
        """
        return self._var

    @property
    def alias(self):
        """
        A workflow-level unique alias that downstream nodes can refer to in their input.

        :rtype: Text
        """
        return self._alias

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.workflow_pb2.Alias
        """
        return _core_workflow.Alias(var=self.var, alias=self.alias)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.workflow_pb2.Alias pb2_object:

        :return: Alias
        """
        return cls(pb2_object.var, pb2_object.alias)
