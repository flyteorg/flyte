import sys
from typing import Tuple, Union

import click
from flyteidl.admin.launch_plan_pb2 import LaunchPlan
from flyteidl.admin.task_pb2 import TaskSpec
from flyteidl.admin.workflow_pb2 import WorkflowSpec
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import workflow_pb2 as _workflow_pb2

from flytekit.configuration import DOMAIN_PLACEHOLDER, PROJECT_PLACEHOLDER, VERSION_PLACEHOLDER


def parse_args_into_dict(input_arguments):
    """
    Takes a tuple like (u'input_b=mystr', u'input_c=18') and returns a dictionary of input name to the
    original string value

    :param Tuple[Text] input_arguments:
    :rtype: dict[Text, Text]
    """

    return {split_arg[0]: split_arg[1] for split_arg in [input_arg.split("=", 1) for input_arg in input_arguments]}


def str2bool(str):
    """
    bool('False') is True in Python, so we need to do some string parsing.  Use the same words in ConfigParser
    :param Text str:
    :rtype: bool
    """
    return str.lower() not in ["false", "0", "off", "no"]


# TODO Deprecated delete after deleting flyte_cli register
def _hydrate_identifier(
    project: str, domain: str, version: str, identifier: _identifier_pb2.Identifier
) -> _identifier_pb2.Identifier:
    if not identifier.project or identifier.project == PROJECT_PLACEHOLDER:
        identifier.project = project

    if not identifier.domain or identifier.domain == DOMAIN_PLACEHOLDER:
        identifier.domain = domain

    if not identifier.version or identifier.version == VERSION_PLACEHOLDER:
        identifier.version = version
    return identifier


# TODO Deprecated delete after deleting flyte_cli register
def _hydrate_node(project: str, domain: str, version: str, node: _workflow_pb2.Node) -> _workflow_pb2.Node:
    if node.HasField("task_node"):
        task_node = node.task_node
        task_node.reference_id.CopyFrom(_hydrate_identifier(project, domain, version, task_node.reference_id))
        node.task_node.CopyFrom(task_node)
    elif node.HasField("workflow_node"):
        workflow_node = node.workflow_node
        if workflow_node.HasField("launchplan_ref"):
            workflow_node.launchplan_ref.CopyFrom(
                _hydrate_identifier(project, domain, version, workflow_node.launchplan_ref)
            )
        elif workflow_node.HasField("sub_workflow_ref"):
            workflow_node.sub_workflow_ref.CopyFrom(
                _hydrate_identifier(project, domain, version, workflow_node.sub_workflow_ref)
            )
    elif node.HasField("branch_node"):
        node.branch_node.if_else.case.then_node.CopyFrom(
            _hydrate_node(project, domain, version, node.branch_node.if_else.case.then_node)
        )
        if node.branch_node.if_else.other is not None:
            others = []
            for if_block in node.branch_node.if_else.other:
                if_block.then_node.CopyFrom(_hydrate_node(project, domain, version, if_block.then_node))
                others.append(if_block)
            del node.branch_node.if_else.other[:]
            node.branch_node.if_else.other.extend(others)
        if node.branch_node.if_else.HasField("else_node"):
            node.branch_node.if_else.else_node.CopyFrom(
                _hydrate_node(project, domain, version, node.branch_node.if_else.else_node)
            )
    return node


# TODO Deprecated delete after deleting flyte_cli register
def _hydrate_workflow_template_nodes(
    project: str, domain: str, version: str, template: _workflow_pb2.WorkflowTemplate
) -> _workflow_pb2.WorkflowTemplate:
    refreshed_nodes = []
    for node in template.nodes:
        node = _hydrate_node(project, domain, version, node)
        refreshed_nodes.append(node)
    # Reassign nodes with the newly hydrated ones.
    del template.nodes[:]
    template.nodes.extend(refreshed_nodes)
    return template


# TODO Deprecated delete after deleting flyte_cli register
def hydrate_registration_parameters(
    resource_type: int,
    project: str,
    domain: str,
    version: str,
    entity: Union[LaunchPlan, WorkflowSpec, TaskSpec],
) -> Tuple[_identifier_pb2.Identifier, Union[LaunchPlan, WorkflowSpec, TaskSpec]]:
    """
    This is called at registration time to fill out identifier fields (e.g. project, domain, version) that are mutable.
    """

    if resource_type == _identifier_pb2.LAUNCH_PLAN:
        identifier = _hydrate_identifier(project, domain, version, entity.id)
        entity.spec.workflow_id.CopyFrom(_hydrate_identifier(project, domain, version, entity.spec.workflow_id))
        return identifier, entity

    identifier = _hydrate_identifier(project, domain, version, entity.template.id)
    entity.template.id.CopyFrom(identifier)
    if identifier.resource_type == _identifier_pb2.TASK:
        return identifier, entity

    # Workflows (the only possible entity type at this point) are a little more complicated.
    # Workflow nodes that are defined inline with the workflows will be missing project/domain/version so we fill those
    # in now.
    # (entity is of type flyteidl.admin.workflow_pb2.WorkflowSpec)
    entity.template.CopyFrom(_hydrate_workflow_template_nodes(project, domain, version, entity.template))
    refreshed_sub_workflows = []
    for sub_workflow in entity.sub_workflows:
        refreshed_sub_workflow = _hydrate_workflow_template_nodes(project, domain, version, sub_workflow)
        refreshed_sub_workflow.id.CopyFrom(_hydrate_identifier(project, domain, version, refreshed_sub_workflow.id))
        refreshed_sub_workflows.append(refreshed_sub_workflow)
    # Reassign subworkflows with the newly hydrated ones.
    del entity.sub_workflows[:]
    entity.sub_workflows.extend(refreshed_sub_workflows)
    return identifier, entity


def display_help_with_error(ctx: click.Context, message: str):
    click.echo(f"{ctx.get_help()}\n")
    click.secho(message, fg="red")
    sys.exit(1)
