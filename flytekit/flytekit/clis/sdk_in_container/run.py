import asyncio
import importlib
import inspect
import json
import os
import pathlib
import tempfile
import typing
from dataclasses import dataclass, field, fields
from typing import cast, get_args

import rich_click as click
from dataclasses_json import DataClassJsonMixin
from rich.progress import Progress

from flytekit import Annotations, FlyteContext, FlyteContextManager, Labels, Literal
from flytekit.clis.sdk_in_container.helpers import patch_image_config
from flytekit.clis.sdk_in_container.utils import (
    PyFlyteParams,
    domain_option,
    get_option_from_metadata,
    make_click_option_field,
    pretty_print_exception,
    project_option,
)
from flytekit.configuration import DefaultImages, FastSerializationSettings, ImageConfig, SerializationSettings
from flytekit.configuration.plugin import get_plugin
from flytekit.core import context_manager
from flytekit.core.base_task import PythonTask
from flytekit.core.data_persistence import FileAccessProvider
from flytekit.core.type_engine import TypeEngine
from flytekit.core.workflow import PythonFunctionWorkflow, WorkflowBase
from flytekit.exceptions.system import FlyteSystemException
from flytekit.interaction.click_types import FlyteLiteralConverter, key_value_callback
from flytekit.interaction.string_literals import literal_string_repr
from flytekit.loggers import logger
from flytekit.models import security
from flytekit.models.common import RawOutputDataConfig
from flytekit.models.interface import Parameter, Variable
from flytekit.models.types import SimpleType
from flytekit.remote import FlyteLaunchPlan, FlyteRemote, FlyteTask, FlyteWorkflow, remote_fs
from flytekit.remote.executions import FlyteWorkflowExecution
from flytekit.tools import module_loader
from flytekit.tools.script_mode import _find_project_root, compress_scripts
from flytekit.tools.translator import Options


@dataclass
class RunLevelComputedParams:
    """
    This class is used to store the computed parameters that are used to run a workflow / task / launchplan.
    Computed parameters are created during the execution
    """

    project_root: typing.Optional[str] = None
    module: typing.Optional[str] = None
    temp_file_name: typing.Optional[str] = None  # Used to store the temporary location of the file downloaded


