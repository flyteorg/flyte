import os
from unittest.mock import Mock

import pytest

from flytekit.core import context_manager
from flytekit.core.context_manager import ExecutionState
from flytekit.image_spec import ImageSpec
from flytekit.image_spec.image_spec import _F_IMG_ID, ImageBuildEngine, calculate_hash_from_image_spec

REQUIREMENT_FILE = os.path.join(os.path.dirname(os.path.realpath(__file__)), "requirements.txt")
REGISTRY_CONFIG_FILE = os.path.join(os.path.dirname(os.path.realpath(__file__)), "registry_config.json")


def test_image_spec(mock_image_spec_builder):
    image_spec = ImageSpec(
        name="FLYTEKIT",
        builder="dummy",
        packages=["pandas"],
        apt_packages=["git"],
        python_version="3.8",
        registry="",
        base_image="cr.flyte.org/flyteorg/flytekit:py3.8-latest",
        cuda="11.2.2",
        cudnn="8",
        requirements=REQUIREMENT_FILE,
        registry_config=REGISTRY_CONFIG_FILE,
    )

    image_spec = image_spec.with_commands("echo hello")
    image_spec = image_spec.with_packages("numpy")
    image_spec = image_spec.with_apt_packages("wget")

    assert image_spec.python_version == "3.8"
    assert image_spec.base_image == "cr.flyte.org/flyteorg/flytekit:py3.8-latest"
    assert image_spec.packages == ["pandas", "numpy"]
    assert image_spec.apt_packages == ["git", "wget"]
    assert image_spec.registry == ""
    assert image_spec.requirements == REQUIREMENT_FILE
    assert image_spec.registry_config == REGISTRY_CONFIG_FILE
    assert image_spec.cuda == "11.2.2"
    assert image_spec.cudnn == "8"
    assert image_spec.name == "flytekit"
    assert image_spec.builder == "dummy"
    assert image_spec.source_root is None
    assert image_spec.env is None
    assert image_spec.pip_index is None
    assert image_spec.is_container() is True
    assert image_spec.commands == ["echo hello"]

    tag = calculate_hash_from_image_spec(image_spec)
    assert image_spec.image_name() == f"flytekit:{tag}"
    ctx = context_manager.FlyteContext.current_context()
    with context_manager.FlyteContextManager.with_context(
        ctx.with_execution_state(ctx.execution_state.with_params(mode=ExecutionState.Mode.TASK_EXECUTION))
    ):
        os.environ[_F_IMG_ID] = "flytekit:123"
        assert image_spec.is_container() is False

    ImageBuildEngine.register("dummy", mock_image_spec_builder)
    ImageBuildEngine.build(image_spec)

    assert "dummy" in ImageBuildEngine._REGISTRY
    assert calculate_hash_from_image_spec(image_spec) == tag
    assert image_spec.exist() is False

    # Remove the dummy builder, and build the image again
    # The image has already been built, so it shouldn't fail.
    del ImageBuildEngine._REGISTRY["dummy"]
    ImageBuildEngine.build(image_spec)

    with pytest.raises(Exception):
        image_spec.builder = "flyte"
        ImageBuildEngine.build(image_spec)

    # ImageSpec should be immutable
    image_spec.with_commands("ls")
    assert image_spec.commands == ["echo hello"]


def test_image_spec_engine_priority():
    image_spec = ImageSpec(name="FLYTEKIT")
    image_name = image_spec.image_name()

    new_image_name = f"fqn.xyz/{image_name}"
    mock_image_builder_10 = Mock()
    mock_image_builder_10.build_image.return_value = new_image_name
    mock_image_builder_default = Mock()
    mock_image_builder_default.build_image.side_effect = ValueError("should not be called")

    ImageBuildEngine.register("build_10", mock_image_builder_10, priority=10)
    ImageBuildEngine.register("build_default", mock_image_builder_default)

    ImageBuildEngine.build(image_spec)
    mock_image_builder_10.build_image.assert_called_once_with(image_spec)

    assert image_spec.image_name() == new_image_name
    del ImageBuildEngine._REGISTRY["build_10"]
    del ImageBuildEngine._REGISTRY["build_default"]
