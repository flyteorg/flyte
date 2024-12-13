import os
import re
from datetime import timedelta
from functools import partial
from pathlib import Path
from random import choice
from string import Template, ascii_lowercase
from subprocess import run
from textwrap import dedent
from typing import Optional

from flytekit import WorkflowExecutionPhase
from flytekit.remote import FlyteRemote

_run = partial(run, capture_output=True, text=True)

TASK_TEMPLATE = Template("""\
from flytekit import task
from union.actor import ActorEnvironment

actor_env = ActorEnvironment(
    name="$RANDOM_NAME",
    container_image="$CONTAINER_IMAGE",
    ttl_seconds=100,
)

$TASK_BODY
""")

TASK_TEST_CASES = [
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "one"
            return my_var
        """),
        WorkflowExecutionPhase.SUCCEEDED,
        "one",
    ),
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "two"
            return my_var
        """),
        WorkflowExecutionPhase.SUCCEEDED,
        "two",
    ),
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "two"
            raise ValueError("This is bad")
            return my_var
        """),
        WorkflowExecutionPhase.FAILED,
        None,
    ),
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "five"
            return my_var
        """),
        WorkflowExecutionPhase.SUCCEEDED,
        "five",
    ),
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "one"
            return my_var
        """),
        WorkflowExecutionPhase.SUCCEEDED,
        "one",
    ),
    (
        dedent("""\
        @actor_env.task
        def check_out() -> str:
            my_var = "three"
            raise ValueError("This is bad")
            return my_var
        """),
        WorkflowExecutionPhase.FAILED,
        None,
    ),
]


def test_fast_register_actor(config_path, union_image, remote, tmp_path):
    """Check fast-register works with actors.

    1. Run multiple actor tasks sequentially while updating the body of a task.
    2. Check for the expected the phase and the output.
    """
    workflows_dir = tmp_path / "workflows"
    workflows_dir.mkdir()
    task_file = workflows_dir / "main.py"
    random_name = "".join(choice(ascii_lowercase) for _ in range(20))

    for idx, (task_body, phase, expected_output) in enumerate(TASK_TEST_CASES):
        task_code = TASK_TEMPLATE.substitute(
            CONTAINER_IMAGE=union_image,
            TASK_BODY=task_body,
            RANDOM_NAME=random_name,
        )
        task_file.write_text(task_code)
        _run_and_check(config_path, workflows_dir, remote, phase, idx, task_code, expected_output)


def _run_and_check(
    config_path: Path,
    workflows_dir: Path,
    remote: FlyteRemote,
    phase: WorkflowExecutionPhase,
    idx: int,
    task_code: str,
    expected_output: Optional[str],
):
    result = _run(
        [
            "union",
            "--config",
            os.fspath(config_path),
            "run",
            "--remote",
            "main.py",
            "check_out",
        ],
        cwd=workflows_dir,
    )
    assert result.returncode == 0, result.stderr

    match = re.search(r"executions/(\w+)", result.stdout)

    execution_id = match.group(1)
    ex1 = remote.fetch_execution(name=execution_id)
    ex1 = remote.wait(ex1, poll_interval=timedelta(seconds=1))
    assert ex1.closure.phase == phase, (idx, execution_id, task_code)
    if phase == WorkflowExecutionPhase.SUCCEEDED:
        assert ex1.outputs["o0"] == expected_output, (idx, execution_id, task_code)