@dataclass
class RunLevelParams(PyFlyteParams):
    """
    This class is used to store the parameters that are used to run a workflow / task / launchplan.
    """

    project: str = make_click_option_field(project_option)
    domain: str = make_click_option_field(domain_option)
    destination_dir: str = make_click_option_field(
        click.Option(
            param_decls=["--destination-dir", "destination_dir"],
            required=False,
            type=str,
            default=".",
            show_default=True,
            help="Directory inside the image where the tar file containing the code will be copied to",
        )
    )
    copy_all: bool = make_click_option_field(
        click.Option(
            param_decls=["--copy-all", "copy_all"],
            required=False,
            is_flag=True,
            default=False,
            show_default=True,
            help="Copy all files in the source root directory to the destination directory",
        )
    )
    image_config: ImageConfig = make_click_option_field(
        click.Option(
            param_decls=["-i", "--image", "image_config"],
            required=False,
            multiple=True,
            type=click.UNPROCESSED,
            callback=ImageConfig.validate_image,
            default=[DefaultImages.default_image()],
            show_default=True,
            help="Image used to register and run.",
        )
    )
    service_account: str = make_click_option_field(
        click.Option(
            param_decls=["--service-account", "service_account"],
            required=False,
            type=str,
            default="",
            help="Service account used when executing this workflow",
        )
    )
    wait_execution: bool = make_click_option_field(
        click.Option(
            param_decls=["--wait-execution", "wait_execution"],
            required=False,
            is_flag=True,
            default=False,
            show_default=True,
            help="Whether to wait for the execution to finish",
        )
    )
    dump_snippet: bool = make_click_option_field(
        click.Option(
            param_decls=["--dump-snippet", "dump_snippet"],
            required=False,
            is_flag=True,
            default=False,
            show_default=True,
            help="Whether to dump a code snippet instructing how to load the workflow execution using flyteremote",
        )
    )
    overwrite_cache: bool = make_click_option_field(
        click.Option(
            param_decls=["--overwrite-cache", "overwrite_cache"],
            required=False,
            is_flag=True,
            default=False,
            show_default=True,
            help="Whether to overwrite the cache if it already exists",
        )
    )
    envvars: typing.Dict[str, str] = make_click_option_field(
        click.Option(
            param_decls=["--envvars", "--env"],
            required=False,
            multiple=True,
            type=str,
            show_default=True,
            callback=key_value_callback,
            help="Environment variables to set in the container, of the format `ENV_NAME=ENV_VALUE`",
        )
    )
    tags: typing.List[str] = make_click_option_field(
        click.Option(
            param_decls=["--tags", "--tag"],
            required=False,
            multiple=True,
            type=str,
            show_default=True,
            help="Tags to set for the execution",
        )
    )
    name: str = make_click_option_field(
        click.Option(
            param_decls=["--name"],
            required=False,
            type=str,
            show_default=True,
            help="Name to assign to this execution",
        )
    )
    labels: typing.Dict[str, str] = make_click_option_field(
        click.Option(
            param_decls=["--labels", "--label"],
            required=False,
            multiple=True,
            type=str,
            show_default=True,
            callback=key_value_callback,
            help="Labels to be attached to the execution of the format `label_key=label_value`.",
        )
    )
    annotations: typing.Dict[str, str] = make_click_option_field(
        click.Option(
            param_decls=["--annotations", "--annotation"],
            required=False,
            multiple=True,
            type=str,
            show_default=True,
            callback=key_value_callback,
            help="Annotations to be attached to the execution of the format `key=value`.",
        )
    )
    raw_output_data_prefix: str = make_click_option_field(
        click.Option(
            param_decls=["--raw-output-data-prefix", "--raw-data-prefix"],
            required=False,
            type=str,
            show_default=True,
            help="File Path prefix to store raw output data."
            " Examples are file://, s3://, gs:// etc as supported by fsspec."
            " If not specified, raw data will be stored in default configured location in remote of locally"
            " to temp file system."
            + click.style(
                "Note, this is not metadata, but only the raw data location "
                "used to store Flytefile, Flytedirectory, Structuredataset,"
                " dataframes"
            ),
        )
    )
    max_parallelism: int = make_click_option_field(
        click.Option(
            param_decls=["--max-parallelism"],
            required=False,
            type=int,
            show_default=True,
            help="Number of nodes of a workflow that can be executed in parallel. If not specified,"
            " project/domain defaults are used. If 0 then it is unlimited.",
        )
    )
    disable_notifications: bool = make_click_option_field(
        click.Option(
            param_decls=["--disable-notifications"],
            required=False,
            is_flag=True,
            default=False,
            show_default=True,
            help="Should notifications be disabled for this execution.",
        )
    )
    remote: bool = make_click_option_field(
        click.Option(
            param_decls=["-r", "--remote"],
            required=False,
            is_flag=True,
            default=False,
            is_eager=True,
            show_default=True,
            help="Whether to register and run the workflow on a Flyte deployment",
        )
    )
    limit: int = make_click_option_field(
        click.Option(
            param_decls=["--limit", "limit"],
            required=False,
            type=int,
            default=50,
            hidden=True,
            show_default=True,
            help="Use this to limit number of entities to fetch",
        )
    )
    cluster_pool: str = make_click_option_field(
        click.Option(
            param_decls=["--cluster-pool", "cluster_pool"],
            required=False,
            type=str,
            default="",
            help="Assign newly created execution to a given cluster pool",
        )
    )
    computed_params: RunLevelComputedParams = field(default_factory=RunLevelComputedParams)
    _remote: typing.Optional[FlyteRemote] = None

    def remote_instance(self) -> FlyteRemote:
        if self._remote is None:
            data_upload_location = None
            if self.is_remote:
                data_upload_location = remote_fs.REMOTE_PLACEHOLDER
            self._remote = get_plugin().get_remote(self.config_file, self.project, self.domain, data_upload_location)
        return self._remote

    @property
    def is_remote(self) -> bool:
        return self.remote

    @classmethod
    def from_dict(cls, d: typing.Dict[str, typing.Any]) -> "RunLevelParams":
        return cls(**d)

    @classmethod
    def options(cls) -> typing.List[click.Option]:
        """
        Return the set of base parameters added to every pyflyte run workflow subcommand.
        """
        return [get_option_from_metadata(f.metadata) for f in fields(cls) if f.metadata]


