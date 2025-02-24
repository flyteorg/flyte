import flyteidl.admin.launch_plan_pb2 as _launch_plan_pb2
import flyteidl.admin.task_pb2 as _task_pb2
import flyteidl.admin.workflow_pb2 as _workflow_pb2
import flyteidl.core.tasks_pb2 as _core_task_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import workflow_pb2 as _core_workflow_pb2
from flyteidl.core.identifier_pb2 import LAUNCH_PLAN

from flytekit.clis import helpers
from flytekit.clis.helpers import _hydrate_identifier, _hydrate_workflow_template_nodes, hydrate_registration_parameters


def test_parse_args_into_dict():
    sample_args1 = ("input_b=mystr", "input_c=18")
    sample_args2 = ("input_a=mystr===d",)
    sample_args3 = ()
    output = helpers.parse_args_into_dict(sample_args1)
    assert output["input_b"] == "mystr"
    assert output["input_c"] == "18"

    output = helpers.parse_args_into_dict(sample_args2)
    assert output["input_a"] == "mystr===d"

    output = helpers.parse_args_into_dict(sample_args3)
    assert output == {}


def test_strtobool():
    assert not helpers.str2bool("False")
    assert not helpers.str2bool("OFF")
    assert not helpers.str2bool("no")
    assert not helpers.str2bool("0")
    assert helpers.str2bool("t")
    assert helpers.str2bool("true")
    assert helpers.str2bool("stuff")


def test_hydrate_identifier():
    identifier = _hydrate_identifier("project", "domain", "12345", _identifier_pb2.Identifier())
    assert identifier.project == "project"
    assert identifier.domain == "domain"
    assert identifier.version == "12345"

    identifier = _hydrate_identifier(
        "project2", "domain2", "abc", _identifier_pb2.Identifier(project="project", domain="domain", version="12345")
    )
    assert identifier.project == "project"
    assert identifier.domain == "domain"
    assert identifier.version == "12345"

    identifier = _hydrate_identifier(
        "project",
        "domain",
        "12345",
        _identifier_pb2.Identifier(
            project="{{ registration.project }}",
            domain="{{ registration.domain }}",
            version="{{ registration.version }}",
        ),
    )
    assert identifier.project == "project"
    assert identifier.domain == "domain"
    assert identifier.version == "12345"


def test_hydrate_workflow_template():
    workflow_template = _core_workflow_pb2.WorkflowTemplate()
    workflow_template.nodes.append(
        _core_workflow_pb2.Node(
            id="task_node",
            task_node=_core_workflow_pb2.TaskNode(
                reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK)
            ),
        )
    )
    workflow_template.nodes.append(
        _core_workflow_pb2.Node(
            id="launchplan_ref",
            workflow_node=_core_workflow_pb2.WorkflowNode(
                launchplan_ref=_identifier_pb2.Identifier(
                    resource_type=_identifier_pb2.LAUNCH_PLAN,
                    project="project2",
                )
            ),
        )
    )
    workflow_template.nodes.append(
        _core_workflow_pb2.Node(
            id="sub_workflow_ref",
            workflow_node=_core_workflow_pb2.WorkflowNode(
                sub_workflow_ref=_identifier_pb2.Identifier(
                    resource_type=_identifier_pb2.WORKFLOW,
                    project="project2",
                    domain="domain2",
                )
            ),
        )
    )
    workflow_template.nodes.append(
        _core_workflow_pb2.Node(
            id="unchanged",
            task_node=_core_workflow_pb2.TaskNode(
                reference_id=_identifier_pb2.Identifier(
                    resource_type=_identifier_pb2.TASK, project="project2", domain="domain2", version="abc"
                )
            ),
        )
    )
    hydrated_workflow_template = _hydrate_workflow_template_nodes("project", "domain", "12345", workflow_template)
    assert len(hydrated_workflow_template.nodes) == 4
    task_node_identifier = hydrated_workflow_template.nodes[0].task_node.reference_id
    assert task_node_identifier.project == "project"
    assert task_node_identifier.domain == "domain"
    assert task_node_identifier.version == "12345"

    launchplan_ref_identifier = hydrated_workflow_template.nodes[1].workflow_node.launchplan_ref
    assert launchplan_ref_identifier.project == "project2"
    assert launchplan_ref_identifier.domain == "domain"
    assert launchplan_ref_identifier.version == "12345"

    sub_workflow_ref_identifier = hydrated_workflow_template.nodes[2].workflow_node.sub_workflow_ref
    assert sub_workflow_ref_identifier.project == "project2"
    assert sub_workflow_ref_identifier.domain == "domain2"
    assert sub_workflow_ref_identifier.version == "12345"

    unchanged_identifier = hydrated_workflow_template.nodes[3].task_node.reference_id
    assert unchanged_identifier.project == "project2"
    assert unchanged_identifier.domain == "domain2"
    assert unchanged_identifier.version == "abc"


