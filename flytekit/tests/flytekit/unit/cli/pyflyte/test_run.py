import json
import os
import pathlib
import sys
from enum import Enum

import mock
import pytest
from click.testing import CliRunner

from flytekit.clis.sdk_in_container import pyflyte
from flytekit.clis.sdk_in_container.run import RunLevelParams, get_entities_in_file, run_command
from flytekit.configuration import Config, Image, ImageConfig
from flytekit.core.task import task
from flytekit.image_spec.image_spec import ImageBuildEngine
from flytekit.interaction.click_types import DirParamType, FileParamType
from flytekit.remote import FlyteRemote

pytest.importorskip("pandas")

WORKFLOW_FILE = os.path.join(os.path.dirname(os.path.realpath(__file__)), "workflow.py")
REMOTE_WORKFLOW_FILE = "https://raw.githubusercontent.com/flyteorg/flytesnacks/8337b64b33df046b2f6e4cba03c74b7bdc0c4fb1/cookbook/core/flyte_basics/basic_workflow.py"
IMPERATIVE_WORKFLOW_FILE = os.path.join(os.path.dirname(os.path.realpath(__file__)), "imperative_wf.py")
DIR_NAME = os.path.dirname(os.path.realpath(__file__))


@pytest.fixture
def remote():
    with mock.patch("flytekit.clients.friendly.SynchronousFlyteClient") as mock_client:
        flyte_remote = FlyteRemote(config=Config.auto(), default_project="p1", default_domain="d1")
        flyte_remote._client = mock_client
        return flyte_remote


@pytest.mark.parametrize(
    "remote_flag",
    [
        "-r",
        "--remote",
    ],
)
def test_pyflyte_run_wf(remote, remote_flag):
    with mock.patch("flytekit.configuration.plugin.FlyteRemote"):
        runner = CliRunner()
        module_path = WORKFLOW_FILE
        result = runner.invoke(
            pyflyte.main, ["run", remote_flag, module_path, "my_wf", "--help"], catch_exceptions=False
        )

        assert result.exit_code == 0


def test_imperative_wf():
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.main,
        ["run", IMPERATIVE_WORKFLOW_FILE, "wf", "--in1", "hello", "--in2", "world"],
        catch_exceptions=False,
    )
    assert result.exit_code == 0


def test_copy_all_files():
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.main,
        ["run", "--copy-all", IMPERATIVE_WORKFLOW_FILE, "wf", "--in1", "hello", "--in2", "world"],
        catch_exceptions=False,
    )
    assert result.exit_code == 0


def test_remote_files():
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.main,
        ["run", REMOTE_WORKFLOW_FILE, "my_wf", "--a", "1", "--b", "Hello"],
        catch_exceptions=False,
    )
    assert result.exit_code == 0


def test_pyflyte_run_cli():
    runner = CliRunner()
    parquet_file = os.path.join(DIR_NAME, "testdata/df.parquet")
    result = runner.invoke(
        pyflyte.main,
        [
            "run",
            WORKFLOW_FILE,
            "my_wf",
            "--a",
            "1",
            "--b",
            "Hello",
            "--c",
            "1.1",
            "--d",
            '{"i":1,"a":["h","e"]}',
            "--e",
            "[1,2,3]",
            "--f",
            '{"x":1.0, "y":2.0}',
            "--g",
            parquet_file,
            "--i",
            "2020-05-01",
            "--j",
            "20H",
            "--k",
            "RED",
            "--l",
            '{"hello": "world"}',
            "--remote",
            os.path.join(DIR_NAME, "testdata"),
            "--image",
            os.path.join(DIR_NAME, "testdata"),
            "--h",
            "--n",
            json.dumps([{"x": parquet_file}]),
            "--o",
            json.dumps({"x": [parquet_file]}),
            "--p",
            "Any",
            "--q",
            DIR_NAME,
        ],
        catch_exceptions=False,
    )
    assert result.exit_code == 0, result.stdout


@pytest.mark.parametrize(
    "input",
    ["1", os.path.join(DIR_NAME, "testdata/df.parquet"), '{"x":1.0, "y":2.0}', "2020-05-01", "RED"],
)
def test_union_type1(input):
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.main,
        [
            "run",
            os.path.join(DIR_NAME, "workflow.py"),
            "test_union1",
            "--a",
            input,
        ],
        catch_exceptions=False,
    )
    assert result.exit_code == 0