def load_naive_entity(module_name: str, entity_name: str, project_root: str) -> typing.Union[WorkflowBase, PythonTask]:
    """
    Load the workflow of a script file.
    N.B.: it assumes that the file is self-contained, in other words, there are no relative imports.
    """
    flyte_ctx_builder = context_manager.FlyteContextManager.current_context().new_builder()
    with context_manager.FlyteContextManager.with_context(flyte_ctx_builder):
        with module_loader.add_sys_path(project_root):
            importlib.import_module(module_name)
    return module_loader.load_object_from_module(f"{module_name}.{entity_name}")


def dump_flyte_remote_snippet(execution: FlyteWorkflowExecution, project: str, domain: str):
    click.secho(
        f"""
In order to have programmatic access to the execution, use the following snippet:

from flytekit.configuration import Config
from flytekit.remote import FlyteRemote
remote = FlyteRemote(Config.auto(), default_project="{project}", default_domain="{domain}")
exec = remote.fetch_execution(name="{execution.id.name}")
remote.sync(exec)
print(exec.outputs)
    """
    )


class Entities(typing.NamedTuple):
    """
    NamedTuple to group all entities in a file
    """

    workflows: typing.List[str]
    tasks: typing.List[str]

    def all(self) -> typing.List[str]:
        e = []
        e.extend(self.workflows)
        e.extend(self.tasks)
        return e


def get_entities_in_file(filename: pathlib.Path, should_delete: bool) -> Entities:
    """
    Returns a list of flyte workflow names and list of Flyte tasks in a file.
    """
    flyte_ctx = context_manager.FlyteContextManager.current_context().new_builder()
    module_name = os.path.splitext(os.path.relpath(filename))[0].replace(os.path.sep, ".")
    with context_manager.FlyteContextManager.with_context(flyte_ctx):
        with module_loader.add_sys_path(os.getcwd()):
            importlib.import_module(module_name)

    workflows = []
    tasks = []
    module = importlib.import_module(module_name)
    for name in dir(module):
        o = module.__dict__[name]
        if isinstance(o, WorkflowBase):
            workflows.append(name)
        elif isinstance(o, PythonTask):
            tasks.append(name)

    if should_delete and os.path.exists(filename):
        os.remove(filename)
    return Entities(workflows, tasks)


def to_click_option(
    ctx: click.Context,
    flyte_ctx: FlyteContext,
    input_name: str,
    literal_var: Variable,
    python_type: typing.Type,
    default_val: typing.Any,
    required: bool,
) -> click.Option:
    """
    This handles converting workflow input types to supported click parameters with callbacks to initialize
    the input values to their expected types.
    """
    run_level_params: RunLevelParams = ctx.obj

    literal_converter = FlyteLiteralConverter(
        flyte_ctx,
        literal_type=literal_var.type,
        python_type=python_type,
        is_remote=run_level_params.is_remote,
    )

    if literal_converter.is_bool() and not default_val:
        default_val = False

    description_extra = ""
    if literal_var.type.simple == SimpleType.STRUCT:
        if default_val:
            if type(default_val) == dict or type(default_val) == list:
                default_val = json.dumps(default_val)
            else:
                default_val = cast(DataClassJsonMixin, default_val).to_json()
        if literal_var.type.metadata:
            description_extra = f": {json.dumps(literal_var.type.metadata)}"

    return click.Option(
        param_decls=[f"--{input_name}"],
        type=literal_converter.click_type,
        is_flag=literal_converter.is_bool(),
        default=default_val,
        show_default=True,
        required=required,
        help=literal_var.description + description_extra,
        callback=literal_converter.convert,
    )


def options_from_run_params(run_level_params: RunLevelParams) -> Options:
    return Options(
        labels=Labels(run_level_params.labels) if run_level_params.labels else None,
        annotations=Annotations(run_level_params.annotations) if run_level_params.annotations else None,
        raw_output_data_config=RawOutputDataConfig(output_location_prefix=run_level_params.raw_output_data_prefix)
        if run_level_params.raw_output_data_prefix
        else None,
        max_parallelism=run_level_params.max_parallelism,
        disable_notifications=run_level_params.disable_notifications,
        security_context=security.SecurityContext(
            run_as=security.Identity(k8s_service_account=run_level_params.service_account)
        )
        if run_level_params.service_account
        else None,
        notifications=[],
    )