def test_hydrate_workflow_template__branch_node():
    workflow_template = _core_workflow_pb2.WorkflowTemplate()
    branch_node = _core_workflow_pb2.Node(
        id="branch_node",
        branch_node=_core_workflow_pb2.BranchNode(
            if_else=_core_workflow_pb2.IfElseBlock(
                case=_core_workflow_pb2.IfBlock(
                    then_node=_core_workflow_pb2.Node(
                        task_node=_core_workflow_pb2.TaskNode(
                            reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK, name="if_case"),
                        ),
                    )
                ),
                else_node=_core_workflow_pb2.Node(
                    task_node=_core_workflow_pb2.TaskNode(
                        reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK, name="else_node"),
                    ),
                ),
            ),
        ),
    )
    branch_node.branch_node.if_else.other.extend(
        [
            _core_workflow_pb2.IfBlock(
                then_node=_core_workflow_pb2.Node(
                    task_node=_core_workflow_pb2.TaskNode(
                        reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK, name="other_1"),
                    ),
                ),
            ),
            _core_workflow_pb2.IfBlock(
                then_node=_core_workflow_pb2.Node(
                    task_node=_core_workflow_pb2.TaskNode(
                        reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK, name="other_2"),
                    ),
                ),
            ),
        ]
    )
    workflow_template.nodes.append(branch_node)
    hydrated_workflow_template = _hydrate_workflow_template_nodes("project", "domain", "12345", workflow_template)
    if_case_id = hydrated_workflow_template.nodes[0].branch_node.if_else.case.then_node.task_node.reference_id
    assert if_case_id.project == "project"
    assert if_case_id.domain == "domain"
    assert if_case_id.name == "if_case"
    assert if_case_id.version == "12345"

    other_1_id = hydrated_workflow_template.nodes[0].branch_node.if_else.other[0].then_node.task_node.reference_id
    assert other_1_id.project == "project"
    assert other_1_id.domain == "domain"
    assert other_1_id.name == "other_1"
    assert other_1_id.version == "12345"

    other_2_id = hydrated_workflow_template.nodes[0].branch_node.if_else.other[1].then_node.task_node.reference_id
    assert other_2_id.project == "project"
    assert other_2_id.domain == "domain"
    assert other_2_id.name == "other_2"
    assert other_2_id.version == "12345"

    else_id = hydrated_workflow_template.nodes[0].branch_node.if_else.else_node.task_node.reference_id
    assert else_id.project == "project"
    assert else_id.domain == "domain"
    assert else_id.name == "else_node"
    assert else_id.version == "12345"


def test_hydrate_registration_parameters__launch_plan_already_set():
    launch_plan = _launch_plan_pb2.LaunchPlan(
        id=_identifier_pb2.Identifier(
            resource_type=_identifier_pb2.LAUNCH_PLAN,
            project="project2",
            domain="domain2",
            name="workflow_name",
            version="abc",
        ),
        spec=_launch_plan_pb2.LaunchPlanSpec(
            workflow_id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.WORKFLOW,
                project="project2",
                domain="domain2",
                name="workflow_name",
                version="abc",
            )
        ),
    )
    identifier, entity = hydrate_registration_parameters(
        LAUNCH_PLAN,
        "project",
        "domain",
        "12345",
        launch_plan,
    )
    assert identifier == _identifier_pb2.Identifier(
        resource_type=_identifier_pb2.LAUNCH_PLAN,
        project="project2",
        domain="domain2",
        name="workflow_name",
        version="abc",
    )
    assert entity.spec.workflow_id == launch_plan.spec.workflow_id


def test_hydrate_registration_parameters__launch_plan_nothing_set():
    launch_plan = _launch_plan_pb2.LaunchPlan(
        id=_identifier_pb2.Identifier(
            resource_type=_identifier_pb2.LAUNCH_PLAN,
            name="lp_name",
        ),
        spec=_launch_plan_pb2.LaunchPlanSpec(
            workflow_id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.WORKFLOW,
                name="workflow_name",
            )
        ),
    )
    identifier, entity = hydrate_registration_parameters(
        _identifier_pb2.LAUNCH_PLAN,
        "project",
        "domain",
        "12345",
        launch_plan,
    )
    assert identifier == _identifier_pb2.Identifier(
        resource_type=_identifier_pb2.LAUNCH_PLAN,
        project="project",
        domain="domain",
        name="lp_name",
        version="12345",
    )
    assert entity.spec.workflow_id == _identifier_pb2.Identifier(
        resource_type=_identifier_pb2.WORKFLOW,
        project="project",
        domain="domain",
        name="workflow_name",
        version="12345",
    )