@pytest.mark.parametrize(
    "input",
    [2.0, '{"i":1,"a":["h","e"]}', "[1, 2, 3]"],
)
def test_union_type2(input):
    runner = CliRunner()
    env = "foo=bar"
    result = runner.invoke(
        pyflyte.main,
        [
            "--verbose",
            "run",
            "--overwrite-cache",
            "--envvars",
            env,
            "--tag",
            "flyte",
            "--tag",
            "hello",
            os.path.join(DIR_NAME, "workflow.py"),
            "test_union2",
            "--a",
            input,
        ],
        catch_exceptions=False,
    )
    assert result.exit_code == 0, result.stdout


def test_union_type_with_invalid_input():
    runner = CliRunner()
    result = runner.invoke(
        pyflyte.main,
        [
            "--verbose",
            "run",
            os.path.join(DIR_NAME, "workflow.py"),
            "test_union2",
            "--a",
            "hello",
        ],
        catch_exceptions=False,
    )
    assert result.exit_code == 2


def test_get_entities_in_file():
    e = get_entities_in_file(WORKFLOW_FILE, False)
    assert e.workflows == ["my_wf", "wf_with_env_vars", "wf_with_none"]
    assert e.tasks == [
        "get_subset_df",
        "print_all",
        "show_sd",
        "task_with_env_vars",
        "task_with_optional",
        "test_union1",
        "test_union2",
    ]
    assert e.all() == [
        "my_wf",
        "wf_with_env_vars",
        "wf_with_none",
        "get_subset_df",
        "print_all",
        "show_sd",
        "task_with_env_vars",
        "task_with_optional",
        "test_union1",
        "test_union2",
    ]


@pytest.mark.parametrize(
    "working_dir, wf_path",
    [
        (pathlib.Path("test_nested_wf"), os.path.join("a", "b", "c", "d", "wf.py")),
        (pathlib.Path("test_nested_wf", "a"), os.path.join("b", "c", "d", "wf.py")),
        (pathlib.Path("test_nested_wf", "a", "b"), os.path.join("c", "d", "wf.py")),
        (pathlib.Path("test_nested_wf", "a", "b", "c"), os.path.join("d", "wf.py")),
        (pathlib.Path("test_nested_wf", "a", "b", "c", "d"), os.path.join("wf.py")),
    ],
)
def test_nested_workflow(working_dir, wf_path, monkeypatch: pytest.MonkeyPatch):
    runner = CliRunner()
    base_path = os.path.dirname(os.path.realpath(__file__))
    # Change working directory without side-effects (i.e. just for this test)
    monkeypatch.chdir(os.path.join(base_path, working_dir))
    result = runner.invoke(
        pyflyte.main,
        [
            "run",
            wf_path,
            "wf_id",
            "--m",
            "Running Execution on local.",
        ],
        catch_exceptions=False,
    )
    assert result.stdout.strip() == "Running Execution on local.\nRunning Execution on local."
    assert result.exit_code == 0


@pytest.mark.parametrize(
    "wf_path",
    [("collection_wf.py"), ("map_wf.py"), ("dataclass_wf.py")],
)
def test_list_default_arguments(wf_path):
    runner = CliRunner()
    dir_name = os.path.dirname(os.path.realpath(__file__))
    result = runner.invoke(
        pyflyte.main,
        [
            "run",
            os.path.join(dir_name, "default_arguments", wf_path),
            "wf",
        ],
        catch_exceptions=False,
    )
    assert result.exit_code == 0


# default case, what comes from click if no image is specified, the click param is configured to use the default.
ic_result_1 = ImageConfig(
    default_image=Image(name="default", fqn="ghcr.io/flyteorg/mydefault", tag="py3.9-latest"),
    images=[Image(name="default", fqn="ghcr.io/flyteorg/mydefault", tag="py3.9-latest")],
)
# test that command line args are merged with the file
ic_result_2 = ImageConfig(
    default_image=Image(name="default", fqn="cr.flyte.org/flyteorg/flytekit", tag="py3.9-latest"),
    images=[
        Image(name="default", fqn="cr.flyte.org/flyteorg/flytekit", tag="py3.9-latest"),
        Image(name="asdf", fqn="ghcr.io/asdf/asdf", tag="latest"),
        Image(name="xyz", fqn="docker.io/xyz", tag="latest"),
        Image(name="abc", fqn="docker.io/abc", tag=None),
    ],
)
# test that command line args override the file
ic_result_3 = ImageConfig(
    default_image=Image(name="default", fqn="cr.flyte.org/flyteorg/flytekit", tag="py3.9-latest"),
    images=[
        Image(name="default", fqn="cr.flyte.org/flyteorg/flytekit", tag="py3.9-latest"),
        Image(name="xyz", fqn="ghcr.io/asdf/asdf", tag="latest"),
        Image(name="abc", fqn="docker.io/abc", tag=None),
    ],
)

