import rich
import rich_click as click
from rich.panel import Panel

from flytekit.clis.sdk_in_container.helpers import get_and_save_remote_with_click_context
from flytekit.remote import FlyteRemote

Content = """
This CLI is meant to be used within a virtual environment that has Flytekit installed. Ideally it is used to iterate on your Flyte workflows and tasks.

Flytekit Version: [cyan]{version}[reset]
Flyte Backend Endpoint: [cyan]{endpoint}
"""


@click.command("info")
@click.pass_context
def info(ctx: click.Context):
    """
    Print out information about the current Flyte Python CLI environment - like the version of Flytekit, backend endpoint
    currently configured, etc.
    """
    import flytekit

    remote: FlyteRemote = get_and_save_remote_with_click_context(ctx, project="flytesnacks", domain="development")
    c = Content.format(version=flytekit.__version__, endpoint=remote.client.url)
    rich.print(Panel(c, title="Flytekit CLI Info", border_style="purple", padding=(1, 1, 1, 1)))