def run_remote(
    remote: FlyteRemote,
    entity: typing.Union[FlyteWorkflow, FlyteTask, FlyteLaunchPlan],
    project: str,
    domain: str,
    inputs: typing.Dict[str, typing.Any],
    run_level_params: RunLevelParams,
    type_hints: typing.Optional[typing.Dict[str, typing.Type]] = None,
):
    """
    Helper method that executes the given remote FlyteLaunchplan, FlyteWorkflow or FlyteTask
    """

    execution = remote.execute(
        entity,
        inputs=inputs,
        project=project,
        domain=domain,
        execution_name=run_level_params.name,
        wait=run_level_params.wait_execution,
        options=options_from_run_params(run_level_params),
        type_hints=type_hints,
        overwrite_cache=run_level_params.overwrite_cache,
        envs=run_level_params.envvars,
        tags=run_level_params.tags,
        cluster_pool=run_level_params.cluster_pool,
    )

    console_url = remote.generate_console_url(execution)
    s = (
        click.style("\n[✔] ", fg="green")
        + "Go to "
        + click.style(console_url, fg="cyan")
        + " to see execution in the console."
    )
    click.echo(s)

    if run_level_params.dump_snippet:
        dump_flyte_remote_snippet(execution, project, domain)


def _update_flyte_context(params: RunLevelParams) -> FlyteContext.Builder:
    # Update the flyte context for the local execution.
    ctx = FlyteContextManager.current_context()
    output_prefix = params.raw_output_data_prefix
    if not ctx.file_access.is_remote(output_prefix):
        return ctx.current_context().new_builder()

    file_access = FileAccessProvider(
        local_sandbox_dir=tempfile.mkdtemp(prefix="flyte"), raw_output_prefix=output_prefix
    )

    # The task might run on a remote machine if raw_output_prefix is a remote path,
    # so workflow file needs to be uploaded to the remote location to make pyflyte-fast-execute work.
    if output_prefix and ctx.file_access.is_remote(output_prefix):
        with tempfile.TemporaryDirectory() as tmp_dir:
            archive_fname = pathlib.Path(os.path.join(tmp_dir, "script_mode.tar.gz"))
            compress_scripts(params.computed_params.project_root, str(archive_fname), params.computed_params.module)
            remote_dir = file_access.get_random_remote_directory()
            remote_archive_fname = f"{remote_dir}/script_mode.tar.gz"
            file_access.put_data(str(archive_fname), remote_archive_fname)

        ctx_builder = ctx.with_file_access(ctx.file_access).with_serialization_settings(
            SerializationSettings(
                source_root=params.computed_params.project_root,
                image_config=params.image_config,
                fast_serialization_settings=FastSerializationSettings(
                    enabled=True,
                    destination_dir=params.destination_dir,
                    distribution_location=remote_archive_fname,
                ),
            )
        )
        return ctx_builder.with_file_access(file_access)


def run_command(ctx: click.Context, entity: typing.Union[PythonFunctionWorkflow, PythonTask]):
    """
    Returns a function that is used to implement WorkflowCommand and execute a flyte workflow.
    """

    def _run(*args, **kwargs):
        """
        Click command function that is used to execute a flyte workflow from the given entity in the file.
        """
        # By the time we get to this function, all the loading has already happened

        run_level_params: RunLevelParams = ctx.obj
        logger.info(f"Running {entity.name} with {kwargs} and run_level_params {run_level_params}")

        click.secho(f"Running Execution on {'Remote' if run_level_params.is_remote else 'local'}.", fg="cyan")
        try:
            inputs = {}
            for input_name, _ in entity.python_interface.inputs.items():
                inputs[input_name] = kwargs.get(input_name)

            if not run_level_params.is_remote:
                with FlyteContextManager.with_context(_update_flyte_context(run_level_params)):
                    if run_level_params.envvars:
                        for env_var, value in run_level_params.envvars.items():
                            os.environ[env_var] = value
                    if run_level_params.overwrite_cache:
                        os.environ["FLYTE_LOCAL_CACHE_OVERWRITE"] = "true"
                    output = entity(**inputs)
                    if inspect.iscoroutine(output):
                        # TODO: make eager mode workflows run with local-mode
                        output = asyncio.run(output)
                    click.echo(output)
                    return

            remote = run_level_params.remote_instance()
            config_file = run_level_params.config_file

            image_config = run_level_params.image_config
            image_config = patch_image_config(config_file, image_config)

            with context_manager.FlyteContextManager.with_context(remote.context.new_builder()):
                remote_entity = remote.register_script(
                    entity,
                    project=run_level_params.project,
                    domain=run_level_params.domain,
                    image_config=image_config,
                    destination_dir=run_level_params.destination_dir,
                    source_path=run_level_params.computed_params.project_root,
                    module_name=run_level_params.computed_params.module,
                    copy_all=run_level_params.copy_all,
                )

                run_remote(
                    remote,
                    remote_entity,
                    run_level_params.project,
                    run_level_params.domain,
                    inputs,
                    run_level_params,
                    type_hints=entity.python_interface.inputs,
                )
        finally:
            if run_level_params.computed_params.temp_file_name:
                os.remove(run_level_params.computed_params.temp_file_name)

    return _run


