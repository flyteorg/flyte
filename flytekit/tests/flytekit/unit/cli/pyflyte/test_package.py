import os
import shutil

from click.testing import CliRunner
from flyteidl.admin import task_pb2

import flytekit
import flytekit.clis.sdk_in_container.utils
import flytekit.configuration
import flytekit.tools.serialize_helpers
from flytekit import TaskMetadata
from flytekit.clis.sdk_in_container import pyflyte
from flytekit.core import context_manager
from flytekit.models.admin.workflow import WorkflowSpec
from flytekit.models.core.identifier import Identifier, ResourceType
from flytekit.models.launch_plan import LaunchPlan
from flytekit.models.task import TaskSpec
from flytekit.remote import FlyteTask
from flytekit.remote.interface import TypedInterface
from flytekit.remote.remote_callable import RemoteEntity

sample_file_contents = """
from flytekit import task, workflow

@task(cache=True, cache_version="1", retries=3)
def sum(x: int, y: int) -> int:
    return x + y

@task(cache=True, cache_version="1", retries=3)
def square(z: int) -> int:
    return z*z

@workflow
def my_workflow(x: int, y: int) -> int:
    return sum(x=square(z=x), y=square(z=y))
"""


@flytekit.task
def foo():
    pass


@flytekit.workflow
def wf():
    return foo()


def test_get_registrable_entities():
    ctx = context_manager.FlyteContextManager.current_context().with_serialization_settings(
        flytekit.configuration.SerializationSettings(
            project="p",
            domain="d",
            version="v",
            image_config=flytekit.configuration.ImageConfig(
                default_image=flytekit.configuration.Image("def", "docker.io/def", "latest")
            ),
        )
    )
    context_manager.FlyteEntities.entities = [
        foo,
        wf,
        "str",
        FlyteTask(
            id=Identifier(ResourceType.TASK, "p", "d", "n", "v"),
            type="t",
            metadata=TaskMetadata().to_taskmetadata_model(),
            interface=TypedInterface(inputs={}, outputs={}),
            custom=None,
        ),
    ]
    entities = flytekit.tools.serialize_helpers.get_registrable_entities(ctx)
    assert entities
    assert len(entities) == 3

    for e in entities:
        if isinstance(e, RemoteEntity):
            assert False, "found unexpected remote entity"
        if isinstance(e, WorkflowSpec) or isinstance(e, TaskSpec) or isinstance(e, LaunchPlan):
            continue
        assert False, f"found unknown entity {type(e)}"


def test_package_with_fast_registration_and_envvars():
    runner = CliRunner()
    with runner.isolated_filesystem():
        os.makedirs("core", exist_ok=True)
        with open(os.path.join("core", "sample.py"), "w") as f:
            f.write(sample_file_contents)
            f.close()
        result = runner.invoke(
            pyflyte.main,
            [
                "--pkgs",
                "core",
                "package",
                "--image",
                "core:v1",
                "--fast",
                "--env",
                "abc=42",
                "--env",
                "euler=2.71828",
            ],
        )
        assert result.exit_code == 0
        assert "Successfully serialized" in result.output
        assert "Successfully packaged" in result.output

        # verify existence of flyte-package.tgz file
        assert os.path.exists("flyte-package.tgz")

        # verify the contents of the flyte-package.tgz file
        import tarfile

        # Uncompress flyte-package.tgz
        tarfile.open("flyte-package.tgz", "r:gz").extractall()

        # Load the proto message from file 3_core.sample.sum_1.pb
        task_spec = task_pb2.TaskSpec()
        task_spec.ParseFromString(open("3_core.sample.sum_1.pb", "rb").read())

        assert task_spec.template.container.env[0].key == "abc"
        assert task_spec.template.container.env[0].value == "42"
        assert task_spec.template.container.env[1].key == "euler"
        assert task_spec.template.container.env[1].value == "2.71828"

        result = runner.invoke(pyflyte.main, ["--pkgs", "core", "package", "--image", "core:v1", "--fast"])
        assert result.exit_code == 2
        assert "flyte-package.tgz already exists, specify -f to override" in result.output
        result = runner.invoke(
            pyflyte.main,
            [
                "--pkgs",
                "core",
                "package",
                "--image",
                "core:v1",
                "--fast",
                "--force",
                "--env",
                "k1=v1",
                "--env",
                "pi=3.14159265",
            ],
        )
        assert result.exit_code == 0
        assert "deleting and re-creating it" in result.output

        tarfile.open("flyte-package.tgz", "r:gz").extractall()

        # Load the proto message from file 3_core.sample.sum_1.pb
        task_spec = task_pb2.TaskSpec()
        task_spec.ParseFromString(open("3_core.sample.sum_1.pb", "rb").read())

        assert task_spec.template.container.env[0].key == "k1"
        assert task_spec.template.container.env[0].value == "v1"
        assert task_spec.template.container.env[1].key == "pi"
        assert task_spec.template.container.env[1].value == "3.14159265"
        shutil.rmtree("core")


def test_package():
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(
            pyflyte.main,
            [
                "--pkgs",
                "flytekit.unit.cli.pyflyte.test_package",
                "package",
                "--image",
                "myapp:03eccc1cf101adbd8c4734dba865d3fdeb720aa7",
            ],
        )
        assert result.exit_code == 1
        assert result.output is not None


def test_pkgs():
    pp = flytekit.clis.sdk_in_container.utils.validate_package(None, None, ["a.b", "a.c,b.a", "cc.a"])
    assert pp == ["a.b", "a.c", "b.a", "cc.a"]


def test_package_with_no_pkgs():
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(pyflyte.main, ["package"])
        assert result.exit_code == 1
        assert "No packages to scan for flyte entities. Aborting!" in result.output


def test_package_with_envs_wrong_format():
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(
            pyflyte.main,
            [
                "--pkgs",
                "flytekit.unit.cli.pyflyte.test_package",
                "package",
                "--image",
                "myapp:03eccc1cf101adbd8c4734dba865d3fdeb720aa7",
                "--env",
                "Key0:Value0",
            ],
        )
        assert result.exit_code == 2
        assert "Expected key-value pair of the form key=value, got" in result.output
