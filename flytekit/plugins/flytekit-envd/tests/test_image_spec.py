from pathlib import Path
from textwrap import dedent

import pytest
from flytekitplugins.envd.image_builder import EnvdImageSpecBuilder, create_envd_config

from flytekit.image_spec.image_spec import ImageBuildEngine, ImageSpec


@pytest.fixture(scope="module", autouse=True)
def register_envd_higher_priority():
    # Register a new envd platform with the highest priority so the test in this file uses envd
    highest_priority_builder = max(ImageBuildEngine._REGISTRY, key=ImageBuildEngine._REGISTRY.get)
    highest_priority = ImageBuildEngine._REGISTRY[highest_priority_builder][1]
    yield ImageBuildEngine.register(
        "envd_high_priority",
        EnvdImageSpecBuilder(),
        priority=highest_priority + 1,
    )
    del ImageBuildEngine._REGISTRY["envd_high_priority"]


def test_image_spec():
    image_spec = ImageSpec(
        packages=["pandas"],
        apt_packages=["git"],
        python_version="3.8",
        base_image="cr.flyte.org/flyteorg/flytekit:py3.8-latest",
        pip_index="https://private-pip-index/simple",
    )

    image_spec = image_spec.with_commands("echo hello")

    EnvdImageSpecBuilder().build_image(image_spec)
    config_path = create_envd_config(image_spec)
    assert image_spec.platform == "linux/amd64"
    image_name = image_spec.image_name()
    contents = Path(config_path).read_text()
    assert (
        contents
        == f"""# syntax=v1

def build():
    base(image="cr.flyte.org/flyteorg/flytekit:py3.8-latest", dev=False)
    run(commands=["echo hello"])
    install.python_packages(name=["pandas"])
    install.apt_packages(name=["git"])
    runtime.environ(env={{'PYTHONPATH': '/root', '_F_IMG_ID': '{image_name}'}}, extra_path=['/root'])
    config.pip_index(url="https://private-pip-index/simple")
    install.python(version="3.8")
"""
    )


def test_image_spec_conda():
    image_spec = ImageSpec(
        base_image="ubuntu:20.04",
        python_version="3.11",
        packages=["flytekit"],
        conda_packages=["pytorch", "cpuonly"],
        conda_channels=["pytorch"],
    )

    EnvdImageSpecBuilder().build_image(image_spec)
    config_path = create_envd_config(image_spec)
    assert image_spec.platform == "linux/amd64"
    image_name = image_spec.image_name()
    contents = Path(config_path).read_text()
    expected_contents = dedent(
        f"""\
    # syntax=v1

    def build():
        base(image="ubuntu:20.04", dev=False)
        run(commands=[])
        install.python_packages(name=["flytekit"])
        install.apt_packages(name=[])
        runtime.environ(env={{'PYTHONPATH': '/root', '_F_IMG_ID': '{image_name}'}}, extra_path=['/root'])
        config.pip_index(url="https://pypi.org/simple")
        install.conda(use_mamba=True)
        install.conda_packages(name=["pytorch", "cpuonly"], channel=["pytorch"])
        install.python(version="3.11")
    """
    )

    assert contents == expected_contents
