import os
import sys
import typing
from enum import Enum as _Enum

import rich_click as click

from flytekit.clis.sdk_in_container import constants
from flytekit.clis.sdk_in_container.constants import CTX_PACKAGES
from flytekit.configuration import FastSerializationSettings, ImageConfig, SerializationSettings
from flytekit.exceptions.scopes import system_entry_point
from flytekit.interaction.click_types import key_value_callback
from flytekit.tools.fast_registration import fast_package
from flytekit.tools.repo import serialize_to_folder

CTX_IMAGE = "image"
CTX_LOCAL_SRC_ROOT = "local_source_root"
CTX_FLYTEKIT_VIRTUALENV_ROOT = "flytekit_virtualenv_root"
CTX_PYTHON_INTERPRETER = "python_interpreter"
CTX_ENV = "env"


class SerializationMode(_Enum):
    DEFAULT = 0
    FAST = 1


@system_entry_point
def serialize_all(
    pkgs: typing.List[str] = None,
    local_source_root: typing.Optional[str] = None,
    folder: typing.Optional[str] = None,
    mode: typing.Optional[SerializationMode] = None,
    image_config: typing.Optional[ImageConfig] = None,
    flytekit_virtualenv_root: typing.Optional[str] = None,
    python_interpreter: typing.Optional[str] = None,
    config_file: typing.Optional[str] = None,
    env: typing.Optional[typing.Dict[str, str]] = None,
):
    """
    This function will write to the folder specified the following protobuf types ::
        flyteidl.admin.launch_plan_pb2.LaunchPlan
        flyteidl.admin.workflow_pb2.WorkflowSpec
        flyteidl.admin.task_pb2.TaskSpec

    These can be inspected by calling (in the launch plan case) ::
        flyte-cli parse-proto -f filename.pb -p flyteidl.admin.launch_plan_pb2.LaunchPlan

    See :py:class:`flytekit.models.core.identifier.ResourceType` to match the trailing index in the file name with the
    entity type.
    :param pkgs: Dot-delimited Python packages/subpackages to look into for serialization.
    :param local_source_root: Where to start looking for the code.
    :param folder: Where to write the output protobuf files
    :param mode: Regular vs fast
    :param image_config: ImageConfig object to use
    :param flytekit_virtualenv_root: The full path of the virtual env in the container.
    """

    if not (mode == SerializationMode.DEFAULT or mode == SerializationMode.FAST):
        raise AssertionError(f"Unrecognized serialization mode: {mode}")

    serialization_settings = SerializationSettings(
        image_config=image_config or ImageConfig.auto(config_file),
        fast_serialization_settings=FastSerializationSettings(
            enabled=mode == SerializationMode.FAST,
            # TODO: if we want to move the destination dir as a serialization argument, we should initialize it here
        ),
        flytekit_virtualenv_root=flytekit_virtualenv_root,
        python_interpreter=python_interpreter,
        env=env,
    )

    serialize_to_folder(pkgs, serialization_settings, local_source_root, folder)