ic_result_4 = ImageConfig(
    default_image=Image(name="default", fqn="flytekit", tag="DgQMqIi61py4I4P5iOeS0Q.."),
    images=[
        Image(name="default", fqn="flytekit", tag="DgQMqIi61py4I4P5iOeS0Q.."),
        Image(name="xyz", fqn="docker.io/xyz", tag="latest"),
        Image(name="abc", fqn="docker.io/abc", tag=None),
    ],
)

IMAGE_SPEC = os.path.join(os.path.dirname(os.path.realpath(__file__)), "imageSpec.yaml")


@mock.patch("flytekit.configuration.default_images.DefaultImages.default_image")
@pytest.mark.parametrize(
    "image_string, leaf_configuration_file_name, final_image_config",
    [
        ("ghcr.io/flyteorg/mydefault:py3.9-latest", "no_images.yaml", ic_result_1),
        ("asdf=ghcr.io/asdf/asdf:latest", "sample.yaml", ic_result_2),
        ("xyz=ghcr.io/asdf/asdf:latest", "sample.yaml", ic_result_3),
        (IMAGE_SPEC, "sample.yaml", ic_result_4),
    ],
)
@pytest.mark.skipif(
    os.environ.get("GITHUB_ACTIONS") == "true" and sys.platform == "darwin",
    reason="Github macos-latest image does not have docker installed as per https://github.com/orgs/community/discussions/25777",
)
def test_pyflyte_run_run(
    mock_image, image_string, leaf_configuration_file_name, final_image_config, mock_image_spec_builder
):
    mock_image.return_value = "cr.flyte.org/flyteorg/flytekit:py3.9-latest"
    ImageBuildEngine.register("test", mock_image_spec_builder)

    @task
    def tk():
        ...

    mock_click_ctx = mock.MagicMock()
    mock_remote = mock.MagicMock()
    image_tuple = (image_string,)
    image_config = ImageConfig.validate_image(None, "", image_tuple)

    pp = pathlib.Path.joinpath(
        pathlib.Path(__file__).parent.parent.parent, "configuration/configs/", leaf_configuration_file_name
    )

    obj = RunLevelParams(
        project="p",
        domain="d",
        image_config=image_config,
        remote=True,
        config_file=str(pp),
    )
    obj._remote = mock_remote
    mock_click_ctx.obj = obj

    def check_image(*args, **kwargs):
        assert kwargs["image_config"] == final_image_config

    mock_remote.register_script.side_effect = check_image

    run_command(mock_click_ctx, tk)()


def test_file_param():
    m = mock.MagicMock()
    flyte_file = FileParamType().convert(__file__, m, m)
    assert flyte_file.path == __file__
    flyte_file = FileParamType().convert("https://tmp/file", m, m)
    assert flyte_file.path == "https://tmp/file"


def test_dir_param():
    m = mock.MagicMock()
    flyte_file = DirParamType().convert(DIR_NAME, m, m)
    assert flyte_file.path == DIR_NAME


class Color(Enum):
    RED = "red"
    GREEN = "green"
    BLUE = "blue"


@pytest.mark.parametrize("a_val", ["foo", "1", None])
def test_pyflyte_run_with_none(a_val):
    runner = CliRunner()
    args = [
        "run",
        WORKFLOW_FILE,
        "wf_with_none",
    ]
    if a_val is not None:
        args.extend(["--a", a_val])
    result = runner.invoke(
        pyflyte.main,
        args,
        catch_exceptions=False,
    )
    output = result.stdout.strip().split("\n")[-1].strip()
    if a_val is None:
        assert output == "default"
    else:
        assert output == a_val
    assert result.exit_code == 0


@pytest.mark.parametrize(
    "envs, envs_argument, expected_output",
    [
        (["--env", "MY_ENV_VAR=hello"], '["MY_ENV_VAR"]', "hello"),
        (["--env", "MY_ENV_VAR=hello", "--env", "ABC=42"], '["MY_ENV_VAR","ABC"]', "hello,42"),
    ],
)
def test_envvar_local_execution(envs, envs_argument, expected_output):
    runner = CliRunner()
    args = (
        [
            "run",
        ]
        + envs
        + [
            WORKFLOW_FILE,
            "wf_with_env_vars",
            "--env_vars",
        ]
        + [envs_argument]
    )
    result = runner.invoke(
        pyflyte.main,
        args,
        catch_exceptions=False,
    )
    output = result.stdout.strip().split("\n")[-1].strip()
    assert output == expected_output
