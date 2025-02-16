import configparser as _configparser
import importlib as _importlib
import os
import os as _os
import stat as _stat
import sys as _sys
from dataclasses import replace
from typing import Callable, Dict, List, Tuple, Union

import click as _click
import requests as _requests
from flyteidl.admin import launch_plan_pb2 as _launch_plan_pb2
from flyteidl.admin import task_pb2 as _task_pb2
from flyteidl.admin import workflow_pb2 as _workflow_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _core_tasks_pb2
from flyteidl.core import workflow_pb2 as _core_workflow_pb2
from google.protobuf.json_format import MessageToJson
from google.protobuf.pyext.cpp_message import GeneratedProtocolMessageType as _GeneratedProtocolMessageType

from flytekit import __version__, configuration
from flytekit.clients import friendly as _friendly_client
from flytekit.clis.helpers import hydrate_registration_parameters
from flytekit.core import utils
from flytekit.core.context_manager import FlyteContextManager
from flytekit.exceptions import user as _user_exceptions
from flytekit.interfaces import cli_identifiers
from flytekit.models import common as _common_models
from flytekit.models import filters as _filters
from flytekit.models import launch_plan as _launch_plan
from flytekit.models import literals as _literals
from flytekit.models import named_entity as _named_entity
from flytekit.models.admin import common as _admin_common
from flytekit.models.common import AuthRole as _AuthRole
from flytekit.models.common import RawOutputDataConfig as _RawOutputDataConfig
from flytekit.models.core import execution as _core_execution_models
from flytekit.models.core import identifier as _core_identifier
from flytekit.models.matchable_resource import ClusterResourceAttributes as _ClusterResourceAttributes
from flytekit.models.matchable_resource import ExecutionClusterLabel as _ExecutionClusterLabel
from flytekit.models.matchable_resource import ExecutionQueueAttributes as _ExecutionQueueAttributes
from flytekit.models.matchable_resource import MatchableResource as _MatchableResource
from flytekit.models.matchable_resource import MatchingAttributes as _MatchingAttributes
from flytekit.models.matchable_resource import PluginOverride as _PluginOverride
from flytekit.models.matchable_resource import PluginOverrides as _PluginOverrides
from flytekit.models.project import Project as _Project
from flytekit.models.schedule import Schedule as _Schedule
from flytekit.tools.fast_registration import get_additional_distribution_loc as _get_additional_distribution_loc

try:  # Python 3
    import urllib.parse as _urlparse
except ImportError:  # Python 2
    import urlparse as _urlparse

_tt = str

# Similar to how kubectl has a config file in the users home directory, this Flyte CLI will also look for one.
# The format of this config file is the same as a workflow's config file, except that the relevant fields are different.
# Please see the example.config file
_default_config_file_dir = ".flyte"
_default_config_file_name = "config"


def _welcome_message():
    _click.secho(
        "\n################################################################################################################################",
        bold=True,
    )
    _click.secho(
        "# flyte-cli is being deprecated in favor of flytectl. More details about flytectl in https://docs.flyte.org/projects/flytectl/ #",
        bold=True,
    )
    _click.secho(
        "################################################################################################################################\n",
        bold=True,
    )
    _click.secho("Welcome to Flyte CLI! Version: {}\n".format(_tt(__version__)), bold=True)


def _get_user_filepath_home():
    return _os.path.expanduser("~")


def _get_config_file_path():
    home = _get_user_filepath_home()
    return _os.path.join(home, _default_config_file_dir, _default_config_file_name)


def _detect_default_config_file():
    config_file = _get_config_file_path()
    if _get_user_filepath_home() and _os.path.exists(config_file):
        _click.secho("Using default config file at {}".format(_tt(config_file)), fg="blue")
        return config_file
    else:
        _click.secho(
            """Config file not found at default location, relying on environment variables instead.
                        To setup your config file run 'flyte-cli setup-config'""",
            fg="blue",
        )
    return None


def _get_io_string(literal_map, verbose=False):
    """
    :param flytekit.models.literals.LiteralMap literal_map:
    :param bool verbose:
    :rtype: Text
    """
    return str(literal_map)


def _fetch_and_stringify_literal_map(path, verbose=False):
    """
    :param Text path:
    :param bool verbose:
    :rtype: Text
    """
    ctx = FlyteContextManager.current_context()
    with utils.AutoDeletingTempDir("flytecli") as tmp:
        try:
            fname = tmp.get_named_tempfile("literalmap.pb")
            ctx.file_access.get_data(path, fname)
            literal_map = _literals.LiteralMap.from_flyte_idl(
                utils.load_proto_from_file(_literals_pb2.LiteralMap, fname)
            )
            return _get_io_string(literal_map, verbose=verbose)
        except Exception:
            return "Failed to pull data from {}. Do you have permissions?".format(path)


def _prefix_lines(prefix, txt):
    """
    :param Text prefix:
    :param Text txt:
    :rtype: Text
    """
    return "\n{}".format(prefix).join(txt.splitlines())


def _secho_workflow_status(status, nl=True):
    red_phases = {
        _core_execution_models.WorkflowExecutionPhase.FAILED,
        _core_execution_models.WorkflowExecutionPhase.ABORTED,
        _core_execution_models.WorkflowExecutionPhase.FAILING,
        _core_execution_models.WorkflowExecutionPhase.TIMED_OUT,
    }
    yellow_phases = {
        _core_execution_models.WorkflowExecutionPhase.QUEUED,
        _core_execution_models.WorkflowExecutionPhase.UNDEFINED,
    }
    green_phases = {
        _core_execution_models.WorkflowExecutionPhase.SUCCEEDED,
        _core_execution_models.WorkflowExecutionPhase.SUCCEEDING,
    }
    if status in red_phases:
        fg = "red"
    elif status in yellow_phases:
        fg = "yellow"
    elif status in green_phases:
        fg = "green"
    else:
        fg = "blue"

    _click.secho(
        "{:10} ".format(_tt(_core_execution_models.WorkflowExecutionPhase.enum_to_string(status))),
        bold=True,
        fg=fg,
        nl=nl,
    )


def _secho_node_execution_status(status, nl=True):
    red_phases = {
        _core_execution_models.NodeExecutionPhase.FAILING,
        _core_execution_models.NodeExecutionPhase.FAILED,
        _core_execution_models.NodeExecutionPhase.ABORTED,
        _core_execution_models.NodeExecutionPhase.TIMED_OUT,
    }
    yellow_phases = {
        _core_execution_models.NodeExecutionPhase.QUEUED,
        _core_execution_models.NodeExecutionPhase.UNDEFINED,
    }
    green_phases = {_core_execution_models.NodeExecutionPhase.SUCCEEDED}
    if status in red_phases:
        fg = "red"
    elif status in yellow_phases:
        fg = "yellow"
    elif status in green_phases:
        fg = "green"
    else:
        fg = "blue"

    _click.secho(
        "{:10} ".format(_tt(_core_execution_models.NodeExecutionPhase.enum_to_string(status))),
        bold=True,
        fg=fg,
        nl=nl,
    )


def _secho_task_execution_status(status, nl=True):
    red_phases = {
        _core_execution_models.TaskExecutionPhase.ABORTED,
        _core_execution_models.TaskExecutionPhase.FAILED,
    }
    yellow_phases = {
        _core_execution_models.TaskExecutionPhase.QUEUED,
        _core_execution_models.TaskExecutionPhase.UNDEFINED,
        _core_execution_models.TaskExecutionPhase.RUNNING,
    }
    green_phases = {_core_execution_models.TaskExecutionPhase.SUCCEEDED}
    if status in red_phases:
        fg = "red"
    elif status in yellow_phases:
        fg = "yellow"
    elif status in green_phases:
        fg = "green"
    else:
        fg = "blue"

    _click.secho(
        "{:10} ".format(_tt(_core_execution_models.TaskExecutionPhase.enum_to_string(status))),
        bold=True,
        fg=fg,
        nl=nl,
    )


def _secho_one_execution(ex, urns_only):
    if not urns_only:
        _click.echo(
            "{:100} {:40} {:40}".format(
                _tt(cli_identifiers.WorkflowExecutionIdentifier.promote_from_model(ex.id)),
                _tt(ex.id.name),
                _tt(ex.spec.launch_plan.name),
            ),
            nl=False,
        )
        _secho_workflow_status(ex.closure.phase)
    else:
        _click.echo(
            "{:100}".format(_tt(cli_identifiers.WorkflowExecutionIdentifier.promote_from_model(ex.id))),
            nl=True,
        )


def _terminate_one_execution(client, urn, cause, shouldPrint=True):
    if shouldPrint:
        _click.echo("{:100} {:40}".format(_tt(urn), _tt(cause)))
    client.terminate_execution(cli_identifiers.WorkflowExecutionIdentifier.from_python_std(urn), cause)


