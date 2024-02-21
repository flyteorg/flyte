import os
import traceback
import typing
from dataclasses import Field, dataclass, field
from types import MappingProxyType

import grpc
import rich_click as click
from google.protobuf.json_format import MessageToJson

from flytekit.exceptions.base import FlyteException
from flytekit.exceptions.user import FlyteInvalidInputException
from flytekit.loggers import get_level_from_cli_verbosity, logger, upgrade_to_rich_logging

project_option = click.Option(
    param_decls=["-p", "--project"],
    required=False,
    type=str,
    default=os.getenv("FLYTE_DEFAULT_PROJECT", "flytesnacks"),
    show_default=True,
    help="Project to register and run this workflow in. Can also be set through envvar " "``FLYTE_DEFAULT_PROJECT``",
)

domain_option = click.Option(
    param_decls=["-d", "--domain"],
    required=False,
    type=str,
    default=os.getenv("FLYTE_DEFAULT_DOMAIN", "development"),
    show_default=True,
    help="Domain to register and run this workflow in, can also be set through envvar " "``FLYTE_DEFAULT_DOMAIN``",
)

project_option_dec = click.option(
    "-p",
    "--project",
    required=False,
    type=str,
    default=os.getenv("FLYTE_DEFAULT_PROJECT", "flytesnacks"),
    show_default=True,
    help="Project for workflow/launchplan. Can also be set through envvar " "``FLYTE_DEFAULT_PROJECT``",
)

domain_option_dec = click.option(
    "-d",
    "--domain",
    required=False,
    type=str,
    default=os.getenv("FLYTE_DEFAULT_DOMAIN", "development"),
    show_default=True,
    help="Domain for workflow/launchplan, can also be set through envvar " "``FLYTE_DEFAULT_DOMAIN``",
)


def validate_package(ctx, param, values):
    """
    This method will validate the packages passed in by the user. It will check that the packages are in the correct
    format, and will also split the packages if the user passed in a comma separated list.
    """
    pkgs = []
    for val in values:
        if "/" in val or "-" in val or "\\" in val:
            raise click.BadParameter(
                f"Illegal package value {val} for parameter: {param}. Expected for the form [a.b.c]"
            )
        elif "," in val:
            pkgs.extend(val.split(","))
        else:
            pkgs.append(val)
    logger.debug(f"Using packages: {pkgs}")
    return pkgs


def pretty_print_grpc_error(e: grpc.RpcError):
    """
    This method will print the grpc error that us more human readable.
    """
    if isinstance(e, grpc._channel._InactiveRpcError):  # noqa
        click.secho(f"RPC Failed, with Status: {e.code()}", fg="red", bold=True)
        click.secho(f"\tdetails: {e.details()}", fg="magenta", bold=True)
        click.secho(f"\tDebug string {e.debug_error_string()}", dim=True)
    return


def pretty_print_traceback(e):
    """
    This method will print the Traceback of an error.
    """
    if e.__traceback__:
        stack_list = traceback.format_list(traceback.extract_tb(e.__traceback__))
        click.secho("Traceback:", fg="red")
        for i in stack_list:
            click.secho(f"{i}", fg="red")


def pretty_print_exception(e: Exception):
    """
    This method will print the exception in a nice way. It will also check if the exception is a grpc.RpcError and
    print it in a human-readable way.
    """
    if isinstance(e, click.exceptions.Exit):
        raise e

    if isinstance(e, click.ClickException):
        click.secho(e.message, fg="red")
        raise e

    if isinstance(e, FlyteException):
        click.secho(f"Failed with Exception Code: {e._ERROR_CODE}", fg="red")  # noqa
        if isinstance(e, FlyteInvalidInputException):
            click.secho("Request rejected by the API, due to Invalid input.", fg="red")
            click.secho(f"\tInput Request: {MessageToJson(e.request)}", dim=True)

        cause = e.__cause__
        if cause:
            if isinstance(cause, grpc.RpcError):
                pretty_print_grpc_error(cause)
            else:
                pretty_print_traceback(cause)
        return

    if isinstance(e, grpc.RpcError):
        pretty_print_grpc_error(e)
        return

    click.secho(f"Failed with Unknown Exception {type(e)} Reason: {e}", fg="red")  # noqa
    pretty_print_traceback(e)


class ErrorHandlingCommand(click.RichGroup):
    """
    Helper class that wraps the invoke method of a click command to catch exceptions and print them in a nice way.
    """

    def invoke(self, ctx: click.Context) -> typing.Any:
        verbose = ctx.params["verbose"]
        log_level = get_level_from_cli_verbosity(verbose)
        upgrade_to_rich_logging(log_level=log_level)
        try:
            return super().invoke(ctx)
        except Exception as e:
            if verbose > 0:
                click.secho("Verbose mode on")
                if isinstance(e, FlyteException):
                    raise e.with_traceback(None)
                raise e
            pretty_print_exception(e)
            raise SystemExit(e) from e


def make_click_option_field(o: click.Option) -> Field:
    if o.multiple:
        o.help = click.style("Multiple values allowed.", bold=True) + f"{o.help}"
        return field(default_factory=lambda: o.default, metadata={"click.option": o})
    return field(default=o.default, metadata={"click.option": o})


def get_option_from_metadata(metadata: MappingProxyType) -> click.Option:
    return metadata["click.option"]


@dataclass
class PyFlyteParams:
    config_file: typing.Optional[str] = None
    verbose: bool = False
    pkgs: typing.List[str] = field(default_factory=list)

    @classmethod
    def from_dict(cls, d: typing.Dict[str, typing.Any]) -> "PyFlyteParams":
        return cls(**d)
