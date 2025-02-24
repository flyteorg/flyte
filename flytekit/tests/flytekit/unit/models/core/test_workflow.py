from datetime import timedelta

from flyteidl.core import tasks_pb2

from flytekit.extras.accelerators import T4
from flytekit.models import interface as _interface
from flytekit.models import literals as _literals
from flytekit.models import types as _types
from flytekit.models.core import condition as _condition
from flytekit.models.core import identifier as _identifier
from flytekit.models.core import workflow as _workflow
from flytekit.models.task import Resources

_generic_id = _identifier.Identifier(_identifier.ResourceType.WORKFLOW, "project", "domain", "name", "version")


def test_node_metadata():
    obj = _workflow.NodeMetadata(name="node1", timeout=timedelta(seconds=10), retries=_literals.RetryStrategy(0))
    assert obj.timeout.seconds == 10
    assert obj.retries.retries == 0
    obj2 = _workflow.NodeMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.timeout.seconds == 10
    assert obj2.retries.retries == 0


def test_alias():
    obj = _workflow.Alias(var="myvar", alias="myalias")
    assert obj.alias == "myalias"
    assert obj.var == "myvar"
    obj2 = _workflow.Alias.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.alias == "myalias"
    assert obj2.var == "myvar"


def test_workflow_template():
    task = _workflow.TaskNode(reference_id=_generic_id)
    nm = _get_sample_node_metadata()
    int_type = _types.LiteralType(_types.SimpleType.INTEGER)
    wf_metadata = _workflow.WorkflowMetadata()
    wf_metadata_defaults = _workflow.WorkflowMetadataDefaults()
    typed_interface = _interface.TypedInterface(
        {"a": _interface.Variable(int_type, "description1")},
        {"b": _interface.Variable(int_type, "description2"), "c": _interface.Variable(int_type, "description3")},
    )
    wf_node = _workflow.Node(
        id="some:node:id",
        metadata=nm,
        inputs=[],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task,
    )
    obj = _workflow.WorkflowTemplate(
        id=_generic_id,
        metadata=wf_metadata,
        metadata_defaults=wf_metadata_defaults,
        interface=typed_interface,
        nodes=[wf_node],
        outputs=[],
    )
    obj2 = _workflow.WorkflowTemplate.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj


def test_workflow_metadata_failure_policy():
    obj = _workflow.WorkflowMetadata(
        on_failure=_workflow.WorkflowMetadata.OnFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
    )
    obj2 = _workflow.WorkflowMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj.on_failure == _workflow.WorkflowMetadata.OnFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
    assert obj2.on_failure == _workflow.WorkflowMetadata.OnFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE


def test_workflow_metadata():
    obj = _workflow.WorkflowMetadata()
    obj2 = _workflow.WorkflowMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_task_node():
    obj = _workflow.TaskNode(reference_id=_generic_id)
    assert obj.reference_id == _generic_id

    obj2 = _workflow.TaskNode.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.reference_id == _generic_id


def test_workflow_node_lp():
    obj = _workflow.WorkflowNode(launchplan_ref=_generic_id)
    assert obj.launchplan_ref == _generic_id
    assert obj.reference == _generic_id

    obj2 = _workflow.WorkflowNode.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.reference == _generic_id
    assert obj2.launchplan_ref == _generic_id


def test_workflow_node_sw():
    obj = _workflow.WorkflowNode(sub_workflow_ref=_generic_id)
    assert obj.sub_workflow_ref == _generic_id
    assert obj.reference == _generic_id

    obj2 = _workflow.WorkflowNode.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.reference == _generic_id
    assert obj2.sub_workflow_ref == _generic_id


def _get_sample_node_metadata():
    return _workflow.NodeMetadata(name="node1", timeout=timedelta(seconds=10), retries=_literals.RetryStrategy(0))


def test_node_task_with_no_inputs():
    nm = _get_sample_node_metadata()
    task = _workflow.TaskNode(reference_id=_generic_id)

    obj = _workflow.Node(
        id="some:node:id",
        metadata=nm,
        inputs=[],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task,
    )
    assert obj.target == task
    assert obj.id == "some:node:id"
    assert obj.metadata == nm

    obj2 = _workflow.Node.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.target == task
    assert obj2.id == "some:node:id"
    assert obj2.metadata == nm


def test_node_task_with_inputs():
    nm = _get_sample_node_metadata()
    task = _workflow.TaskNode(reference_id=_generic_id)
    bd = _literals.BindingData(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=5)))
    bd2 = _literals.BindingData(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=99)))
    binding = _literals.Binding(var="myvar", binding=bd)
    binding2 = _literals.Binding(var="myothervar", binding=bd2)

    obj = _workflow.Node(
        id="some:node:id",
        metadata=nm,
        inputs=[binding, binding2],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task,
    )
    assert obj.target == task
    assert obj.id == "some:node:id"
    assert obj.metadata == nm
    assert len(obj.inputs) == 2
    assert obj.inputs[0] == binding

    obj2 = _workflow.Node.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.target == task
    assert obj2.id == "some:node:id"
    assert obj2.metadata == nm
    assert len(obj2.inputs) == 2
    assert obj2.inputs[1] == binding2