def _update_one_launch_plan(client: _friendly_client.SynchronousFlyteClient, urn, state):
    if state == "active":
        state = _launch_plan.LaunchPlanState.ACTIVE
    else:
        state = _launch_plan.LaunchPlanState.INACTIVE
    client.update_launch_plan(cli_identifiers.Identifier.from_python_std(urn), state)
    _click.echo("Successfully updated {}".format(_tt(urn)))


def _render_schedule_expr(lp):
    sched_expr = "NONE"
    if lp.spec.entity_metadata.schedule and lp.spec.entity_metadata.schedule.cron_expression:
        sched_expr = "cron({cron_expr})".format(cron_expr=_tt(lp.spec.entity_metadata.schedule.cron_expression))
    elif lp.spec.entity_metadata.schedule and lp.spec.entity_metadata.schedule.rate:
        sched_expr = "rate({unit}={value})".format(
            unit=_tt(_Schedule.FixedRateUnit.enum_to_string(lp.spec.entity_metadata.schedule.rate.unit)),
            value=_tt(lp.spec.entity_metadata.schedule.rate.value),
        )
    return "{:30}".format(sched_expr)


def _get_client(host: str, insecure: bool) -> _friendly_client.SynchronousFlyteClient:
    parent_ctx = _click.get_current_context(silent=True)
    kwargs = {}
    if parent_ctx.obj["cacert"]:
        kwargs["root_certificates"] = parent_ctx.obj["cacert"]
    cfg = parent_ctx.obj["config"]
    cfg = replace(cfg, endpoint=host, insecure=insecure)

    return _friendly_client.SynchronousFlyteClient(cfg, **kwargs)


_PROJECT_FLAGS = ["-p", "--project"]
_DOMAIN_FLAGS = ["-d", "--domain"]
_NAME_FLAGS = ["-n", "--name"]
_VERSION_FLAGS = ["-v", "--version"]
_HOST_FLAGS = ["-h", "--host"]
_CONFIG_FLAGS = ["-c", "--config"]
_PRINCIPAL_FLAGS = ["-r", "--principal"]
_INSECURE_FLAGS = ["-i", "--insecure"]
_CERT_FLAGS = ["--cacert"]

_project_option = _click.option(*_PROJECT_FLAGS, required=True, help="The project namespace to query.")
_optional_project_option = _click.option(
    *_PROJECT_FLAGS,
    required=False,
    default=None,
    help="[Optional] The project namespace to query.",
)
_domain_option = _click.option(*_DOMAIN_FLAGS, required=True, help="The domain namespace to query.")
_optional_domain_option = _click.option(
    *_DOMAIN_FLAGS,
    required=False,
    default=None,
    help="[Optional] The domain namespace to query.",
)
_name_option = _click.option(*_NAME_FLAGS, required=True, help="The name to query.")
_optional_name_option = _click.option(
    *_NAME_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] The name to query.",
)
_principal_option = _click.option(*_PRINCIPAL_FLAGS, required=True, help="Your team name, or your name")
_optional_principal_option = _click.option(
    *_PRINCIPAL_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] Your team name, or your name",
)
_insecure_option = _click.option(*_INSECURE_FLAGS, is_flag=True, help="Do not use SSL")
_urn_option = _click.option("-u", "--urn", required=True, help="The unique identifier for an entity.")
_optional_urn_option = _click.option("-u", "--urn", required=False, help="The unique identifier for an entity.")

_host_option = _click.option(
    *_HOST_FLAGS,
    required=False,
    help="The URL for the Flyte Admin Service. If you intend for this to be consistent, set the FLYTE_PLATFORM_URL "
    "environment variable to the desired URL and this will not need to be set.",
)
_token_option = _click.option(
    "-t",
    "--token",
    required=False,
    default="",
    type=str,
    help="Pagination token from which to start listing in the list of results.",
)
_limit_option = _click.option(
    "-l",
    "--limit",
    required=False,
    default=100,
    type=int,
    help="Maximum number of results to return for this call.",
)
_show_all_option = _click.option(
    "-a",
    "--show-all",
    is_flag=True,
    default=False,
    help="Set this flag to page through and list all results.",
)
# TODO: Provide documentation on filter format
_filter_option = _click.option(
    "-f",
    "--filter",
    multiple=True,
    help="""Filter to be applied.  Multiple filters can be applied and they will be ANDed together.
    Filters may be supplied as strings such as 'eq(name, workflow_name)'. Additional documentation on filter
    syntax can be found here: https://docs.flyte.org/en/latest/dive_deep/admin_service.html#adding-request-filters""",
)
_state_choice = _click.option(
    "--state",
    type=_click.Choice(["active", "inactive"]),
    required=True,
    help="Whether or not to set schedule as active.",
)
_named_entity_state_choice = _click.option(
    "--state",
    type=_click.Choice(["active", "archived"]),
    required=True,
    help="The state change to apply to a named entity",
)
_named_entity_description_option = _click.option(
    "--description",
    required=False,
    type=str,
    help="Concise description for the entity.",
)
_sort_by_option = _click.option(
    "--sort-by",
    required=False,
    help="Provide an entity field to be sorted.  i.e. asc(name) or desc(name)",
)
_show_io_option = _click.option(
    "--show-io",
    is_flag=True,
    default=False,
    help="Set this flag to view inputs and outputs.  Pair with the --verbose flag to get the full textual description"
    " inputs and outputs.",
)
_verbose_option = _click.option(
    "--verbose",
    is_flag=True,
    default=False,
    help="Set this flag to view the full textual description of all fields.",
)

_filename_option = _click.option("-f", "--filename", required=True, help="File path of pb file")
_idl_class_option = _click.option(
    "-p",
    "--proto_class",
    required=True,
    help="Dot (.) separated path to Python IDL class. (e.g. flyteidl.core.workflow_closure_pb2.WorkflowClosure)",
)
_cause_option = _click.option(
    "-c",
    "--cause",
    required=True,
    help="The message signaling the cause of the termination of the execution(s)",
)
_optional_urns_only_option = _click.option(
    "--urns-only",
    is_flag=True,
    default=False,
    required=False,
    help="[Optional] Set the flag if you want to output the urn(s) only. Setting this will override the verbose flag",
)
_project_identifier_option = _click.option(
    "-p",
    "--identifier",
    required=True,
    type=str,
    help="Unique identifier for the project.",
)
_project_name_option = _click.option(
    "-n",
    "--name",
    required=True,
    type=str,
    help="The human-readable name for the project.",
)
_project_description_option = _click.option(
    "-d",
    "--description",
    required=True,
    type=str,
    help="Concise description for the project.",
)
_watch_option = _click.option(
    "-w",
    "--watch",
    is_flag=True,
    default=False,
    help="Set the flag if you want the command to keep watching the execution until its completion",
)

_assumable_iam_role_option = _click.option(
    "--assumable-iam-role", help="Custom assumable iam auth role to register launch plans with"
)
_kubernetes_service_acct_option = _click.option(
    "-s",
    "--kubernetes-service-account",
    help="Custom kubernetes service account auth role to register launch plans with",
)
_output_location_prefix_option = _click.option(
    "-o", "--output-location-prefix", help="Custom output location prefix for offloaded types (files/schemas)"
)
_files_argument = _click.argument(
    "files",
    type=_click.Path(exists=True),
    nargs=-1,
)


