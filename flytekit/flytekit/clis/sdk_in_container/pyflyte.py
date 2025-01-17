import os
import typing

import rich_click as click

from flytekit import configuration
from flytekit.clis.sdk_in_container.backfill import backfill
from flytekit.clis.sdk_in_container.build import build
from flytekit.clis.sdk_in_container.constants import CTX_CONFIG_FILE, CTX_PACKAGES, CTX_VERBOSE
from flytekit.clis.sdk_in_container.fetch import fetch
from flytekit.clis.sdk_in_container.get import get
from flytekit.clis.sdk_in_container.init import init
from flytekit.clis.sdk_in_container.launchplan import launchplan
from flytekit.clis.sdk_in_container.local_cache import local_cache
from flytekit.clis.sdk_in_container.metrics import metrics
from flytekit.clis.sdk_in_container.package import package
from flytekit.clis.sdk_in_container.register import register
from flytekit.clis.sdk_in_container.run import run
from flytekit.clis.sdk_in_container.serialize import serialize
from flytekit.clis.sdk_in_container.serve import serve
from flytekit.clis.sdk_in_container.utils import ErrorHandlingCommand, validate_package
from flytekit.clis.version import info
from flytekit.configuration.file import FLYTECTL_CONFIG_ENV_VAR, FLYTECTL_CONFIG_ENV_VAR_OVERRIDE
from flytekit.configuration.internal import LocalSDK
from flytekit.configuration.plugin import get_plugin
from flytekit.loggers import logger


@click.group("pyflyte", invoke_without_command=True, cls=ErrorHandlingCommand)
@click.option(
    "-v",
    "--verbose",
    required=False,
    help="Show verbose messages and exception traces",
    count=True,
    default=0,
    type=int,
)
@click.option(
    "-k",
    "--pkgs",
    required=False,
    multiple=True,
    callback=validate_package,
    help="Dot-delineated python packages to operate on. Multiple may be specified (can use commas, or specify the "
    "switch multiple times. Please note that this "
    "option will override the option specified in the configuration file, or environment variable",
)
@click.option(
    "-c",
    "--config",
    required=False,
    type=str,
    help="Path to config file for use within container",
)
@click.pass_context
def main(ctx, pkgs: typing.List[str], config: str, verbose: int):
    """
    Entrypoint for all the user commands.
    """
    ctx.obj = dict()

    # Handle package management - get from the command line, the environment variables, then the config file.
    pkgs = pkgs or LocalSDK.WORKFLOW_PACKAGES.read() or []
    if config:
        ctx.obj[CTX_CONFIG_FILE] = config
        cfg = configuration.ConfigFile(config)
        # Set here so that if someone has Config.auto() in their user code, the config here will get used.
        if FLYTECTL_CONFIG_ENV_VAR in os.environ:
            logger.info(
                f"Config file arg {config} will override env var {FLYTECTL_CONFIG_ENV_VAR}: {os.environ[FLYTECTL_CONFIG_ENV_VAR]}"
            )
        os.environ[FLYTECTL_CONFIG_ENV_VAR_OVERRIDE] = config
        if not pkgs:
            pkgs = LocalSDK.WORKFLOW_PACKAGES.read(cfg)
            if pkgs is None:
                pkgs = []

    ctx.obj[CTX_PACKAGES] = pkgs
    ctx.obj[CTX_VERBOSE] = verbose


main.add_command(serialize)
main.add_command(package)
main.add_command(local_cache)
main.add_command(init)
main.add_command(run)
main.add_command(register)
main.add_command(backfill)
main.add_command(serve)
main.add_command(build)
main.add_command(metrics)
main.add_command(launchplan)
main.add_command(fetch)
main.add_command(info)
main.add_command(get)
main.epilog

get_plugin().configure_pyflyte_cli(main)

if __name__ == "__main__":
    main()
