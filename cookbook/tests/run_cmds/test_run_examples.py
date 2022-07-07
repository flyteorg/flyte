import re
import os
import subprocess
import typing
from dataclasses import dataclass, field
from pathlib import Path

import pytest
from flytekit.configuration import Config
from flytekit.remote import FlyteRemote


N_SYNCS = 200


COOKBOOK_ROOT_DIR = Path(__file__).parent / ".." / ".."

FLYTECTL_CMD = "sandbox" if os.getenv("RUN_CMDS_CI", False) else "demo"
NO_CLUSTER_MSG = "🛑 no demo cluster found" if FLYTECTL_CMD == "demo" else "🛑 no Sandbox found"


@dataclass
class ExampleTestCase:
    id: str
    script_rel_path: Path
    expected_output: typing.Any
    root_dir: Path = field(init=False)
    pyflyte_run_args: typing.List[str] = field(init=False)
    flytekit_remote_args: typing.List[str] = field(init=False)

    def __post_init__(self):
        full_path = COOKBOOK_ROOT_DIR / self.script_rel_path
        self.root_dir = full_path.parent
        self.pyflyte_run_args = ["bash", f"_run/run_{full_path.stem}.sh"]
        self.flytekit_remote_args = ["python", "-m", f"_run.run_{full_path.stem}"]


@pytest.fixture(scope="session")
def flyte_remote():
    p_status = subprocess.run(["flytectl", FLYTECTL_CMD, "status"], capture_output=True)

    cluster_preexists = True
    if p_status.stdout.decode().strip() == NO_CLUSTER_MSG:
        # if a demo cluster didn't exist already, then start one.
        cluster_preexists = False
        subprocess.run(["flytectl", FLYTECTL_CMD, "start"])

    remote = FlyteRemote(
        config=Config.auto(),
        default_project="flytesnacks",
        default_domain="development",
    )
    projects, *_ = remote.client.list_projects_paginated(limit=5, token=None)
    assert "flytesnacks" in [p.id for p in projects]
    assert "flytesnacks" in [p.name for p in projects]

    yield remote

    if not cluster_preexists:
        # only teardown the demo cluster if it didn't preexist
        subprocess.run(["flytectl", FLYTECTL_CMD, "teardown"])


def get_execution(flyte_remote, execution_name):
    execution = flyte_remote.fetch_execution(name=execution_name)
    flyte_remote.wait(execution, sync_nodes=False)
    return flyte_remote.sync_execution(execution)


def get_execution_name_pyflyte(x):
    """
    Get the execution name from pyflyte run stdout, e.g.:

    "Go to localhost:30081/console/.../executions/asdf1234" -> "asdf1234"
    """
    return re.match(r".+\/executions\/(\S+) .+", x).group(1)


def get_execution_name_flytekit_remote(x):
    """
    Get the execution name from FlyteRemote python script example, e.g.:

    "Execution successfully started: asdf1234" -> "asdf1234"
    """
    return re.match(r".+: (\S+)$", x).group(1)


@pytest.mark.parametrize(
    "example_test_case",
    [
        ExampleTestCase("hello-world", "core/flyte_basics/hello_world.py", {"o0": "hello world"}),
        ExampleTestCase("task", "core/flyte_basics/task.py", {"o0": 16}),
        ExampleTestCase("basic-workflow", "core/flyte_basics/basic_workflow.py", {"o0": 102, "o1": "helloworld"}),
    ],
    ids=lambda x: x.id
)
@pytest.mark.parametrize(
    "run_type", ["pyflyte_run", "flytekit_remote"], ids=lambda x: x.replace("_", "-")
)
def test_example_suite(flyte_remote: FlyteRemote, example_test_case: ExampleTestCase, run_type: str):
    args = example_test_case.pyflyte_run_args
    get_execution_name = get_execution_name_pyflyte
    if run_type == "flytekit_remote":
        args = example_test_case.flytekit_remote_args
        get_execution_name = get_execution_name_flytekit_remote

    proc = subprocess.run(args, cwd=example_test_case.root_dir, capture_output=True)
    std_out = proc.stdout.decode()
    execution_name = get_execution_name(std_out.strip())
    execution = get_execution(flyte_remote, execution_name)
    assert execution.outputs == example_test_case.expected_output