class _FlyteSubCommand(_click.Command):
    _PASSABLE_ARGS = {
        "project": _PROJECT_FLAGS[0],
        "domain": _DOMAIN_FLAGS[0],
        "name": _NAME_FLAGS[0],
        "host": _HOST_FLAGS[0],
        "config": _CONFIG_FLAGS[0],
    }

    _PASSABLE_FLAGS = {
        "insecure": _INSECURE_FLAGS[0],
    }

    def make_context(self, cmd_name, args, parent=None):
        # This list represents the set of args/flags that can be, and are, set on both the parent and subcommand.
        prefix_args = []
        for param in self.params:
            if (
                param.name in type(self)._PASSABLE_ARGS
                and param.name in parent.params
                and parent.params[param.name] is not None
            ):
                prefix_args.extend([type(self)._PASSABLE_ARGS[param.name], str(parent.params[param.name])])

            # For flags, we don't append the value of the flag, otherwise click will fail to parse
            if param.name in type(self)._PASSABLE_FLAGS and param.name in parent.params and parent.params[param.name]:
                prefix_args.append(type(self)._PASSABLE_FLAGS[param.name])

        # Pass the parameters to the subcommand "setup-config", so both of the below commands can work.
        # flyte-cli -h localhost:30081 -i setup-config
        # flyte-cli setup-config -h localhost:30081 -i
        ctx = super(_FlyteSubCommand, self).make_context(cmd_name, prefix_args + args, parent=parent)
        ctx.obj = ctx.obj or {}
        ctx.obj["cacert"] = parent.params["cacert"] or None
        if cmd_name == "setup-config":
            return ctx

        config_file = parent.params["config"]
        if config_file is None:
            # Run this as the module is loading to pick up settings that click can
            # then use when constructing the commands
            config_file = _detect_default_config_file()

        config_obj = configuration.PlatformConfig.auto(config_file=config_file)

        # These two flags are special in that they are specifiable in both the user's default ~/.flyte/config file,
        # and in the flyte-cli command itself, both in the parent-command position (flyte-cli) , and in the
        # child-command position (e.g. list-task-names).
        # For both host and insecure, command line values will override the setting in a config file.
        #
        # The host url option is a required setting, so if missing it will fail, but it may be set in the click command,
        # so we don't have to check now. It will be checked later.

        # If the host was not specified in the parent command, and also not in the subcommand, but is in the file,
        # then add in the switch and value before creating the context
        if _HOST_FLAGS[0] not in prefix_args and _HOST_FLAGS[0] not in args and config_obj.endpoint:
            prefix_args.extend([_HOST_FLAGS[0], config_obj.endpoint])

        # If insecure was not in the parent command and not in the subcommand, but is true in the config object (the
        # default is False), then add the flag to the args before creating the context.
        if _INSECURE_FLAGS[0] not in prefix_args and _INSECURE_FLAGS[0] not in args and config_obj.insecure:
            prefix_args.append(_INSECURE_FLAGS[0])

        # Create context object and fill in with additional objects
        ctx = super(_FlyteSubCommand, self).make_context(cmd_name, prefix_args + args, parent=parent)
        ctx.obj = ctx.obj or {}
        ctx.obj["cacert"] = parent.params["cacert"] or None
        ctx.obj["config"] = config_obj

        return ctx


@_click.option(
    *_CONFIG_FLAGS,
    required=False,
    type=_click.Path(exists=True),
    default=None,
    help="[Optional] The filepath to the config file to pass to the sub-command (if applicable)."
    "  If set again in the sub-command, the sub-command's parameter takes precedence.",
)
@_click.option(
    *_HOST_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] The host to pass to the sub-command (if applicable).  If set again in the sub-command, "
    "the sub-command's parameter takes precedence.",
)
@_click.option(
    *_PROJECT_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] The project to pass to the sub-command (if applicable)  If set again in the sub-command, "
    "the sub-command's parameter takes precedence.",
)
@_click.option(
    *_DOMAIN_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] The domain to pass to the sub-command (if applicable)  If set again in the sub-command, "
    "the sub-command's parameter takes precedence.",
)
@_click.option(
    *_NAME_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] The name to pass to the sub-command (if applicable)  If set again in the sub-command, "
    "the sub-command's parameter takes precedence.",
)
@_click.option(
    *_CERT_FLAGS,
    required=False,
    type=str,
    default=None,
    help="[Optional] Path to certificate file to be used to do establish SSL connection with Admin",
)
@_insecure_option
@_click.group("flyte-cli", deprecated=True)
@_click.pass_context
def _flyte_cli(ctx, host, config, project, domain, name, cacert, insecure):
    """
    Command line tool for interacting with all entities on the Flyte Platform.
    """
    if cacert and insecure:
        raise _user_exceptions.FlyteValidationException("Should not pass both certificate and insecure options!")


########################################################################################################################
#
#  Miscellaneous Commands
#
########################################################################################################################


@_flyte_cli.command("parse-proto", cls=_click.Command)
@_filename_option
@_idl_class_option
def parse_proto(filename, proto_class):
    _welcome_message()
    split = proto_class.split(".")
    idl_module = ".".join(split[:-1])
    idl_obj = split[-1]
    mod = _importlib.import_module(idl_module)
    idl = getattr(mod, idl_obj)
    obj = utils.load_proto_from_file(idl, filename)

    jsonObj = MessageToJson(obj)

    _click.echo(jsonObj)
    _click.echo("")


########################################################################################################################
#
#  Task Commands
#
########################################################################################################################