@click.group("serialize", cls=click.RichGroup)
@click.option(
    "-i",
    "--image",
    "image_config",
    required=False,
    multiple=True,
    type=click.UNPROCESSED,
    callback=ImageConfig.validate_image,
    help="A fully qualified tag for an docker image, for example ``somedocker.com/myimage:someversion123``. This is a "
    "multi-option and can be of the form ``--image xyz.io/docker:latest"
    " --image my_image=xyz.io/docker2:latest``. Note, the ``name=image_uri``. The name is optional, if not "
    "provided the image will be used as the default image. All the names have to be unique, and thus "
    "there can only be one ``--image`` option with no name.",
)
@click.option(
    "--local-source-root",
    required=False,
    default=lambda: os.getcwd(),
    help="Root dir for Python code containing workflow definitions to operate on when not the current working directory. "
    "Optional when running ``pyflyte serialize`` in out-of-container-mode and your code lies outside of your working directory.",
)
@click.option(
    "--in-container-config-path",
    required=False,
    help="This is where the configuration for your task lives inside the container. "
    "The reason it needs to be a separate option is because this pyflyte utility cannot know where the Dockerfile "
    "writes the config file to. Required for running ``pyflyte serialize`` in out-of-container-mode",
)
@click.option(
    "--in-container-virtualenv-root",
    required=False,
    help="DEPRECATED: This flag is ignored! This is the root of the flytekit virtual env in your container. "
    "The reason it needs to be a separate option is because this pyflyte utility cannot know where flytekit is "
    "installed inside your container. Required for running `pyflyte serialize` in out of container mode when "
    "your container installs the flytekit virtualenv outside of the default `/opt/venv`",
)
@click.option(
    "--env",
    "--envvars",
    required=False,
    multiple=True,
    type=str,
    callback=key_value_callback,
    help="Environment variables to set in the container, of the format `ENV_NAME=ENV_VALUE`",
)
@click.pass_context
def serialize(
    ctx,
    image_config: ImageConfig,
    local_source_root,
    in_container_config_path,
    in_container_virtualenv_root,
    env: typing.Optional[typing.Dict[str, str]],
):
    """
    This command produces protobufs for tasks and templates.
    For tasks, one pb file is produced for each task, representing one TaskTemplate object.
    For workflows, one pb file is produced for each workflow, representing a WorkflowClosure object. The closure
    object contains the WorkflowTemplate, along with the relevant tasks for that workflow. In lieu of Admin,
    this serialization step will set the URN of the tasks to the fully qualified name of the task function.
    """
    ctx.obj[CTX_IMAGE] = image_config
    ctx.obj[CTX_LOCAL_SRC_ROOT] = local_source_root
    ctx.obj[CTX_ENV] = env
    click.echo(f"Serializing Flyte elements with image {image_config}")

    if in_container_virtualenv_root:
        ctx.obj[CTX_FLYTEKIT_VIRTUALENV_ROOT] = in_container_virtualenv_root
        ctx.obj[CTX_PYTHON_INTERPRETER] = os.path.join(in_container_virtualenv_root, "/bin/python3")
    else:
        # For in container serialize we make sure to never accept an override the entrypoint path and determine it here
        # instead.
        import flytekit

        flytekit_install_loc = os.path.abspath(flytekit.__file__)
        ctx.obj[CTX_FLYTEKIT_VIRTUALENV_ROOT] = os.path.dirname(flytekit_install_loc)
        ctx.obj[CTX_PYTHON_INTERPRETER] = sys.executable


@click.command("workflows", cls=click.RichCommand)
# For now let's just assume that the directory needs to exist. If you're docker run -v'ing, docker will create the
# directory for you so it shouldn't be a problem.
@click.option("-f", "--folder", type=click.Path(exists=True))
@click.pass_context
def workflows(ctx, folder=None):
    if folder:
        click.echo(f"Writing output to {folder}")

    pkgs = ctx.obj[CTX_PACKAGES]
    dir = ctx.obj[CTX_LOCAL_SRC_ROOT]
    serialize_all(
        pkgs,
        dir,
        folder,
        SerializationMode.DEFAULT,
        image_config=ctx.obj[CTX_IMAGE],
        flytekit_virtualenv_root=ctx.obj[CTX_FLYTEKIT_VIRTUALENV_ROOT],
        python_interpreter=ctx.obj[CTX_PYTHON_INTERPRETER],
        config_file=ctx.obj.get(constants.CTX_CONFIG_FILE, None),
        env=ctx.obj.get(CTX_ENV, None),
    )


@click.group("fast", cls=click.RichGroup)
@click.pass_context
def fast(ctx):
    pass


@click.command("workflows", cls=click.RichCommand)
@click.option(
    "--deref-symlinks",
    default=False,
    is_flag=True,
    help="Enables symlink dereferencing when packaging files in fast registration",
)
@click.option("-f", "--folder", type=click.Path(exists=True))
@click.pass_context
def fast_workflows(ctx, folder=None, deref_symlinks=False):
    if folder:
        click.echo(f"Writing output to {folder}")

    source_dir = ctx.obj[CTX_LOCAL_SRC_ROOT]
    # Write using gzip
    archive_fname = fast_package(source_dir, folder, deref_symlinks)
    click.echo(f"Wrote compressed archive to {archive_fname}")

    pkgs = ctx.obj[CTX_PACKAGES]
    dir = ctx.obj[CTX_LOCAL_SRC_ROOT]
    serialize_all(
        pkgs,
        dir,
        folder,
        SerializationMode.FAST,
        image_config=ctx.obj[CTX_IMAGE],
        flytekit_virtualenv_root=ctx.obj[CTX_FLYTEKIT_VIRTUALENV_ROOT],
        python_interpreter=ctx.obj[CTX_PYTHON_INTERPRETER],
        config_file=ctx.obj.get(constants.CTX_CONFIG_FILE, None),
        env=ctx.obj.get(CTX_ENV, None),
    )


fast.add_command(fast_workflows)
serialize.add_command(workflows)
serialize.add_command(fast)
