from datetime import datetime

import rich_click as click
import yaml
from flyteidl.admin.execution_pb2 import WorkflowExecutionGetMetricsRequest
from flyteidl.core.identifier_pb2 import WorkflowExecutionIdentifier

from flytekit.clis.sdk_in_container.constants import CTX_DOMAIN, CTX_PROJECT
from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context

CTX_DEPTH = "depth"

_dump_help = """
The dump command aggregates workflow execution metrics and displays them. This aggregation is meant to provide an easy
to understand breakdown of where time is spent in a hierarchical manner.

- execution_id refers to the id of the workflow execution
"""

_explain_help = """
The explain command prints each individual execution span and the associated timestamps and Flyte entity reference.
This breakdown provides precise information into exactly how and when Flyte processes a workflow execution.

- execution_id refers to the id of the workflow execution
"""


@click.group("metrics")
@click.option(
    "-d",
    "--depth",
    required=False,
    type=int,
    default=-1,
    help="The depth of Flyte entity hierarchy to traverse when computing metrics for this execution",
)
@click.option(
    "-p",
    "--project",
    required=False,
    type=str,
    default="flytesnacks",
    help="The project of the workflow execution",
)
@click.option(
    "-d",
    "--domain",
    required=False,
    type=str,
    default="development",
    help="The domain of the workflow execution",
)
@click.pass_context
def metrics(ctx: click.Context, depth, domain, project):
    ctx.obj[CTX_DEPTH] = depth
    ctx.obj[CTX_DOMAIN] = domain
    ctx.obj[CTX_PROJECT] = project


@click.command("dump", help=_dump_help)
@click.argument("execution_id", type=str)
@click.pass_context
def metrics_dump(
    ctx: click.Context,
    execution_id: str,
):
    depth = ctx.obj[CTX_DEPTH]
    domain = ctx.obj[CTX_DOMAIN]
    project = ctx.obj[CTX_PROJECT]

    # retrieve remote
    remote = get_and_save_remote_with_click_context(ctx, project, domain)
    sync_client = remote.client

    # retrieve workflow execution metrics
    workflow_execution_id = WorkflowExecutionIdentifier(project=project, domain=domain, name=execution_id)

    request = WorkflowExecutionGetMetricsRequest(id=workflow_execution_id, depth=depth)
    response = sync_client.get_execution_metrics(request)

    # aggregate spans and print
    id, info = aggregate_reference_span(response.span)
    yaml.emitter.Emitter.process_tag = lambda self, *args, **kw: None
    print(yaml.dump({id: info}, indent=2))


def aggregate_reference_span(span):
    id = ""
    id_type = span.WhichOneof("id")
    if id_type == "workflow_id":
        id = span.workflow_id.name
    elif id_type == "node_id":
        id = span.node_id.node_id
    elif id_type == "task_id":
        id = span.task_id.retry_attempt

    spans = aggregate_spans(span.spans)
    return id, spans


def aggregate_spans(spans):
    breakdown = {}

    tasks = {}
    nodes = {}
    workflows = {}

    for span in spans:
        id_type = span.WhichOneof("id")
        if id_type == "operation_id":
            operation_id = span.operation_id

            start_time = datetime.fromtimestamp(span.start_time.seconds + span.start_time.nanos / 1e9)
            end_time = datetime.fromtimestamp(span.end_time.seconds + span.end_time.nanos / 1e9)
            total_time = (end_time - start_time).total_seconds()

            if operation_id in breakdown:
                breakdown[operation_id] += total_time
            else:
                breakdown[operation_id] = total_time
        else:
            id, underlying_span = aggregate_reference_span(span)

            if id_type == "workflow_id":
                workflows[id] = underlying_span
            elif id_type == "node_id":
                nodes[id] = underlying_span
            elif id_type == "task_id":
                tasks[id] = underlying_span

            for operation_id, total_time in underlying_span["breakdown"].items():
                if operation_id in breakdown:
                    breakdown[operation_id] += total_time
                else:
                    breakdown[operation_id] = total_time

    span = {"breakdown": breakdown}

    if len(tasks) > 0:
        span["task_attempts"] = tasks
    if len(nodes) > 0:
        span["nodes"] = nodes
    if len(workflows) > 0:
        span["workflows"] = workflows

    return span


@click.command("explain", help=_explain_help)
@click.argument("execution_id", type=str)
@click.pass_context
def metrics_explain(
    ctx: click.Context,
    execution_id: str,
):
    depth = ctx.obj[CTX_DEPTH]
    domain = ctx.obj[CTX_DOMAIN]
    project = ctx.obj[CTX_PROJECT]

    # retrieve remote
    remote = get_and_save_remote_with_click_context(ctx, project, domain)
    sync_client = remote.client

    # retrieve workflow execution metrics
    workflow_execution_id = WorkflowExecutionIdentifier(project=project, domain=domain, name=execution_id)

    request = WorkflowExecutionGetMetricsRequest(id=workflow_execution_id, depth=depth)
    response = sync_client.get_execution_metrics(request)

    # print execution spans
    print(
        "{:25s}{:25s}{:25s} {:>8s}    {:s}".format(
            "operation", "start_timestamp", "end_timestamp", "duration", "entity"
        )
    )
    print("-" * 140)

    print_span(response.span, -1, "")


def print_span(span, indent, identifier):
    start_time = datetime.fromtimestamp(span.start_time.seconds + span.start_time.nanos / 1e9)
    end_time = datetime.fromtimestamp(span.end_time.seconds + span.end_time.nanos / 1e9)

    id_type = span.WhichOneof("id")
    span_identifier = ""

    if id_type == "operation_id":
        indent_str = ""
        for i in range(indent):
            indent_str += "  "

        print(
            "{:25s}{:25s}{:25s} {:7.2f}s    {:s}{:s}".format(
                span.operation_id,
                start_time.strftime("%m-%d %H:%M:%S.%f"),
                end_time.strftime("%m-%d %H:%M:%S.%f"),
                (end_time - start_time).total_seconds(),
                indent_str,
                identifier,
            )
        )

        span_identifier = identifier + "/" + span.operation_id
    else:
        if id_type == "workflow_id":
            span_identifier = "workflow/" + span.workflow_id.name
        elif id_type == "node_id":
            span_identifier = "node/" + span.node_id.node_id
        elif id_type == "task_id":
            span_identifier = "task/" + str(span.task_id.retry_attempt)

    for under_span in span.spans:
        print_span(under_span, indent + 1, span_identifier)


metrics.add_command(metrics_dump)
metrics.add_command(metrics_explain)
