import base64
import os
from datetime import datetime
from pathlib import Path
from unittest.mock import Mock

import mock
import py
import pytest

import flytekit.configuration.plugin
from flytekit.configuration import (
    SERIALIZED_CONTEXT_ENV_VAR,
    FastSerializationSettings,
    Image,
    ImageConfig,
    SecretsConfig,
    SerializationSettings,
)
from flytekit.core import mock_stats
from flytekit.core.context_manager import ExecutionParameters, FlyteContext, FlyteContextManager, SecretsManager
from flytekit.models.core import identifier as id_models


class SampleTestClass(object):
    def __init__(self, value):
        self.value = value


def test_levels():
    ctx = FlyteContextManager.current_context()
    b = ctx.new_builder()
    b.flyte_client = SampleTestClass(value=1)
    with FlyteContextManager.with_context(b) as outer:
        assert outer.flyte_client.value == 1
        b = outer.new_builder()
        b.flyte_client = SampleTestClass(value=2)
        with FlyteContextManager.with_context(b) as ctx:
            assert ctx.flyte_client.value == 2

        with FlyteContextManager.with_context(outer.with_new_compilation_state()) as ctx:
            assert ctx.flyte_client.value == 1


def test_default():
    ctx = FlyteContext.current_context()
    assert ctx.file_access is not None


def test_look_up_image_info():
    img = Image.look_up_image_info(name="x", tag="docker.io/xyz", optional_tag=True)
    assert img.name == "x"
    assert img.tag is None
    assert img.fqn == "docker.io/xyz"

    img = Image.look_up_image_info(name="x", tag="docker.io/xyz:latest", optional_tag=True)
    assert img.name == "x"
    assert img.tag == "latest"
    assert img.fqn == "docker.io/xyz"

    img = Image.look_up_image_info(name="x", tag="docker.io/xyz:latest", optional_tag=False)
    assert img.name == "x"
    assert img.tag == "latest"
    assert img.fqn == "docker.io/xyz"

    img = Image.look_up_image_info(name="x", tag="localhost:5000/xyz:latest", optional_tag=False)
    assert img.name == "x"
    assert img.tag == "latest"
    assert img.fqn == "localhost:5000/xyz"


@mock.patch("flytekit.configuration.default_images.DefaultImages.default_image")
def test_validate_image(mock_image):
    mock_image.return_value = "cr.flyte.org/flyteorg/flytekit:py3.9-latest"
    ic = ImageConfig.validate_image(None, "image", ())
    assert ic
    assert ic.default_image == Image(name="default", fqn="cr.flyte.org/flyteorg/flytekit", tag="py3.9-latest")

    img1 = "xyz:latest"
    img2 = "docker.io/xyz:latest"
    img3 = "docker.io/xyz:latest"
    img3_cli = f"default={img3}"
    img4 = "docker.io/my:azb"
    img4_cli = f"my_img={img4}"

    ic = ImageConfig.validate_image(None, "image", (img1,))
    assert ic
    assert ic.default_image.full == img1

    ic = ImageConfig.validate_image(None, "image", (img2,))
    assert ic
    assert ic.default_image.full == img2

    ic = ImageConfig.validate_image(None, "image", (img3_cli,))
    assert ic
    assert ic.default_image.full == img3

    with pytest.raises(ValueError):
        ImageConfig.validate_image(None, "image", (img1, img3_cli))

    with pytest.raises(ValueError):
        ImageConfig.validate_image(None, "image", (img1, img2))

    with pytest.raises(ValueError):
        ImageConfig.validate_image(None, "image", (img1, img1))

    ic = ImageConfig.validate_image(None, "image", (img3_cli, img4_cli))
    assert ic
    assert ic.default_image.full == img3
    assert len(ic.images) == 2
    assert ic.images[1].full == img4


def test_secrets_manager_default():
    with pytest.raises(ValueError):
        sec = SecretsManager()
        sec.get("group", "key2")

    with pytest.raises(ValueError):
        _ = sec.group.key2


def test_secrets_manager_get_envvar():
    sec = SecretsManager()
    with pytest.raises(ValueError):
        sec.get_secrets_env_var("", "x")
    cfg = SecretsConfig.auto()
    assert sec.get_secrets_env_var("group", "test") == f"{cfg.env_prefix}GROUP_TEST"
    assert sec.get_secrets_env_var("group", "test", "v1") == f"{cfg.env_prefix}GROUP_V1_TEST"
    assert sec.get_secrets_env_var("group", group_version="v1") == f"{cfg.env_prefix}GROUP_V1"
    assert sec.get_secrets_env_var("group") == f"{cfg.env_prefix}GROUP"


