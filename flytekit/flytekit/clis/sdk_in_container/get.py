import rich_click as click
from google.protobuf.json_format import MessageToJson
from rich import print
from rich.console import Console
from rich.table import Table

from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context
from flytekit.clis.sdk_in_container.utils import domain_option_dec, project_option_dec
from flytekit.interfaces.cli_identifiers import Identifier
from flytekit.models.admin.common import Sort
from flytekit.models.common import NamedEntityIdentifier
from flytekit.models.core.identifier import ResourceType
from flytekit.models.launch_plan import LaunchPlanState
from flytekit.remote import FlyteRemote


@click.group("get")
@click.pass_context
def get(ctx: click.Context):
    """
    Get a single or multiple remote objects.
    """
    pass


@get.command()
@click.option("--active-only", "--scheduled", is_flag=True, default=False, help="Only return active launchplans.")
@project_option_dec
@domain_option_dec
@click.option("--limit", "-l", type=int, default=1000, help="Limit the number of launchplans returned.")
@click.argument("launchplan-name", type=str, required=False, metavar="LAUNCHPLAN-NAME")
@click.argument("version", type=str, required=False, metavar="LAUNCHPLAN-VERSION")
@click.pass_context
def launchplan(
    ctx: click.Context, project: str, domain: str, limit: int, active_only: bool, launchplan_name: str, version: str
):
    """
    Interact with launchplans.
    """
    remote: FlyteRemote = get_and_save_remote_with_click_context(ctx, project="flytesnacks", domain="development")

    console = Console()
    if launchplan_name:
        if not version:
            lps, _ = remote.client.list_launch_plans_paginated(
                NamedEntityIdentifier(project, domain, name=launchplan_name),
                limit=1,
                sort_by=Sort(key="updated_at", direction=Sort.Direction.DESCENDING),
            )
            if len(lps) > 0:
                version = lps[0].id.version
        lp = remote.client.get_launch_plan(
            Identifier(ResourceType.LAUNCH_PLAN, project, domain, launchplan_name, version)
        )
        j = MessageToJson(lp.to_flyte_idl())
        print(j)
        return

    title = f"LaunchPlans for {project}/{domain}"
    if active_only:
        title += " (active only)"
        lps, _ = remote.client.list_active_launch_plans_paginated(project, domain, limit=limit)
    else:
        lps, _ = remote.client.list_launch_plans_paginated(
            NamedEntityIdentifier(project, domain),
            limit=limit,
            sort_by=Sort(key="updated_at", direction=Sort.Direction.DESCENDING),
        )

    table = Table(title=title)

    table.add_column("Name", justify="right", style="cyan")
    table.add_column("Version", justify="right", style="cyan")
    table.add_column("State", justify="right", style="green")
    table.add_column("Schedule", justify="right", style="green")

    for lp in lps:
        s = LaunchPlanState.enum_to_string(lp.closure.state)
        schedule = "None"
        if lp.spec.entity_metadata.schedule:
            if lp.spec.entity_metadata.schedule.cron_schedule:
                schedule = lp.spec.entity_metadata.schedule.cron_schedule.schedule
            elif lp.spec.entity_metadata.schedule.rate:
                schedule = f"{lp.spec.entity_metadata.schedule.rate.value} {lp.spec.entity_metadata.schedule.rate.unit}"
        table.add_row(lp.id.name, lp.id.version, s, schedule)

    console.print(table)