def test_branch_node():
    nm = _get_sample_node_metadata()
    task = _workflow.TaskNode(reference_id=_generic_id)
    bd = _literals.BindingData(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=5)))
    bd2 = _literals.BindingData(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=99)))
    binding = _literals.Binding(var="myvar", binding=bd)
    binding2 = _literals.Binding(var="myothervar", binding=bd2)

    obj = _workflow.Node(
        id="some:node:id",
        metadata=nm,
        inputs=[binding, binding2],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task,
    )

    bn = _workflow.BranchNode(
        _workflow.IfElseBlock(
            case=_workflow.IfBlock(
                condition=_condition.BooleanExpression(
                    comparison=_condition.ComparisonExpression(
                        _condition.ComparisonExpression.Operator.EQ,
                        _condition.Operand(primitive=_literals.Primitive(integer=5)),
                        _condition.Operand(primitive=_literals.Primitive(integer=2)),
                    )
                ),
                then_node=obj,
            ),
            other=[
                _workflow.IfBlock(
                    condition=_condition.BooleanExpression(
                        conjunction=_condition.ConjunctionExpression(
                            _condition.ConjunctionExpression.LogicalOperator.AND,
                            _condition.BooleanExpression(
                                comparison=_condition.ComparisonExpression(
                                    _condition.ComparisonExpression.Operator.EQ,
                                    _condition.Operand(primitive=_literals.Primitive(integer=5)),
                                    _condition.Operand(primitive=_literals.Primitive(integer=2)),
                                )
                            ),
                            _condition.BooleanExpression(
                                comparison=_condition.ComparisonExpression(
                                    _condition.ComparisonExpression.Operator.EQ,
                                    _condition.Operand(primitive=_literals.Primitive(integer=5)),
                                    _condition.Operand(primitive=_literals.Primitive(integer=2)),
                                )
                            ),
                        )
                    ),
                    then_node=obj,
                )
            ],
            else_node=obj,
        )
    )

    bn2 = _workflow.BranchNode.from_flyte_idl(bn.to_flyte_idl())
    assert bn == bn2
    assert bn.if_else.case.then_node == obj


def test_branch_node_with_none():
    nm = _get_sample_node_metadata()
    task = _workflow.TaskNode(reference_id=_generic_id)
    bd = _literals.BindingData(scalar=_literals.Scalar(none_type=_literals.Void()))
    lt = _literals.Literal(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=99)))
    bd2 = _literals.BindingData(
        scalar=_literals.Scalar(
            union=_literals.Union(value=lt, stored_type=_types.LiteralType(_types.SimpleType.INTEGER))
        )
    )
    binding = _literals.Binding(var="myvar", binding=bd)
    binding2 = _literals.Binding(var="myothervar", binding=bd2)

    obj = _workflow.Node(
        id="some:node:id",
        metadata=nm,
        inputs=[binding, binding2],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task,
    )

    bn = _workflow.BranchNode(
        _workflow.IfElseBlock(
            case=_workflow.IfBlock(
                condition=_condition.BooleanExpression(
                    comparison=_condition.ComparisonExpression(
                        _condition.ComparisonExpression.Operator.EQ,
                        _condition.Operand(scalar=_literals.Scalar(none_type=_literals.Void())),
                        _condition.Operand(primitive=_literals.Primitive(integer=2)),
                    )
                ),
                then_node=obj,
            ),
            other=[
                _workflow.IfBlock(
                    condition=_condition.BooleanExpression(
                        conjunction=_condition.ConjunctionExpression(
                            _condition.ConjunctionExpression.LogicalOperator.AND,
                            _condition.BooleanExpression(
                                comparison=_condition.ComparisonExpression(
                                    _condition.ComparisonExpression.Operator.EQ,
                                    _condition.Operand(scalar=_literals.Scalar(none_type=_literals.Void())),
                                    _condition.Operand(primitive=_literals.Primitive(integer=2)),
                                )
                            ),
                            _condition.BooleanExpression(
                                comparison=_condition.ComparisonExpression(
                                    _condition.ComparisonExpression.Operator.EQ,
                                    _condition.Operand(scalar=_literals.Scalar(none_type=_literals.Void())),
                                    _condition.Operand(primitive=_literals.Primitive(integer=2)),
                                )
                            ),
                        )
                    ),
                    then_node=obj,
                )
            ],
            else_node=obj,
        )
    )

    bn2 = _workflow.BranchNode.from_flyte_idl(bn.to_flyte_idl())
    assert bn == bn2
    assert bn.if_else.case.then_node == obj


def test_task_node_overrides():
    overrides = _workflow.TaskNodeOverrides(
        Resources(
            requests=[Resources.ResourceEntry(Resources.ResourceName.CPU, "1")],
            limits=[Resources.ResourceEntry(Resources.ResourceName.CPU, "2")],
        ),
        tasks_pb2.ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
    )
    assert overrides.resources.requests == [Resources.ResourceEntry(Resources.ResourceName.CPU, "1")]
    assert overrides.resources.limits == [Resources.ResourceEntry(Resources.ResourceName.CPU, "2")]
    assert overrides.extended_resources.gpu_accelerator == T4.to_flyte_idl()

    obj = _workflow.TaskNodeOverrides.from_flyte_idl(overrides.to_flyte_idl())
    assert overrides == obj


def test_task_node_with_overrides():
    task_node = _workflow.TaskNode(
        reference_id=_generic_id,
        overrides=_workflow.TaskNodeOverrides(
            Resources(
                requests=[Resources.ResourceEntry(Resources.ResourceName.CPU, "1")],
                limits=[Resources.ResourceEntry(Resources.ResourceName.CPU, "2")],
            ),
            tasks_pb2.ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
        ),
    )

    assert task_node.overrides.resources.requests == [Resources.ResourceEntry(Resources.ResourceName.CPU, "1")]
    assert task_node.overrides.resources.limits == [Resources.ResourceEntry(Resources.ResourceName.CPU, "2")]
    assert task_node.overrides.extended_resources.gpu_accelerator == T4.to_flyte_idl()

    obj = _workflow.TaskNode.from_flyte_idl(task_node.to_flyte_idl())
    assert task_node == obj