class DynamicEntityLaunchCommand(click.RichCommand):
    """
    This is a dynamic command that is created for each launch plan. This is used to execute a launch plan.
    It will fetch the launch plan from remote and create parameters from all the inputs of the launch plan.
    """

    LP_LAUNCHER = "lp"
    TASK_LAUNCHER = "task"

    def __init__(self, name: str, h: str, entity_name: str, launcher: str, **kwargs):
        super().__init__(name=name, help=h, **kwargs)
        self._entity_name = entity_name
        self._launcher = launcher
        self._entity = None

    def _fetch_entity(self, ctx: click.Context) -> typing.Union[FlyteLaunchPlan, FlyteTask]:
        if self._entity:
            return self._entity
        run_level_params: RunLevelParams = ctx.obj
        r = run_level_params.remote_instance()
        if self._launcher == self.LP_LAUNCHER:
            entity = r.fetch_launch_plan(run_level_params.project, run_level_params.domain, self._entity_name)
        else:
            entity = r.fetch_task(run_level_params.project, run_level_params.domain, self._entity_name)
        self._entity = entity
        return entity

    def _get_params(
        self,
        ctx: click.Context,
        inputs: typing.Dict[str, Variable],
        native_inputs: typing.Dict[str, type],
        fixed: typing.Optional[typing.Dict[str, Literal]] = None,
        defaults: typing.Optional[typing.Dict[str, Parameter]] = None,
    ) -> typing.List["click.Parameter"]:
        params = []
        flyte_ctx = context_manager.FlyteContextManager.current_context()
        for name, var in inputs.items():
            if fixed and name in fixed:
                continue
            required = True
            default_val = None
            if defaults and name in defaults:
                if not defaults[name].required:
                    required = False
                    default_val = literal_string_repr(defaults[name].default) if defaults[name].default else None
            params.append(to_click_option(ctx, flyte_ctx, name, var, native_inputs[name], default_val, required))
        return params

    def get_params(self, ctx: click.Context) -> typing.List["click.Parameter"]:
        if not self.params:
            self.params = []
            entity = self._fetch_entity(ctx)
            if entity.interface:
                if entity.interface.inputs:
                    types = TypeEngine.guess_python_types(entity.interface.inputs)
                    if isinstance(entity, FlyteLaunchPlan):
                        self.params = self._get_params(
                            ctx,
                            entity.interface.inputs,
                            types,
                            entity.fixed_inputs.literals,
                            entity.default_inputs.parameters,
                        )
                    else:
                        self.params = self._get_params(ctx, entity.interface.inputs, types)

        return super().get_params(ctx)

    def invoke(self, ctx: click.Context) -> typing.Any:
        """
        Default or None values should be ignored. Only values that are provided by the user should be passed to the
        remote execution.
        """
        run_level_params: RunLevelParams = ctx.obj
        r = run_level_params.remote_instance()
        entity = self._fetch_entity(ctx)
        run_remote(
            r,
            entity,
            run_level_params.project,
            run_level_params.domain,
            ctx.params,
            run_level_params,
            type_hints=entity.python_interface.inputs if entity.python_interface else None,
        )


