import typing
from datetime import datetime, timedelta

import rich_click as click

from flytekit import WorkflowFailurePolicy
from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context
from flytekit.clis.sdk_in_container.utils import domain_option_dec, project_option_dec
from flytekit.interaction.click_types import DateTimeType, DurationParamType

_backfill_help = """
The backfill command generates and registers a new workflow based on the input launchplan to run an
automated backfill. The workflow can be managed using the Flyte UI and can be canceled, relaunched, and recovered.

    - ``launchplan`` refers to the name of the Launchplan
    - ``launchplan_version`` is optional and should be a valid version for a Launchplan version.
"""


def resolve_backfill_window(
    from_date: datetime = None,
    to_date: datetime = None,
    backfill_window: timedelta = None,
) -> typing.Tuple[datetime, datetime]:
    """
    Resolves the from_date -> to_date
    """
    if from_date and to_date and backfill_window:
        raise click.BadParameter("Setting from-date, to-date and backfill_window at the same time is not allowed.")
    if not (from_date or to_date):
        raise click.BadParameter(
            "One of following pairs are required -> (from-date, to-date) | (from-date, backfill_window) |"
            " (to-date, backfill_window)"
        )
    if from_date and to_date:
        pass
    elif not backfill_window:
        raise click.BadParameter("One of start-date and end-date are needed with duration")
    elif from_date:
        to_date = from_date + backfill_window
    else:
        from_date = to_date - backfill_window
    return from_date, to_date


@click.command("backfill", help=_backfill_help)
@project_option_dec
@domain_option_dec
@click.option(
    "-v",
    "--version",
    required=False,
    type=str,
    default=None,
    help="Version for the registered workflow. If not specified it is auto-derived using the start and end date",
)
@click.option(
    "-n",
    "--execution-name",
    required=False,
    type=str,
    default=None,
    help="Create a named execution for the backfill. This can prevent launching multiple executions.",
)
@click.option(
    "--dry-run",
    required=False,
    type=bool,
    is_flag=True,
    default=False,
    show_default=True,
    help="Just generate the workflow - do not register or execute",
)
@click.option(
    "--parallel/--serial",
    required=False,
    type=bool,
    is_flag=True,
    default=False,
    show_default=True,
    help="All backfill steps can be run in parallel (limited by max-parallelism), if using ``--parallel.``"
    " Else all steps will be run sequentially [``--serial``].",
)
@click.option(
    "--execute/--do-not-execute",
    required=False,
    type=bool,
    is_flag=True,
    default=True,
    show_default=True,
    help="Generate the workflow and register, do not execute",
)
@click.option(
    "--from-date",
    required=False,
    type=DateTimeType(),
    default=None,
    help="Date from which the backfill should begin. Start date is inclusive.",
)
@click.option(
    "--to-date",
    required=False,
    type=DateTimeType(),
    default=None,
    help="Date to which the backfill should run_until. End date is inclusive",
)
@click.option(
    "--backfill-window",
    required=False,
    type=DurationParamType(),
    default=None,
    help="Timedelta for number of days, minutes hours after the from-date or before the to-date to compute the "
    "backfills between. This is needed with from-date / to-date. Optional if both from-date and to-date are "
    "provided",
)
@click.option(
    "--fail-fast/--no-fail-fast",
    required=False,
    type=bool,
    is_flag=True,
    default=True,
    show_default=True,
    help="If set to true, the backfill will fail immediately (WorkflowFailurePolicy.FAIL_IMMEDIATELY) if any of the "
    "backfill steps fail. If set to false, the backfill will continue to run even if some of the backfill steps "
    "fail (WorkflowFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE).",
)
@click.argument(
    "launchplan",
    required=True,
    type=str,
)
@click.argument(
    "launchplan-version",
    required=False,
    type=str,
    default=None,
)
@click.pass_context
def backfill(
    ctx: click.Context,
    project: str,
    domain: str,
    from_date: datetime,
    to_date: datetime,
    backfill_window: timedelta,
    launchplan: str,
    launchplan_version: str,
    dry_run: bool,
    execute: bool,
    parallel: bool,
    execution_name: str,
    version: str,
    fail_fast: bool,
):
    from_date, to_date = resolve_backfill_window(from_date, to_date, backfill_window)
    remote = get_and_save_remote_with_click_context(ctx, project, domain)
    try:
        entity = remote.launch_backfill(
            project=project,
            domain=domain,
            from_date=from_date,
            to_date=to_date,
            launchplan=launchplan,
            launchplan_version=launchplan_version,
            execution_name=execution_name,
            version=version,
            dry_run=dry_run,
            execute=execute,
            parallel=parallel,
            failure_policy=WorkflowFailurePolicy.FAIL_IMMEDIATELY
            if fail_fast
            else WorkflowFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE,
        )
        if dry_run:
            return
        console_url = remote.generate_console_url(entity)
        if execute:
            click.secho(f"\n Execution launched {console_url} to see execution in the console.", fg="green")
            return
        click.secho(f"\n Workflow registered at {console_url}", fg="green")
    except StopIteration as e:
        click.secho(f"{e.value}", fg="red")
