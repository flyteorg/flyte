import rich_click as click

from flytekit.core.local_cache import LocalTaskCache


@click.group("local-cache")
def local_cache():
    """
    Interact with the local cache.
    """
    pass


@click.command("clear")
def clear_local_cache():
    """
    This command will remove all stored objects from local cache.
    """
    LocalTaskCache.clear()


local_cache.add_command(clear_local_cache)