class RemoteEntityGroup(click.RichGroup):
    """
    click multicommand that retrieves launchplans from a remote flyte instance and executes them.
    """

    LAUNCHPLAN_COMMAND = "remote-launchplan"
    WORKFLOW_COMMAND = "remote-workflow"
    TASK_COMMAND = "remote-task"

    def __init__(self, command_name: str):
        super().__init__(
            name=command_name,
            help=f"Retrieve {command_name} from a remote flyte instance and execute them.",
            params=[
                click.Option(
                    ["--limit", "limit"],
                    help=f"Limit the number of {command_name}'s to retrieve.",
                    default=50,
                    show_default=True,
                )
            ],
        )
        self._command_name = command_name
        self._entities = []

    def _get_entities(self, r: FlyteRemote, project: str, domain: str, limit: int) -> typing.List[str]:
        """
        Retreieves the right entities from the remote flyte instance.
        """
        if self._command_name == self.LAUNCHPLAN_COMMAND:
            lps = r.client.list_launch_plan_ids_paginated(project=project, domain=domain, limit=limit)
            return [l.name for l in lps[0]]
        elif self._command_name == self.WORKFLOW_COMMAND:
            wfs = r.client.list_workflow_ids_paginated(project=project, domain=domain, limit=limit)
            return [w.name for w in wfs[0]]
        elif self._command_name == self.TASK_COMMAND:
            tasks = r.client.list_task_ids_paginated(project=project, domain=domain, limit=limit)
            return [t.name for t in tasks[0]]
        return []

    def list_commands(self, ctx):
        if self._entities or ctx.obj is None:
            return self._entities

        run_level_params: RunLevelParams = ctx.obj
        r = run_level_params.remote_instance()
        progress = Progress(transient=True)
        task = progress.add_task(f"[cyan]Gathering [{run_level_params.limit}] remote LaunchPlans...", total=None)
        with progress:
            progress.start_task(task)
            try:
                self._entities = self._get_entities(
                    r, run_level_params.project, run_level_params.domain, run_level_params.limit
                )
                return self._entities
            except FlyteSystemException as e:
                pretty_print_exception(e)
                return []

    def get_command(self, ctx, name):
        if self._command_name in [self.LAUNCHPLAN_COMMAND, self.WORKFLOW_COMMAND]:
            return DynamicEntityLaunchCommand(
                name=name,
                h=f"Execute a {self._command_name}.",
                entity_name=name,
                launcher=DynamicEntityLaunchCommand.LP_LAUNCHER,
            )
        return DynamicEntityLaunchCommand(
            name=name,
            h=f"Execute a {self._command_name}.",
            entity_name=name,
            launcher=DynamicEntityLaunchCommand.TASK_LAUNCHER,
        )


