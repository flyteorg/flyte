import os
from pathlib import Path

import pytest
from flytekit.configuration import Config
from flytekit.remote import FlyteRemote


@pytest.fixture(scope="session")
def remote() -> FlyteRemote:
    return FlyteRemote(
        Config.for_sandbox(),
        default_project="flytesnacks",
        default_domain="development",
    )


@pytest.fixture(scope="session")
def workflows_dir() -> Path:
    return Path(__file__).parent / "workflows"


@pytest.fixture(scope="session")
def config_path() -> str:
    sandbox_path = Path(__file__).parent / "sandbox-yaml.yaml"
    return os.fspath(sandbox_path.resolve())


@pytest.fixture(scope="session")
def union_image() -> str:
    image = os.getenv("UNION_RUNTIME_TEST_IMAGE")
    if image is None:
        pytest.fail("Please set UNION_RUNTIME_TEST_IMAGE")
    return image
