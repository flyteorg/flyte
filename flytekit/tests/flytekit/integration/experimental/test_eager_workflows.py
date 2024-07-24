"""Eager workflow integration tests.

These tests are currently not run in CI. In order to run this locally you'll need to start a
local flyte cluster, and build and push a flytekit development image:

```

# if you already have a local cluster running, tear it down and start fresh
flytectl demo teardown -v

# start a local flyte cluster
flytectl demo start

# build and push the image
docker build . -f Dockerfile.dev -t localhost:30000/flytekit:dev --build-arg PYTHON_VERSION=3.9
docker push localhost:30000/flytekit:dev

# run the tests
pytest tests/flytekit/integration/experimental/test_eager_workflows.py
```
"""

import asyncio
import os
import subprocess
import time
from pathlib import Path

import pytest

from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

from .eager_workflows import eager_wf_local_entrypoint

MODULE = "eager_workflows"
MODULE_PATH = Path(__file__).parent / f"{MODULE}.py"
CONFIG = os.environ.get("FLYTECTL_CONFIG", str(Path.home() / ".flyte" / "config-sandbox.yaml"))
IMAGE = os.environ.get("FLYTEKIT_IMAGE", "localhost:30000/flytekit:dev")


@pytest.fixture(scope="session")
def register():
    subprocess.run(
        [
            "pyflyte",
            "-c",
            CONFIG,
            "register",
            "--image",
            IMAGE,
            "--project",
            "flytesnacks",
            "--domain",
            "development",
            MODULE_PATH,
        ]
    )


@pytest.mark.skipif(
    os.environ.get("FLYTEKIT_CI", False), reason="Running workflows with sandbox cluster fails due to memory pressure"
)
@pytest.mark.parametrize(
    "entity_type, entity_name, input, output",
    [
        ("eager", "simple_eager_wf", 1, 4),
        ("eager", "conditional_eager_wf", 1, -1),
        ("eager", "conditional_eager_wf", -10, 1),
        ("eager", "try_except_eager_wf", 1, 1),
        ("eager", "try_except_eager_wf", 0, -1),
        ("eager", "gather_eager_wf", 1, [2] * 10),
        ("eager", "nested_eager_wf", 1, 8),
        ("eager", "eager_wf_with_subworkflow", 1, 4),
        ("eager", "eager_wf_structured_dataset", None, 6),
        ("eager", "eager_wf_flyte_file", None, "some data"),
        ("eager", "eager_wf_flyte_directory", None, "some data"),
        ("workflow", "wf_with_eager_wf", 1, 8),
    ],
)
def test_eager_workflows(register, entity_type, entity_name, input, output):
    remote = FlyteRemote(
        config=Config.auto(config_file=CONFIG),
        default_project="flytesnacks",
        default_domain="development",
    )

    fetch_method = {
        "eager": remote.fetch_task,
        "workflow": remote.fetch_workflow,
    }[entity_type]

    entity = None
    for i in range(100):
        try:
            entity = fetch_method(name=f"{MODULE}.{entity_name}")
            break
        except Exception:
            print(f"retry {i}")
            time.sleep(6)
            continue

    if entity is None:
        raise RuntimeError("failed to fetch entity")

    inputs = {} if input is None else {"x": input}
    execution = remote.execute(entity, inputs=inputs, wait=True)
    assert execution.outputs["o0"] == output


@pytest.mark.skipif(
    os.environ.get("FLYTEKIT_CI", False), reason="Running workflows with sandbox cluster fails due to memory pressure"
)
def test_eager_workflow_local_entrypoint(register):
    result = asyncio.run(eager_wf_local_entrypoint(x=1))
    assert result == 4