class WorkflowCommand(click.RichGroup):
    """
    click multicommand at the python file layer, subcommands should be all the workflows in the file.
    """

    def __init__(self, filename: str, *args, **kwargs):
        super().__init__(*args, **kwargs)

        ctx = context_manager.FlyteContextManager.current_context()
        if ctx.file_access.is_remote(filename):
            local_path = os.path.join(os.path.curdir, filename.rsplit("/", 1)[1])
            ctx.file_access.download(filename, local_path)
            self._filename = pathlib.Path(local_path).resolve()
            self._should_delete = True
        else:
            self._filename = pathlib.Path(filename).resolve()
            self._should_delete = False
        self._entities = None

    def list_commands(self, ctx):
        if self._entities:
            return self._entities.all()
        entities = get_entities_in_file(self._filename, self._should_delete)
        self._entities = entities
        return entities.all()

    def _create_command(
        self,
        ctx: click.Context,
        entity_name: str,
        run_level_params: RunLevelParams,
        loaded_entity: typing.Any,
        is_workflow: bool,
    ):
        """
        Delegate that creates the command for a given entity.
        """

        # If this is a remote execution, which we should know at this point, then create the remote object
        r = run_level_params.remote_instance()
        flyte_ctx = r.context

        # Add options for each of the workflow inputs
        params = []
        for input_name, input_type_val in loaded_entity.python_interface.inputs_with_defaults.items():
            literal_var = loaded_entity.interface.inputs.get(input_name)
            python_type, default_val = input_type_val
            required = type(None) not in get_args(python_type) and default_val is None
            params.append(to_click_option(ctx, flyte_ctx, input_name, literal_var, python_type, default_val, required))

        entity_type = "Workflow" if is_workflow else "Task"
        h = f"{click.style(entity_type, bold=True)} ({run_level_params.computed_params.module}.{entity_name})"
        if loaded_entity.__doc__:
            h = h + click.style(f"{loaded_entity.__doc__}", dim=True)
        cmd = click.RichCommand(
            name=entity_name,
            params=params,
            callback=run_command(ctx, loaded_entity),
            help=h,
        )
        return cmd

    def get_command(self, ctx, exe_entity):
        """
        This command uses the filename with which this command was created, and the string name of the entity passed
          after the Python filename on the command line, to load the Python object, and then return the Command that
          click should run.
        :param ctx: The click Context object.
        :param exe_entity: string of the flyte entity provided by the user. Should be the name of a workflow, or task
          function.
        :return:
        """
        is_workflow = False
        if self._entities:
            is_workflow = exe_entity in self._entities.workflows
        if not os.path.exists(self._filename):
            raise ValueError(f"File {self._filename} does not exist")
        rel_path = os.path.relpath(self._filename)
        if rel_path.startswith(".."):
            raise ValueError(
                f"You must call pyflyte from the same or parent dir, {self._filename} not under {os.getcwd()}"
            )

        project_root = _find_project_root(self._filename)

        # Find the relative path for the filename relative to the root of the project.
        # N.B.: by construction project_root will necessarily be an ancestor of the filename passed in as
        # a parameter.
        rel_path = self._filename.relative_to(project_root)
        module = os.path.splitext(rel_path)[0].replace(os.path.sep, ".")

        run_level_params: RunLevelParams = ctx.obj

        # update computed params
        run_level_params.computed_params.project_root = project_root
        run_level_params.computed_params.module = module

        if self._should_delete:
            run_level_params.computed_params.temp_file_name = self._filename

        entity = load_naive_entity(module, exe_entity, project_root)

        return self._create_command(ctx, exe_entity, run_level_params, entity, is_workflow)


class RunCommand(click.RichGroup):
    """
    A click command group for registering and executing flyte workflows & tasks in a file.
    """

    _run_params: typing.Type[RunLevelParams] = RunLevelParams

    def __init__(self, *args, **kwargs):
        if "params" not in kwargs:
            params = self._run_params.options()
            kwargs["params"] = params
        super().__init__(*args, **kwargs)
        self._files = []

    def list_commands(self, ctx, add_remote: bool = True):
        if self._files:
            return self._files
        self._files = [str(p) for p in pathlib.Path(".").glob("*.py") if str(p) != "__init__.py"]
        self._files = sorted(self._files)
        if add_remote:
            self._files = self._files + [
                RemoteEntityGroup.LAUNCHPLAN_COMMAND,
                RemoteEntityGroup.WORKFLOW_COMMAND,
                RemoteEntityGroup.TASK_COMMAND,
            ]
        return self._files

    def get_command(self, ctx, filename):
        if ctx.obj is None:
            ctx.obj = {}
        if not isinstance(ctx.obj, self._run_params):
            params = {}
            params.update(ctx.params)
            params.update(ctx.obj)
            ctx.obj = self._run_params.from_dict(params)
        if filename == RemoteEntityGroup.LAUNCHPLAN_COMMAND:
            return RemoteEntityGroup(RemoteEntityGroup.LAUNCHPLAN_COMMAND)
        elif filename == RemoteEntityGroup.WORKFLOW_COMMAND:
            return RemoteEntityGroup(RemoteEntityGroup.WORKFLOW_COMMAND)
        elif filename == RemoteEntityGroup.TASK_COMMAND:
            return RemoteEntityGroup(RemoteEntityGroup.TASK_COMMAND)
        return WorkflowCommand(filename, name=filename, help=f"Run a [workflow|task] from {filename}")


_run_help = """
This command can execute either a workflow or a task from the command line, allowing for fully self-contained scripts.
Tasks and workflows can be imported from other files.

Note: This command is compatible with regular Python packages, but not with namespace packages.
When determining the root of your project, it identifies the first folder without an ``__init__.py`` file.
"""

run = RunCommand(
    name="run",
    help=_run_help,
)