def test_secret_manager_no_group(monkeypatch):
    plugin_mock = Mock()
    plugin_mock.secret_requires_group.return_value = False
    mock_global_plugin = {"plugin": plugin_mock}
    monkeypatch.setattr(flytekit.configuration.plugin, "_GLOBAL_CONFIG", mock_global_plugin)

    sec = SecretsManager()
    cfg = SecretsConfig.auto()
    sec.check_group_key(None)
    sec.check_group_key("")

    assert sec.get_secrets_env_var(key="ABC") == f"{cfg.env_prefix}ABC"

    default_path = Path(cfg.default_dir)
    expected_path = default_path / f"{cfg.file_prefix}abc"
    assert sec.get_secrets_file(key="ABC") == str(expected_path)


def test_secrets_manager_get_file():
    sec = SecretsManager()
    with pytest.raises(ValueError):
        sec.get_secrets_file("", "x")
    cfg = SecretsConfig.auto()
    assert sec.get_secrets_file("group", "test") == os.path.join(
        cfg.default_dir,
        "group",
        f"{cfg.file_prefix}test",
    )
    assert sec.get_secrets_file("group", "test", "v1") == os.path.join(
        cfg.default_dir,
        "group",
        "v1",
        f"{cfg.file_prefix}test",
    )


def test_secrets_manager_file(tmpdir: py.path.local):
    tmp = tmpdir.mkdir("file_test").dirname
    os.environ["FLYTE_SECRETS_DEFAULT_DIR"] = tmp
    sec = SecretsManager()
    f = os.path.join(tmp, "test")
    with open(f, "w+") as w:
        w.write("my-password")

    with pytest.raises(ValueError):
        sec.get("", "x")
    # Group dir not exists
    with pytest.raises(ValueError):
        sec.get("group", "test")

    g = os.path.join(tmp, "group")
    os.makedirs(g)
    f = os.path.join(g, "test")
    with open(f, "w+") as w:
        w.write("my-password")
    assert sec.get("group", "test") == "my-password"
    assert sec.group.test == "my-password"

    base64_string = "R2Vla3NGb3JHZWV =="
    base64_bytes = base64_string.encode("ascii")
    base64_str = base64.b64encode(base64_bytes)
    with open(f, "wb") as w:
        w.write(base64_str)
    assert sec.get("group", "test") != base64_str
    assert sec.get("group", "test", encode_mode="rb") == base64_str

    del os.environ["FLYTE_SECRETS_DEFAULT_DIR"]


def test_secrets_manager_bad_env():
    with pytest.raises(ValueError):
        os.environ["TEST"] = "value"
        sec = SecretsManager()
        sec.get("group", "test")


def test_secrets_manager_env():
    sec = SecretsManager()
    os.environ[sec.get_secrets_env_var("group", "test")] = "value"
    assert sec.get("group", "test") == "value"

    os.environ[sec.get_secrets_env_var(group="group", key="key")] = "value"
    assert sec.get(group="group", key="key") == "value"


def test_serialization_settings_transport():
    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"hello": "blah"},
        image_config=ImageConfig(
            default_image=default_img,
            images=[default_img],
        ),
        flytekit_virtualenv_root="/opt/venv/blah",
        python_interpreter="/opt/venv/bin/python3",
        fast_serialization_settings=FastSerializationSettings(
            enabled=True,
            destination_dir="/opt/blah/blah/blah",
            distribution_location="s3://my-special-bucket/blah/bha/asdasdasd/cbvsdsdf/asdddasdasdasdasdasdasd.tar.gz",
        ),
    )

    tp = serialization_settings.serialized_context
    with_serialized = serialization_settings.with_serialized_context()
    assert serialization_settings.env == {"hello": "blah"}
    assert with_serialized.env
    assert with_serialized.env[SERIALIZED_CONTEXT_ENV_VAR] == tp
    ss = SerializationSettings.from_transport(tp)
    assert ss is not None
    assert ss == serialization_settings
    assert len(tp) == 400


def test_exec_params():
    ep = ExecutionParameters(
        execution_id=id_models.WorkflowExecutionIdentifier("p", "d", "n"),
        task_id=id_models.Identifier(id_models.ResourceType.TASK, "local", "local", "local", "local"),
        execution_date=datetime.utcnow(),
        stats=mock_stats.MockStats(),
        logging=None,
        tmp_dir="/tmp",
        raw_output_prefix="",
        decks=[],
    )

    assert ep.task_id.name == "local"
