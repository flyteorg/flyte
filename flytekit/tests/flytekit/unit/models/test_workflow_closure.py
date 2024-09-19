from datetime import timedelta

from flyteidl.core import tasks_pb2

from flytekit.extras.accelerators import T4
from flytekit.models import interface as _interface
from flytekit.models import literals as _literals
from flytekit.models import task as _task
from flytekit.models import types as _types
from flytekit.models import workflow_closure as _workflow_closure
from flytekit.models.admin.workflow import WorkflowSpec
from flytekit.models.core import identifier as _identifier
from flytekit.models.core import workflow as _workflow
from flytekit.models.documentation import Description, Documentation, SourceCode


def test_workflow_closure():
    int_type = _types.LiteralType(_types.SimpleType.INTEGER)
    typed_interface = _interface.TypedInterface(
        {"a": _interface.Variable(int_type, "description1")},
        {"b": _interface.Variable(int_type, "description2"), "c": _interface.Variable(int_type, "description3")},
    )

    b0 = _literals.Binding(
        "a",
        _literals.BindingData(scalar=_literals.Scalar(primitive=_literals.Primitive(integer=5))),
    )
    b1 = _literals.Binding("b", _literals.BindingData(promise=_types.OutputReference("my_node", "b")))
    b2 = _literals.Binding("c", _literals.BindingData(promise=_types.OutputReference("my_node", "c")))

    node_metadata = _workflow.NodeMetadata(
        name="node1", timeout=timedelta(seconds=10), retries=_literals.RetryStrategy(0)
    )

    task_metadata = _task.TaskMetadata(
        True,
        _task.RuntimeMetadata(_task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "python"),
        timedelta(days=1),
        _literals.RetryStrategy(3),
        True,
        "0.1.1b0",
        "This is deprecated!",
        True,
        "A",
    )

    cpu_resource = _task.Resources.ResourceEntry(_task.Resources.ResourceName.CPU, "1")
    resources = _task.Resources(requests=[cpu_resource], limits=[cpu_resource])

    task = _task.TaskTemplate(
        _identifier.Identifier(_identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        task_metadata,
        typed_interface,
        {"a": 1, "b": {"c": 2, "d": 3}},
        container=_task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            resources,
            {},
            {},
        ),
        extended_resources=tasks_pb2.ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
    )

    task_node = _workflow.TaskNode(task.id)
    node = _workflow.Node(
        id="my_node",
        metadata=node_metadata,
        inputs=[b0],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=task_node,
    )

    template = _workflow.WorkflowTemplate(
        id=_identifier.Identifier(_identifier.ResourceType.WORKFLOW, "project", "domain", "name", "version"),
        metadata=_workflow.WorkflowMetadata(),
        metadata_defaults=_workflow.WorkflowMetadataDefaults(),
        interface=typed_interface,
        nodes=[node],
        outputs=[b1, b2],
    )

    obj = _workflow_closure.WorkflowClosure(workflow=template, tasks=[task])
    assert len(obj.tasks) == 1

    obj2 = _workflow_closure.WorkflowClosure.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2

    short_description = "short"
    long_description = Description(value="long", icon_link="http://icon")
    source_code = SourceCode(link="https://github.com/flyteorg/flytekit")
    docs = Documentation(
        short_description=short_description, long_description=long_description, source_code=source_code
    )

    workflow_spec = WorkflowSpec(template=template, sub_workflows=[], docs=docs)
    assert WorkflowSpec.from_flyte_idl(workflow_spec.to_flyte_idl()) == workflow_spec
    assert workflow_spec.docs.short_description == short_description
    assert workflow_spec.docs.long_description == long_description
    assert workflow_spec.docs.source_code == source_code
