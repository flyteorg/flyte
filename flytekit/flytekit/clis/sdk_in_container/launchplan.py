import rich_click as click
from rich.progress import Progress

from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context
from flytekit.clis.sdk_in_container.utils import domain_option_dec, project_option_dec
from flytekit.models.launch_plan import LaunchPlanState

_launchplan_help = """
The launchplan command activates or deactivates a specified or the latest version of the launchplan.
If ``--activate`` is chosen then the previous version of the launchplan will be deactivated.

- ``launchplan`` refers to the name of the Launchplan
- ``launchplan_version`` is optional and should be a valid version for a Launchplan version. If not specified the latest will be used.
"""


@click.command("launchplan", help=_launchplan_help)
@project_option_dec
@domain_option_dec
@click.option(
    "--activate/--deactivate",
    required=True,
    type=bool,
    is_flag=True,
    help="Activate or Deactivate the launchplan",
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
def launchplan(
    ctx: click.Context,
    project: str,
    domain: str,
    activate: bool,
    launchplan: str,
    launchplan_version: str,
):
    remote = get_and_save_remote_with_click_context(ctx, project, domain, data_upload_location="flyte://data")
    with Progress() as progress:
        t1 = progress.add_task(f"[cyan] {'Activating' if activate else 'Deactivating'}...", total=1)
        try:
            progress.start_task(t1)
            launchplan = remote.fetch_launch_plan(
                project=project,
                domain=domain,
                name=launchplan,
                version=launchplan_version,
            )
            progress.advance(t1)

            state = LaunchPlanState.ACTIVE if activate else LaunchPlanState.INACTIVE
            remote.client.update_launch_plan(id=launchplan.id, state=state)
            progress.advance(t1)
            progress.update(t1, completed=True, visible=False)
            click.secho(
                f"\n Launchplan was set to {LaunchPlanState.enum_to_string(state)}: {launchplan.name}:{launchplan.id.version}",
                fg="green",
            )
        except StopIteration as e:
            click.secho(f"{e.value}", fg="red")