@_flyte_cli.command("list-task-names", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_sort_by_option
def list_task_names(project, domain, host, insecure, token, limit, show_all, sort_by):
    """
    List the name of the tasks that are in the registered workflow under
    a specific project and domain.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Task Names Found in {}:{}\n".format(_tt(project), _tt(domain)))
    while True:
        task_ids, next_token = client.list_task_ids_paginated(
            project,
            domain,
            limit=limit,
            token=token,
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for t in task_ids:
            _click.echo("\t{}".format(_tt(t.name)))

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("list-task-versions", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_optional_name_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_filter_option
@_sort_by_option
def list_task_versions(project, domain, name, host, insecure, token, limit, show_all, filter, sort_by):
    """
    List all the versions of the tasks under a specific {Project, Domain} tuple.
    If the name of a certain task is supplied, this command will list all the
    versions of that particular task (identifiable by {Project, Domain, Name}).
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Task Versions Found for {}:{}:{}\n".format(_tt(project), _tt(domain), _tt(name or "*")))
    _click.echo("{:50} {:40}".format("Version", "Urn"))
    while True:
        task_list, next_token = client.list_tasks_paginated(
            _common_models.NamedEntityIdentifier(project, domain, name),
            limit=limit,
            token=token,
            filters=[_filters.Filter.from_python_std(f) for f in filter],
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for t in task_list:
            _click.echo(
                "{:50} {:40}".format(
                    _tt(t.id.version),
                    _tt(cli_identifiers.Identifier.promote_from_model(t.id)),
                )
            )

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("get-task", cls=_FlyteSubCommand)
@_urn_option
@_host_option
@_insecure_option
def get_task(urn, host, insecure):
    """
    Get the details of a certain version of a task identified by the URN of it.
    The URN of the versioned task is in the form of ``tsk:<project>:<domain>:<task_name>:<version>``.
    """
    _welcome_message()
    client = _get_client(host, insecure)
    t = client.get_task(cli_identifiers.Identifier.from_python_std(urn))
    _click.echo(_tt(t))
    _click.echo("")


########################################################################################################################
#
#  Workflow Commands
#
########################################################################################################################


@_flyte_cli.command("list-workflow-names", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_sort_by_option
def list_workflow_names(project, domain, host, insecure, token, limit, show_all, sort_by):
    """
    List the names of the workflows under a scope specified by ``{project, domain}``.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Workflow Names Found in {}:{}\n".format(_tt(project), _tt(domain)))
    while True:
        wf_ids, next_token = client.list_workflow_ids_paginated(
            project,
            domain,
            limit=limit,
            token=token,
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for i in wf_ids:
            _click.echo("\t{}".format(_tt(i.name)))

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("list-workflow-versions", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_optional_name_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_filter_option
@_sort_by_option
def list_workflow_versions(project, domain, name, host, insecure, token, limit, show_all, filter, sort_by):
    """
    List all the versions of the workflows under the scope specified by ``{project, domain}``.
    If the name of a a certain workflow is supplied, this command will list all the
    versions of that particular workflow (identifiable by ``{project, domain, name}``).
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Workflow Versions Found for {}:{}:{}\n".format(_tt(project), _tt(domain), _tt(name or "*")))
    _click.echo("{:50} {:40}".format("Version", "Urn"))
    while True:
        wf_list, next_token = client.list_workflows_paginated(
            _common_models.NamedEntityIdentifier(project, domain, name),
            limit=limit,
            token=token,
            filters=[_filters.Filter.from_python_std(f) for f in filter],
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for w in wf_list:
            _click.echo(
                "{:50} {:40}".format(
                    _tt(w.id.version),
                    _tt(cli_identifiers.Identifier.promote_from_model(w.id)),
                )
            )

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("get-workflow", cls=_FlyteSubCommand)
@_urn_option
@_host_option
@_insecure_option
def get_workflow(urn, host, insecure):
    """
    Get the details of a certain version of a workflow identified by the URN in the form of
    ``wf:<project>:<domain>:<workflow_name>:<version>``
    """
    _welcome_message()
    client = _get_client(host, insecure)
    _click.echo(client.get_workflow(cli_identifiers.Identifier.from_python_std(urn)))
    # TODO: Print workflow pretty
    _click.echo("")


########################################################################################################################
#
#  Launch Plan Commands
#
########################################################################################################################


@_flyte_cli.command("list-launch-plan-names", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_sort_by_option
def list_launch_plan_names(project, domain, host, insecure, token, limit, show_all, sort_by):
    """
    List the names of the launch plans under the scope specified by {project, domain}.
    """
    _welcome_message()
    client = _get_client(host, insecure)
    _click.echo("Launch Plan Names Found in {}:{}\n".format(_tt(project), _tt(domain)))
    while True:
        wf_ids, next_token = client.list_launch_plan_ids_paginated(
            project,
            domain,
            limit=limit,
            token=token,
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for i in wf_ids:
            _click.echo("\t{}".format(_tt(i.name)))

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("list-active-launch-plans", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_sort_by_option
@_optional_urns_only_option
def list_active_launch_plans(project, domain, host, insecure, token, limit, show_all, sort_by, urns_only):
    """
    List the information of all the active launch plans under the scope specified by {project, domain}.
    An active launch plan is a launch plan with an active schedule associated with it.
    """
    if not urns_only:
        _welcome_message()
        _click.echo("Active Launch Plan Found in {}:{}\n".format(_tt(project), _tt(domain)))
        _click.echo("{:30} {:50} {:80}".format("Schedule", "Version", "Urn"))

    client = _get_client(host, insecure)
    while True:
        active_lps, next_token = client.list_active_launch_plans_paginated(
            project,
            domain,
            limit=limit,
            token=token,
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )

        for lp in active_lps:
            if urns_only:
                _click.echo("{:80}".format(_tt(cli_identifiers.Identifier.promote_from_model(lp.id))))
            else:
                _click.echo(
                    "{:30} {:50} {:80}".format(
                        _render_schedule_expr(lp),
                        _tt(lp.id.version),
                        _tt(cli_identifiers.Identifier.promote_from_model(lp.id)),
                    ),
                )

        if show_all is not True:
            if next_token and not urns_only:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token

    if not urns_only:
        _click.echo("")
    return


@_flyte_cli.command("list-launch-plan-versions", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_optional_name_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_filter_option
@_sort_by_option
@_optional_urns_only_option
def list_launch_plan_versions(
    project,
    domain,
    name,
    host,
    insecure,
    token,
    limit,
    show_all,
    filter,
    sort_by,
    urns_only,
):
    """
    List the versions of all the launch plans under the scope specified by {project, domain}.
    """
    if not urns_only:
        _welcome_message()
        _click.echo("Launch Plan Versions Found for {}:{}:{}\n".format(_tt(project), _tt(domain), _tt(name)))
        _click.echo("{:50} {:80} {:30} {:15}".format("Version", "Urn", "Schedule", "Schedule State"))

    client = _get_client(host, insecure)
    while True:
        lp_list, next_token = client.list_launch_plans_paginated(
            _common_models.NamedEntityIdentifier(project, domain, name),
            limit=limit,
            token=token,
            filters=[_filters.Filter.from_python_std(f) for f in filter],
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for l in lp_list:
            if urns_only:
                _click.echo(_tt(cli_identifiers.Identifier.promote_from_model(l.id)))
            else:
                _click.echo(
                    "{:50} {:80} ".format(
                        _tt(l.id.version),
                        _tt(cli_identifiers.Identifier.promote_from_model(l.id)),
                    ),
                    nl=False,
                )
                if l.spec.entity_metadata.schedule is not None and (
                    l.spec.entity_metadata.schedule.cron_expression or l.spec.entity_metadata.schedule.rate
                ):
                    _click.echo("{:30} ".format(_render_schedule_expr(l)), nl=False)
                    _click.secho(
                        _launch_plan.LaunchPlanState.enum_to_string(l.closure.state),
                        fg="green" if l.closure.state == _launch_plan.LaunchPlanState.ACTIVE else None,
                    )
                else:
                    _click.echo()

        if show_all is not True:
            if next_token and not urns_only:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    if not urns_only:
        _click.echo("")


@_flyte_cli.command("get-launch-plan", cls=_FlyteSubCommand)
@_urn_option
@_host_option
@_insecure_option
def get_launch_plan(urn, host, insecure):
    """
    Get the details of a certain launch plan identified by the URN of that launch plan.
    The URN of a launch plan is in the form of ``lp:<project>:<domain>:<launch_plan_name>:<version>``
    """
    _welcome_message()
    client = _get_client(host, insecure)
    _click.echo(_tt(client.get_launch_plan(cli_identifiers.Identifier.from_python_std(urn))))
    # TODO: Print launch plan pretty
    _click.echo("")


@_flyte_cli.command("get-active-launch-plan", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_name_option
@_host_option
@_insecure_option
def get_active_launch_plan(project, domain, name, host, insecure):
    """
    List the versions of all the launch plans under the scope specified by {project, domain}.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    lp = client.get_active_launch_plan(_common_models.NamedEntityIdentifier(project, domain, name))
    _click.echo("Active Launch Plan for {}:{}:{}\n".format(_tt(project), _tt(domain), _tt(name)))
    _click.echo(lp)
    _click.echo("")


@_flyte_cli.command("update-launch-plan", cls=_FlyteSubCommand)
@_state_choice
@_host_option
@_insecure_option
@_optional_urn_option
def update_launch_plan(state, host, insecure, urn=None):
    _welcome_message()
    client = _get_client(host, insecure)

    if urn is None:
        try:
            # Examine whether the input is from the named pipe
            if _stat.S_ISFIFO(_os.fstat(0).st_mode):
                for line in _sys.stdin.readlines():
                    _update_one_launch_plan(client, urn=line.rstrip(), state=state)
            else:
                # If the commandline parameter urn is not supplied, and neither
                # the input comes from a pipe, it means the user is not using
                # this command appropriately
                raise _click.UsageError('Missing option "-u" / "--urn" or missing pipe inputs')
        except KeyboardInterrupt:
            _sys.stdout.flush()
    else:
        _update_one_launch_plan(client, urn=urn, state=state)


########################################################################################################################
#
#  Execution Commands
#
########################################################################################################################


@_flyte_cli.command("recover-execution", cls=_FlyteSubCommand)
@_urn_option
@_optional_name_option
@_host_option
@_insecure_option
def recover_execution(urn, name, host, insecure):
    """
    Recreates a previously-run workflow execution that will only start executing from the last known failure point.

    In Recover mode, users cannot change any input parameters or update the version of the execution.
    This is extremely useful to recover from system errors and byzantine faults like
        - loss of K8s cluster
        - bugs in platform or instability
        - machine failures
        - downstream system failures (downstream services)
        - or simply to recover executions that failed because of retry exhaustion and should complete if tried again.

    You can optionally assign a name to the recreated execution you trigger or let the system assign one.

    Usage:
        $ flyte-cli recover-execution -u ex:flyteexamples:development:some-workflow:abc123 -n my_retry_name

    These arguments are then collected, and passed into the `lp_args` variable as a Tuple[Text].
    Users should use the get-execution and get-launch-plan commands to ascertain the names of inputs to use.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Recovering execution {}\n".format(_tt(urn)))

    original_workflow_execution_identifier = cli_identifiers.WorkflowExecutionIdentifier.from_python_std(urn)

    execution_identifier_resp = client.recover_execution(id=original_workflow_execution_identifier, name=name)
    execution_identifier = cli_identifiers.WorkflowExecutionIdentifier.promote_from_model(execution_identifier_resp)
    _click.secho("Launched execution: {}".format(execution_identifier), fg="blue")
    _click.echo("")


@_flyte_cli.command("terminate-execution", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_cause_option
@_optional_urn_option
def terminate_execution(host, insecure, cause, urn=None):
    """
    Terminate an execution or a list of executions. This command terminates an execution
    specified by the URN. It can only terminate the executions the status of which are "RUNNING".
    The post-termination status of those executions will become "ABORTED".
    When terminating an execution, the cause of termination is a required input.

    This command also supports batch terminating multiple executions at a time, which can be
    achieved by supplying multiple URNs via the named pipe.

    Note that, the termination of executions might not take immediate effect, as the
    FlyteCLI only sends a termination request to FlyteAdmin. The actual termination
    of the execution(s) depends on many other factors such as the status of the
    machine serving the execution, etc.

    e.g.,
        $ flyte-cli -h localhost:30081 -p flyteexamples -d development terminate-execution \
            -u lp:flyteexamples:development:some-execution:abc123
    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Killing the following executions:\n")
    _click.echo("{:100} {:40}".format("Urn", "Cause"))

    # It first collects the urns in a list, and then send terminate request
    # for them one-by-one
    if urn is None:
        try:
            # Examine whether the input is from FIFO (named pipe)
            if _stat.S_ISFIFO(_os.fstat(0).st_mode):
                for line in _sys.stdin.readlines():
                    _terminate_one_execution(client, line.rstrip(), cause)
            else:
                # If the commandline parameter urn is not supplied, and neither
                # the input is from a pipe, it means the user is not using
                # this command appropriately
                raise _click.UsageError('Missing option "-u" / "--urn" or missing pipe inputs.')
        except KeyboardInterrupt:
            _sys.stdout.flush()
    else:
        _terminate_one_execution(client, urn, cause)


@_flyte_cli.command("list-executions", cls=_FlyteSubCommand)
@_project_option
@_domain_option
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_filter_option
@_sort_by_option
@_optional_urns_only_option
def list_executions(project, domain, host, insecure, token, limit, show_all, filter, sort_by, urns_only):
    """
    List the key information of all the executions under the scope specified by {project, domain}.
    Users can supply additional filter arguments to show only the desired executions.

    Note that, when the ``--urns-only`` flag is not set, this command prints out the complete tabular
    output with key pieces of information such as the URN, the Name and the Status of the executions;
    the column headers are also printed. If the flag is set, on the other hand, only the URNs
    of the executions will be printed. This will come in handy when the user wants to, for example, terminate all the
    running executions at once.
    """
    if not urns_only:
        _welcome_message()
        _click.echo("Executions Found in {}:{}\n".format(_tt(project), _tt(domain)))
        _click.echo("{:100} {:40} {:10}".format("Urn", "Name", "Status"))

    client = _get_client(host, insecure)

    while True:
        exec_ids, next_token = client.list_executions_paginated(
            project,
            domain,
            limit=limit,
            token=token,
            filters=[_filters.Filter.from_python_std(f) for f in filter],
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for ex in exec_ids:
            _secho_one_execution(ex, urns_only)

        if show_all is not True:
            if next_token and not urns_only:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    if not urns_only:
        _click.echo("")


def _get_io(node_executions, wf_execution, show_io, verbose):
    # Fetch I/O if necessary
    uri_to_message_map = {}
    if show_io:
        uris = [ne.input_uri for ne in node_executions]
        uris.extend([ne.closure.output_uri for ne in node_executions if ne.closure.output_uri is not None])
        if (
            wf_execution is not None
            and wf_execution.closure.outputs is not None
            and wf_execution.closure.outputs.uri is not None
        ):
            uris.append(wf_execution.closure.outputs.uri)

        with _click.progressbar(uris, label="Downloading Inputs and Outputs") as progress_bar_uris:
            for uri in progress_bar_uris:
                uri_to_message_map[uri] = _fetch_and_stringify_literal_map(uri, verbose=verbose)
    return uri_to_message_map


def _render_workflow_execution(wf_execution, uri_to_message_map, show_io, verbose):
    _click.echo(
        "\nExecution {project}:{domain}:{name}\n".format(
            project=_tt(wf_execution.id.project),
            domain=_tt(wf_execution.id.domain),
            name=_tt(wf_execution.id.name),
        )
    )
    _click.echo("\t{:15} ".format("State:"), nl=False)
    _secho_workflow_status(wf_execution.closure.phase)
    _click.echo(
        "\t{:15} {}".format(
            "Launch Plan:",
            _tt(cli_identifiers.Identifier.promote_from_model(wf_execution.spec.launch_plan)),
        )
    )

    if show_io:
        _click.secho(
            "\tInputs: {}\n".format(
                _prefix_lines(
                    "\t\t",
                    _get_io_string(wf_execution.closure.computed_inputs, verbose=verbose),
                )
            )
        )
        if wf_execution.closure.outputs is not None:
            if wf_execution.closure.outputs.uri:
                _click.secho(
                    "\tOutputs: {}\n".format(
                        _prefix_lines(
                            "\t\t",
                            uri_to_message_map.get(
                                wf_execution.closure.outputs.uri,
                                wf_execution.closure.outputs.uri,
                            ),
                        )
                    )
                )
            elif wf_execution.closure.outputs.values is not None:
                _click.secho(
                    "\tOutputs: {}\n".format(
                        _prefix_lines(
                            "\t\t",
                            _get_io_string(wf_execution.closure.outputs.values, verbose=verbose),
                        )
                    )
                )
            else:
                _click.echo("\t{:15} (None)".format("Outputs:"))

    if wf_execution.closure.error is not None:
        _click.secho(
            _prefix_lines("\t", _render_error(wf_execution.closure.error)),
            fg="red",
            bold=True,
        )


def _render_error(error):
    out = "Error:\n"
    out += "\tCode: {}\n".format(error.code)
    out += "\tMessage:\n"
    for l in error.message.splitlines():
        out += "\t\t{}".format(_tt(l))
    return out


def _get_all_task_executions_for_node(client, node_execution_identifier):
    fetched_task_execs = []
    token = ""
    while True:
        num_to_fetch = 100
        task_execs, next_token = client.list_task_executions_paginated(
            node_execution_identifier=node_execution_identifier,
            limit=num_to_fetch,
            token=token,
        )
        for te in task_execs:
            fetched_task_execs.append(te)

        if not next_token:
            break
        token = next_token

    return fetched_task_execs


def _get_all_node_executions(client, workflow_execution_identifier=None, task_execution_identifier=None):
    all_node_execs = []
    token = ""
    while True:
        num_to_fetch = 100
        if workflow_execution_identifier:
            node_execs, next_token = client.list_node_executions(
                workflow_execution_identifier=workflow_execution_identifier,
                limit=num_to_fetch,
                token=token,
            )
        else:
            node_execs, next_token = client.list_node_executions_for_task_paginated(
                task_execution_identifier=task_execution_identifier,
                limit=num_to_fetch,
                token=token,
            )
        all_node_execs.extend(node_execs)
        if not next_token:
            break
        token = next_token
    return all_node_execs


def _render_node_executions(client, node_execs, show_io, verbose, host, insecure, wf_execution=None):
    node_executions_to_task_executions = {}
    for node_exec in node_execs:
        node_executions_to_task_executions[node_exec.id] = _get_all_task_executions_for_node(client, node_exec.id)

    uri_to_message_map = _get_io(node_execs, wf_execution, show_io, verbose)
    if wf_execution is not None:
        _render_workflow_execution(wf_execution, uri_to_message_map, show_io, verbose)

    _click.echo("\n\tNode Executions:\n")
    for ne in sorted(node_execs, key=lambda x: x.closure.started_at):
        if ne.id.node_id == "start-node":
            continue
        _click.echo("\t\tID: {}\n".format(_tt(ne.id.node_id)))
        _click.echo("\t\t\t{:15} ".format("Status:"), nl=False)
        _secho_node_execution_status(ne.closure.phase)
        _click.echo("\t\t\t{:15} {:60} ".format("Started:", _tt(ne.closure.started_at)))
        _click.echo("\t\t\t{:15} {:60} ".format("Duration:", _tt(ne.closure.duration)))
        _click.echo(
            "\t\t\t{:15} {}".format(
                "Input:",
                _prefix_lines(
                    "\t\t\t{:15} ".format(""),
                    uri_to_message_map.get(ne.input_uri, ne.input_uri),
                ),
            )
        )
        if ne.closure.output_uri:
            _click.echo(
                "\t\t\t{:15} {}".format(
                    "Output:",
                    _prefix_lines(
                        "\t\t\t{:15} ".format(""),
                        uri_to_message_map.get(ne.closure.output_uri, ne.closure.output_uri),
                    ),
                )
            )
        if ne.closure.error is not None:
            _click.secho(
                _prefix_lines("\t\t\t", _render_error(ne.closure.error)),
                bold=True,
                fg="red",
            )

        task_executions = node_executions_to_task_executions.get(ne.id, [])
        if len(task_executions) > 0:
            _click.echo("\n\t\t\tTask Executions:\n")
            for te in sorted(task_executions, key=lambda x: x.id.retry_attempt):
                _click.echo("\t\t\t\tAttempt {}:\n".format(te.id.retry_attempt))
                _click.echo("\t\t\t\t\t{:15} {:60} ".format("Created:", _tt(te.closure.created_at)))
                _click.echo("\t\t\t\t\t{:15} {:60} ".format("Started:", _tt(te.closure.started_at)))
                _click.echo("\t\t\t\t\t{:15} {:60} ".format("Updated:", _tt(te.closure.updated_at)))
                _click.echo("\t\t\t\t\t{:15} {:60} ".format("Duration:", _tt(te.closure.duration)))
                _click.echo("\t\t\t\t\t{:15} ".format("Status:"), nl=False)
                _secho_task_execution_status(te.closure.phase)
                if len(te.closure.logs) == 0:
                    _click.echo("\t\t\t\t\t{:15} {:60} ".format("Logs:", "(None Found Yet)"))
                else:
                    _click.echo("\t\t\t\t\tLogs:\n")
                    for log in sorted(te.closure.logs, key=lambda x: x.name):
                        _click.echo("\t\t\t\t\t\t{:8} {}".format("Name:", log.name))
                        _click.echo("\t\t\t\t\t\t{:8} {}\n".format("URI:", log.uri))

                if te.closure.error is not None:
                    _click.secho(
                        _prefix_lines("\t\t\t\t\t", _render_error(te.closure.error)),
                        bold=True,
                        fg="red",
                    )

                if te.is_parent:
                    _click.echo(
                        "\t\t\t\t\t{:15} {:60} ".format(
                            "Subtasks:",
                            "flyte-cli get-child-executions -h {host}{insecure} -u {urn}".format(
                                host=host,
                                urn=_tt(cli_identifiers.TaskExecutionIdentifier.promote_from_model(te.id)),
                                insecure=" --insecure" if insecure else "",
                            ),
                        )
                    )
            _click.echo()
        _click.echo()
    _click.echo()


@_flyte_cli.command("get-execution", cls=_FlyteSubCommand)
@_urn_option
@_host_option
@_insecure_option
@_show_io_option
@_verbose_option
def get_execution(urn, host, insecure, show_io, verbose):
    """
    Get the detail information of a certain execution identified by the URN of that launch plan.
    The URN of an execution is in the form of ``ex:<project>:<domain>:<execution_name>``
    """
    _welcome_message()
    client = _get_client(host, insecure)
    e = client.get_execution(cli_identifiers.WorkflowExecutionIdentifier.from_python_std(urn))
    node_execs = _get_all_node_executions(client, workflow_execution_identifier=e.id)
    _render_node_executions(client, node_execs, show_io, verbose, host, insecure, wf_execution=e)


@_flyte_cli.command("get-child-executions", cls=_FlyteSubCommand)
@_urn_option
@_host_option
@_insecure_option
@_show_io_option
@_verbose_option
def get_child_executions(urn, host, insecure, show_io, verbose):
    _welcome_message()
    client = _get_client(host, insecure)
    node_execs = _get_all_node_executions(
        client,
        task_execution_identifier=cli_identifiers.TaskExecutionIdentifier.from_python_std(urn),
    )
    _render_node_executions(client, node_execs, show_io, verbose, host, insecure)


@_flyte_cli.command("register-project", cls=_FlyteSubCommand)
@_project_identifier_option
@_project_name_option
@_project_description_option
@_host_option
@_insecure_option
def register_project(identifier, name, description, host, insecure):
    """
    Register a new project.

    """
    _welcome_message()
    client = _get_client(host, insecure)
    client.register_project(_Project(identifier, name, description))
    _click.echo("Registered project [id: {}, name: {}, description: {}]".format(identifier, name, description))


@_flyte_cli.command("list-projects", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_token_option
@_limit_option
@_show_all_option
@_filter_option
@_sort_by_option
def list_projects(host, insecure, token, limit, show_all, filter, sort_by):
    """
    List projects.

    """
    _welcome_message()
    client = _get_client(host, insecure)

    _click.echo("Projects Found\n")
    while True:
        projects, next_token = client.list_projects_paginated(
            limit=limit,
            token=token,
            filters=[_filters.Filter.from_python_std(f) for f in filter],
            sort_by=_admin_common.Sort.from_python_std(sort_by) if sort_by else None,
        )
        for p in projects:
            _click.echo("\t{}".format(_tt(p.id)))

        if show_all is not True:
            if next_token:
                _click.echo("Received next token: {}\n".format(next_token))
            break
        if not next_token:
            break
        token = next_token
    _click.echo("")


@_flyte_cli.command("archive-project", cls=_FlyteSubCommand)
@_project_identifier_option
@_host_option
@_insecure_option
def archive_project(identifier, host, insecure):
    """
    Archive a project.

    """
    _welcome_message()
    client = _get_client(host, insecure)

    client.update_project(_Project.archived_project(identifier))
    _click.echo("Archived project [id: {}]".format(identifier))


@_flyte_cli.command("activate-project", cls=_FlyteSubCommand)
@_project_identifier_option
@_host_option
@_insecure_option
def activate_project(identifier, host, insecure):
    """
    Activate a project.

    """
    _welcome_message()
    client = _get_client(host, insecure)
    client.update_project(_Project.active_project(identifier))
    _click.echo("Activated project [id: {}]".format(identifier))


_resource_map = {
    _identifier_pb2.LAUNCH_PLAN: _launch_plan_pb2.LaunchPlan,
    _identifier_pb2.WORKFLOW: _workflow_pb2.WorkflowSpec,
    _identifier_pb2.TASK: _task_pb2.TaskSpec,
}


def _extract_pair(
    object_file: str,
    resource_type: int,
    project: str,
    domain: str,
    version: str,
    patches: Dict[int, Callable[[_GeneratedProtocolMessageType], _GeneratedProtocolMessageType]],
) -> Tuple[
    _identifier_pb2.Identifier,
    Union[_core_tasks_pb2.TaskTemplate, _core_workflow_pb2.WorkflowTemplate, _launch_plan_pb2.LaunchPlanSpec],
]:
    """
    :param Text identifier_file:
    :param Text object_file:
    :rtype: (flyteidl.core.identifier_pb2.Identifier, T)
    """
    if resource_type not in _resource_map:
        raise _user_exceptions.FlyteAssertion(
            f"Resource type found in proto file name [{resource_type}] invalid, "
            "must be 1 (task), 2 (workflow) or 3 (launch plan)"
        )
    entity = utils.load_proto_from_file(_resource_map[resource_type], object_file)
    registerable_identifier, registerable_entity = hydrate_registration_parameters(
        resource_type, project, domain, version, entity
    )
    patch_fn = patches.get(resource_type)
    if patch_fn:
        registerable_entity = patch_fn(registerable_entity)
    return registerable_identifier, registerable_entity


def _extract_files(
    project: str,
    domain: str,
    version: str,
    file_paths: List[str],
    patches: Dict[int, Callable[[_GeneratedProtocolMessageType], _GeneratedProtocolMessageType]] = None,
):
    """
    :param file_paths:
    :rtype: List[(flyteidl.core.identifier_pb2.Identifier, T)]
    """
    # Get a manual iterator because we're going to grab files two at a time.
    # The identifier file will always come first because the names are always the same and .identifier.pb sorts before
    # .pb

    results = []
    for proto_file in file_paths:
        # Serialized proto files are of the form: 12_foo.bar_1.pb
        # Where 12 indicates it is the 12 file to process in order and 1 that is of resource type 1, or TASK.
        resource_type = int(proto_file[-4])
        id, entity = _extract_pair(proto_file, resource_type, project, domain, version, patches or {})
        results.append((id, entity))

    return results


def _get_patch_launch_plan_fn(
    assumable_iam_role: str = None, kubernetes_service_account: str = None, output_location_prefix: str = None
) -> Callable[[_GeneratedProtocolMessageType], _GeneratedProtocolMessageType]:
    def patch_launch_plan(entity: _GeneratedProtocolMessageType) -> _GeneratedProtocolMessageType:
        """
        Updates launch plans during registration to add a customizable auth role that overrides any values set in
        the flyte config and/or a custom output_location_prefix.
        """
        # entity is of type flyteidl.admin.launch_plan_pb2.LaunchPlanSpec
        entity.spec.auth_role.CopyFrom(
            _AuthRole(
                assumable_iam_role=assumable_iam_role,
                kubernetes_service_account=kubernetes_service_account,
            ).to_flyte_idl(),
        )

        if output_location_prefix is not None:
            entity.spec.raw_output_data_config.CopyFrom(
                _RawOutputDataConfig(output_location_prefix=output_location_prefix).to_flyte_idl()
            )

        _click.echo(
            f"IAM_Role: {assumable_iam_role}, ServiceAccount: {kubernetes_service_account},"
            f" OutputLocationPrefix: {output_location_prefix}"
        )

        return entity

    return patch_launch_plan


def _extract_and_register(
    client: _friendly_client.SynchronousFlyteClient,
    project: str,
    domain: str,
    version: str,
    file_paths: List[str],
    patches: Dict[int, Callable[[_GeneratedProtocolMessageType], _GeneratedProtocolMessageType]] = None,
):
    flyte_entities_list = _extract_files(project, domain, version, file_paths, patches)
    for id, flyte_entity in flyte_entities_list:
        _click.secho(f"Registering {id}", fg="yellow")
        try:
            if id.resource_type == _identifier_pb2.LAUNCH_PLAN:
                client.raw.create_launch_plan(_launch_plan_pb2.LaunchPlanCreateRequest(id=id, spec=flyte_entity.spec))
            elif id.resource_type == _identifier_pb2.TASK:
                client.raw.create_task(_task_pb2.TaskCreateRequest(id=id, spec=flyte_entity))
            elif id.resource_type == _identifier_pb2.WORKFLOW:
                client.raw.create_workflow(_workflow_pb2.WorkflowCreateRequest(id=id, spec=flyte_entity))
            else:
                raise _user_exceptions.FlyteAssertion(
                    f"Only tasks, launch plans, and workflows can be called with this function, "
                    f"resource type {id.resource_type} was passed"
                )
        except _user_exceptions.FlyteEntityAlreadyExistsException:
            _click.secho(f"Skipping because already registered {id}", fg="cyan")

    _click.echo(f"Finished scanning {len(flyte_entities_list)} files")


@_flyte_cli.command("register-files", cls=_FlyteSubCommand)
@_click.option(*_PROJECT_FLAGS, required=True, help="The project namespace to register with.")
@_click.option(*_DOMAIN_FLAGS, required=True, help="The domain namespace to register with.")
@_click.option(*_VERSION_FLAGS, required=True, help="The entity version to register with")
@_host_option
@_insecure_option
@_assumable_iam_role_option
@_kubernetes_service_acct_option
@_output_location_prefix_option
@_files_argument
def register_files(
    project,
    domain,
    version,
    host,
    insecure,
    assumable_iam_role,
    kubernetes_service_account,
    output_location_prefix,
    files,
):
    """
    Given a list of files, this will (after sorting the input list), attempt to register them against Flyte Admin.
    This command expects the files to be the output of the pyflyte serialize command.  See the code there for more
    information. Valid files need to be:\n
        * Ordered in the order that you want registration to happen. pyflyte should have done the topological sort
          for you and produced file that have a prefix that sets the correct order.\n
        * Of the correct type. That is, they should be the serialized form of one of these Flyte IDL objects
          (or an identifier object).\n
          - flyteidl.admin.launch_plan_pb2.LaunchPlan for launch plans\n
          - flyteidl.admin.workflow_pb2.WorkflowSpec for workflows\n
          - flyteidl.admin.task_pb2.TaskSpec for tasks\n

    :param host:
    :param insecure:
    :param files:
    :return:
    """
    _welcome_message()
    files = list(files)
    files.sort()
    _click.secho("Parsing files...", fg="green", bold=True)
    for f in files:
        _click.echo(f"  {f}")

    patches = {
        _identifier_pb2.LAUNCH_PLAN: _get_patch_launch_plan_fn(
            assumable_iam_role, kubernetes_service_account, output_location_prefix
        )
    }
    client = _get_client(host, insecure)
    _extract_and_register(client, project, domain, version, files, patches)


def _substitute_fast_register_task_args(args: List[str], full_remote_path: str, dest_dir: str) -> List[str]:
    complete_args = []
    for arg in args:
        if arg == "{{ .remote_package_path }}":
            arg = full_remote_path
        elif arg == "{{ .dest_dir }}":
            arg = dest_dir if dest_dir else "."
        complete_args.append(arg)
    return complete_args


@_flyte_cli.command("fast-register-files", cls=_FlyteSubCommand)
@_click.option(*_PROJECT_FLAGS, required=True, help="The project namespace to register with.")
@_click.option(*_DOMAIN_FLAGS, required=True, help="The domain namespace to register with.")
@_click.option(
    *_VERSION_FLAGS,
    required=False,
    help="Version to register entities with. This is normally computed deterministically from your code, but you can "
    "override that here",
)
@_host_option
@_insecure_option
@_click.option("--additional-distribution-dir", required=True, help="Location for additional distributions")
@_click.option(
    "--dest-dir",
    type=str,
    help="[Optional] The output directory for code which is downloaded during fast registration, "
    "if the current working directory at the time of installation is not desired",
)
@_assumable_iam_role_option
@_kubernetes_service_acct_option
@_output_location_prefix_option
@_files_argument
def fast_register_files(
    project,
    domain,
    version,
    host,
    insecure,
    additional_distribution_dir,
    dest_dir,
    assumable_iam_role,
    kubernetes_service_account,
    output_location_prefix,
    files,
):
    """
    Given a list of files, this will (after sorting the input list), attempt to register them against Flyte Admin.
    This command expects the files to be the output of the pyflyte serialize command.  See the code there for more
    information. Valid files need to be:\n
        * Ordered in the order that you want registration to happen. pyflyte should have done the topological sort
          for you and produced file that have a prefix that sets the correct order.\n
        * Of the correct type. That is, they should be the serialized form of one of these Flyte IDL objects
          (or an identifier object).\n
          - flyteidl.admin.launch_plan_pb2.LaunchPlanSpec for launch plans\n
          - flyteidl.admin.workflow_pb2.WorkflowSpec for workflows\n
          - flyteidl.admin.task_pb2.TaskSpec for tasks\n

    :param host:
    :param insecure:
    :param files:
    :return:
    """
    _welcome_message()
    files = list(files)
    files.sort()
    _click.secho("Parsing files...", fg="green", bold=True)
    compressed_source, digest = None, None
    pb_files = []
    for f in files:
        if f.endswith("tar.gz"):
            compressed_source = f
            digest = os.path.basename(f).split(".")[0]
        else:
            _click.echo(f"  {f}")
            pb_files.append(f)

    if compressed_source is None:
        raise _click.UsageError(
            "Could not discover compressed source, did you remember to run `pyflyte serialize fast ...`?"
        )

    version = version if version else digest
    full_remote_path = _get_additional_distribution_loc(additional_distribution_dir, version)
    ctx = FlyteContextManager.current_context()
    full_remote_path = ctx.file_access.put_data(compressed_source, full_remote_path)
    _click.secho(f"Uploaded compressed code archive {compressed_source} to {full_remote_path}", fg="green")

    def fast_register_task(entity: _GeneratedProtocolMessageType) -> _GeneratedProtocolMessageType:
        """
        Updates task definitions during fast-registration in order to use the compatible pyflyte fast execute command at
        task execution.
        """
        # entity is of type flyteidl.admin.task_pb2.TaskSpec

        if entity.template.HasField("container") and len(entity.template.container.args) > 0:
            complete_args = _substitute_fast_register_task_args(
                entity.template.container.args, full_remote_path, dest_dir
            )
            # Because we're dealing with a proto list, we have to delete the existing args before we can extend the list
            # with the substituted ones.
            del entity.template.container.args[:]
            entity.template.container.args.extend(complete_args)

        if entity.template.HasField("k8s_pod"):
            pod_spec_struct = entity.template.k8s_pod.pod_spec
            if "containers" in pod_spec_struct:
                for idx in range(len(pod_spec_struct["containers"])):
                    if "args" in pod_spec_struct["containers"][idx]:
                        # We can directly overwrite the args in the pod spec struct definition.
                        pod_spec_struct["containers"][idx]["args"] = _substitute_fast_register_task_args(
                            pod_spec_struct["containers"][idx]["args"], full_remote_path, dest_dir
                        )
        return entity

    patches = {
        _identifier_pb2.TASK: fast_register_task,
        _identifier_pb2.LAUNCH_PLAN: _get_patch_launch_plan_fn(
            assumable_iam_role, kubernetes_service_account, output_location_prefix
        ),
    }
    client = _get_client(host, insecure)

    _extract_and_register(client, project, domain, version, pb_files, patches)


@_flyte_cli.command("update-workflow-meta", cls=_FlyteSubCommand)
@_named_entity_description_option
@_named_entity_state_choice
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
def update_workflow_meta(description, state, host, insecure, project, domain, name):
    """
    Updates a workflow entity under the scope specified by {project, domain, name} across versions.
    """
    _welcome_message()
    client = _get_client(host, insecure)
    if state == "active":
        state = _named_entity.NamedEntityState.ACTIVE
    elif state == "archived":
        state = _named_entity.NamedEntityState.ARCHIVED
    client.update_named_entity(
        _core_identifier.ResourceType.WORKFLOW,
        _named_entity.NamedEntityIdentifier(project, domain, name),
        _named_entity.NamedEntityMetadata(description, state),
    )
    _click.echo("Successfully updated workflow")


@_flyte_cli.command("update-task-meta", cls=_FlyteSubCommand)
@_named_entity_description_option
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
def update_task_meta(description, host, insecure, project, domain, name):
    """
    Updates a task entity under the scope specified by {project, domain, name} across versions.
    """
    _welcome_message()
    client = _get_client(host, insecure)
    client.update_named_entity(
        _core_identifier.ResourceType.TASK,
        _named_entity.NamedEntityIdentifier(project, domain, name),
        _named_entity.NamedEntityMetadata(description, _named_entity.NamedEntityState.ACTIVE),
    )
    _click.echo("Successfully updated task")


@_flyte_cli.command("update-launch-plan-meta", cls=_FlyteSubCommand)
@_named_entity_description_option
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
def update_launch_plan_meta(description, host, insecure, project, domain, name):
    """
    Updates a launch plan entity under the scope specified by {project, domain, name} across versions.
    """
    _welcome_message()
    client = _get_client(host, insecure)
    client.update_named_entity(
        _core_identifier.ResourceType.LAUNCH_PLAN,
        _named_entity.NamedEntityIdentifier(project, domain, name),
        _named_entity.NamedEntityMetadata(description, _named_entity.NamedEntityState.ACTIVE),
    )
    _click.echo("Successfully updated launch plan")


@_flyte_cli.command("update-cluster-resource-attributes", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
@_click.option("--attributes", type=(str, str), multiple=True)
def update_cluster_resource_attributes(host, insecure, project, domain, name, attributes):
    """
    Sets matchable cluster resource attributes for a project, domain and optionally, workflow name.
    The attribute names should match the templatized values you use to configure these resource
    attributes in your flyteadmin deployment. See
    https://lyft.github.io/flyte/administrator/install/managing_customizable_resources.html#cluster-resources
    for more documentation.

    e.g.
        $ flyte-cli -h localhost:30081 -p flyteexamples -d development update-cluster-resource-attributes \
            --attributes projectQuotaCpu 1 --attributes projectQuotaMemory 500M
    """
    _welcome_message()
    client = _get_client(host, insecure)
    cluster_resource_attributes = _ClusterResourceAttributes({attribute[0]: attribute[1] for attribute in attributes})
    matching_attributes = _MatchingAttributes(cluster_resource_attributes=cluster_resource_attributes)

    if name is not None:
        client.update_workflow_attributes(project, domain, name, matching_attributes)
        _click.echo(
            "Successfully updated cluster resource attributes for project: {}, domain: {}, and workflow: {}".format(
                project, domain, name
            )
        )
    else:
        client.update_project_domain_attributes(project, domain, matching_attributes)
        _click.echo(
            "Successfully updated cluster resource attributes for project: {} and domain: {}".format(project, domain)
        )


@_flyte_cli.command("update-execution-queue-attributes", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
@_click.option("--tags", multiple=True, help="Tag(s) to be applied.")
def update_execution_queue_attributes(host, insecure, project, domain, name, tags):
    """
    Tags used for assigning execution queues for tasks belonging to a project, domain and optionally, workflow name.

    e.g.
        $ flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-queue-attributes \
            --tags critical --tags gpu_intensive
    """
    _welcome_message()
    client = _get_client(host, insecure)
    execution_queue_attributes = _ExecutionQueueAttributes(list(tags))
    matching_attributes = _MatchingAttributes(execution_queue_attributes=execution_queue_attributes)

    if name is not None:
        client.update_workflow_attributes(project, domain, name, matching_attributes)
        _click.echo(
            "Successfully updated execution queue attributes for project: {}, domain: {}, and workflow: {}".format(
                project, domain, name
            )
        )
    else:
        client.update_project_domain_attributes(project, domain, matching_attributes)
        _click.echo(
            "Successfully updated execution queue attributes for project: {} and domain: {}".format(project, domain)
        )


@_flyte_cli.command("update-execution-cluster-label", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
@_click.option("--value", help="Cluster label for which to schedule matching executions")
def update_execution_cluster_label(host, insecure, project, domain, name, value):
    """
    Label value to determine where an execution's task will be run for tasks belonging to a project, domain and
        optionally, workflow name.

    e.g.
        $ flyte-cli -h localhost:30081 -p flyteexamples -d development update-execution-cluster-label --value foo
    """
    _welcome_message()
    client = _get_client(host, insecure)
    execution_cluster_label = _ExecutionClusterLabel(value)
    matching_attributes = _MatchingAttributes(execution_cluster_label=execution_cluster_label)

    if name is not None:
        client.update_workflow_attributes(project, domain, name, matching_attributes)
        _click.echo(
            "Successfully updated execution cluster label for project: {}, domain: {}, and workflow: {}".format(
                project, domain, name
            )
        )
    else:
        client.update_project_domain_attributes(project, domain, matching_attributes)
        _click.echo(
            "Successfully updated execution cluster label for project: {} and domain: {}".format(project, domain)
        )


@_flyte_cli.command("update-plugin-override", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
@_click.option("--task-type", help="Task type for which to apply plugin implementation overrides")
@_click.option("--plugin-id", multiple=True, help="Plugin id(s) to be used in place of the default for the task type.")
@_click.option(
    "--missing-plugin-behavior", help="Behavior when no specified plugin_id has an associated handler.", default="FAIL"
)
def update_plugin_override(host, insecure, project, domain, name, task_type, plugin_id, missing_plugin_behavior):
    """
    Plugin ids designating non-default plugin handlers to be used for tasks of a certain type.

    e.g.
        $ flyte-cli -h localhost:30081 -p flyteexamples -d development update-plugin-override --task-type python \
            --plugin-id my_cool_plugin --plugin-id my_fallback_plugin --missing-plugin-behavior FAIL
    """
    _welcome_message()
    client = _get_client(host, insecure)
    plugin_override = _PluginOverride(
        task_type, list(plugin_id), _PluginOverride.string_to_enum(missing_plugin_behavior.upper())
    )
    matching_attributes = _MatchingAttributes(plugin_overrides=_PluginOverrides(overrides=[plugin_override]))

    if name is not None:
        client.update_workflow_attributes(project, domain, name, matching_attributes)
        _click.echo(
            "Successfully updated plugin override for project: {}, domain: {}, and workflow: {}".format(
                project, domain, name
            )
        )
    else:
        client.update_project_domain_attributes(project, domain, matching_attributes)
        _click.echo("Successfully updated plugin override for project: {} and domain: {}".format(project, domain))


@_flyte_cli.command("get-matching-attributes", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_project_option
@_domain_option
@_optional_name_option
@_click.option(
    "--resource-type",
    help="Resource type",
    required=True,
    type=_click.Choice(
        [
            "task_resource",
            "cluster_resource",
            "execution_queue",
            "execution_cluster_label",
            "quality_of_service_specification",
        ]
    ),
)
def get_matching_attributes(host, insecure, project, domain, name, resource_type):
    """
    Fetches the matchable resource of the given resource type for this project, domain and optionally workflow name
    combination.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    if name is not None:
        attributes = client.get_workflow_attributes(
            project, domain, name, _MatchableResource.string_to_enum(resource_type.upper())
        )
        _click.echo("{}".format(attributes))
    else:
        attributes = client.get_project_domain_attributes(
            project, domain, _MatchableResource.string_to_enum(resource_type.upper())
        )
        _click.echo("{}".format(attributes))


@_flyte_cli.command("list-matching-attributes", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
@_click.option(
    "--resource-type",
    help="Resource type",
    required=True,
    type=_click.Choice(
        [
            "task_resource",
            "cluster_resource",
            "execution_queue",
            "execution_cluster_label",
            "quality_of_service_specification",
        ]
    ),
)
def list_matching_attributes(host, insecure, resource_type):
    """
    Fetches all matchable resources of the given resource type.
    """
    _welcome_message()
    client = _get_client(host, insecure)

    attributes = client.list_matchable_attributes(_MatchableResource.string_to_enum(resource_type.upper()))
    for cfg in attributes.configurations:
        _click.secho(
            "{:20} {:20} {:20} {:20}\n".format(
                _tt(cfg.project),
                _tt(cfg.domain),
                _tt(cfg.workflow),
                _tt(cfg.launch_plan),
            ),
            fg="blue",
            nl=False,
        )
        _click.echo("{}".format(cfg.attributes))


@_flyte_cli.command("setup-config", cls=_FlyteSubCommand)
@_host_option
@_insecure_option
def setup_config(host, insecure):
    """
    Set-up a default config file.

    """
    _welcome_message()
    config_file = _get_config_file_path()
    if _get_user_filepath_home() and _os.path.exists(config_file):
        _click.secho("Config file already exists at {}".format(_tt(config_file)), fg="blue")
        return

    # Before creating check that the directory exists and create if not
    config_dir = _os.path.join(_get_user_filepath_home(), _default_config_file_dir)
    if not _os.path.isdir(config_dir):
        _click.secho(
            "Creating default Flyte configuration directory at {}".format(_tt(config_dir)),
            fg="blue",
        )
        _os.mkdir(config_dir)

    full_host = "http://{}".format(host) if insecure else "https://{}".format(host)
    config_url = _urlparse.urljoin(full_host, "config/v1/flyte_client")
    response = _requests.get(config_url)
    data = response.json()
    platform_config = {"url": str(host), "insecure": str(insecure)}
    credentials_config = None
    if not insecure:
        # We can't get credentials config in insecure mode
        credentials_config = {
            "client_id": data["client_id"],
            "redirect_uri": data["redirect_uri"],
            "scopes": data["scopes"],
            "authorization_metadata_key": data["authorization_metadata_key"],
            "auth_mode": "standard",
        }
    with open(config_file, "w+") as f:
        parser = _configparser.ConfigParser()
        parser.add_section("platform")
        for key in platform_config.keys():
            parser.set("platform", key, platform_config[key])
        if not insecure:
            parser.add_section("credentials")
            for key in credentials_config.keys():
                # ConfigParser needs all keys to be strings
                parser.set("credentials", key, str(credentials_config[key]))
        parser.write(f)
    _click.secho("Wrote default config file to {}".format(_tt(config_file)), fg="blue")


if __name__ == "__main__":
    _flyte_cli()