def test_hydrate_registration_parameters__task_already_set():
    task = _task_pb2.TaskSpec(
        template=_core_task_pb2.TaskTemplate(
            id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.TASK,
                project="project2",
                domain="domain2",
                name="name",
                version="abc",
            ),
        )
    )
    identifier, entity = hydrate_registration_parameters(_identifier_pb2.TASK, "project", "domain", "12345", task)
    assert (
        identifier
        == _identifier_pb2.Identifier(
            resource_type=_identifier_pb2.TASK,
            project="project2",
            domain="domain2",
            name="name",
            version="abc",
        )
        == entity.template.id
    )


def test_hydrate_registration_parameters__task_nothing_set():
    task = _task_pb2.TaskSpec(
        template=_core_task_pb2.TaskTemplate(
            id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.TASK,
                name="name",
            ),
        )
    )
    identifier, entity = hydrate_registration_parameters(_identifier_pb2.TASK, "project", "domain", "12345", task)
    assert (
        identifier
        == _identifier_pb2.Identifier(
            resource_type=_identifier_pb2.TASK,
            project="project",
            domain="domain",
            name="name",
            version="12345",
        )
        == entity.template.id
    )


def test_hydrate_registration_parameters__workflow_already_set():
    workflow = _workflow_pb2.WorkflowSpec(
        template=_core_workflow_pb2.WorkflowTemplate(
            id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.WORKFLOW,
                project="project2",
                domain="domain2",
                name="name",
                version="abc",
            ),
        )
    )
    identifier, entity = hydrate_registration_parameters(
        _identifier_pb2.WORKFLOW, "project", "domain", "12345", workflow
    )
    assert (
        identifier
        == _identifier_pb2.Identifier(
            resource_type=_identifier_pb2.WORKFLOW,
            project="project2",
            domain="domain2",
            name="name",
            version="abc",
        )
        == entity.template.id
    )


def test_hydrate_registration_parameters__workflow_nothing_set():
    workflow = _workflow_pb2.WorkflowSpec(
        template=_core_workflow_pb2.WorkflowTemplate(
            id=_identifier_pb2.Identifier(
                resource_type=_identifier_pb2.WORKFLOW,
                name="name",
            ),
            nodes=[
                _core_workflow_pb2.Node(
                    id="foo",
                    task_node=_core_workflow_pb2.TaskNode(
                        reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK, name="task1")
                    ),
                )
            ],
        )
    )
    identifier, entity = hydrate_registration_parameters(
        _identifier_pb2.WORKFLOW, "project", "domain", "12345", workflow
    )
    assert (
        identifier
        == _identifier_pb2.Identifier(
            resource_type=_identifier_pb2.WORKFLOW,
            project="project",
            domain="domain",
            name="name",
            version="12345",
        )
        == entity.template.id
    )
    assert len(workflow.template.nodes) == 1
    assert workflow.template.nodes[0].task_node.reference_id == _identifier_pb2.Identifier(
        resource_type=_identifier_pb2.TASK,
        project="project",
        domain="domain",
        name="task1",
        version="12345",
    )


def test_hydrate_registration_parameters__subworkflows():
    workflow_template = _core_workflow_pb2.WorkflowTemplate()
    workflow_template.id.CopyFrom(_identifier_pb2.Identifier(resource_type=_identifier_pb2.WORKFLOW, name="workflow"))

    sub_workflow_template = _core_workflow_pb2.WorkflowTemplate()
    sub_workflow_template.id.CopyFrom(
        _identifier_pb2.Identifier(resource_type=_identifier_pb2.WORKFLOW, name="subworkflow")
    )
    sub_workflow_template.nodes.append(
        _core_workflow_pb2.Node(
            id="task_node",
            task_node=_core_workflow_pb2.TaskNode(
                reference_id=_identifier_pb2.Identifier(resource_type=_identifier_pb2.TASK)
            ),
        )
    )
    workflow_spec = _workflow_pb2.WorkflowSpec(template=workflow_template)
    workflow_spec.sub_workflows.append(sub_workflow_template)

    identifier, entity = hydrate_registration_parameters(
        _identifier_pb2.WORKFLOW, "project", "domain", "12345", workflow_spec
    )
    assert (
        identifier
        == _identifier_pb2.Identifier(
            resource_type=_identifier_pb2.WORKFLOW,
            project="project",
            domain="domain",
            name="workflow",
            version="12345",
        )
        == entity.template.id
    )

    assert entity.sub_workflows[0].id == _identifier_pb2.Identifier(
        resource_type=_identifier_pb2.WORKFLOW,
        project="project",
        domain="domain",
        name="subworkflow",
        version="12345",
    )
