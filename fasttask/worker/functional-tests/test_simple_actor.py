import re
from datetime import timedelta
from functools import partial
from subprocess import run

from flytekit import WorkflowExecutionPhase

_run = partial(run, capture_output=True, text=True, check=True)


def test_simple_actor(workflows_dir, remote, config_path):
    """Check simple actor example.

    1. Run simple_actor.py
    2. Checks output is 15.
    """
    result = _run(
        [
            "union",
            "--config",
            config_path,
            "run",
            "--remote",
            "simple_actor.py",
            "add_five",
            "--num",
            "10",
        ],
        cwd=workflows_dir,
    )
    assert result.returncode == 0, result.stderr

    match = re.search(r"executions/(\w+)", result.stdout)

    execution_id = match.group(1)
    ex1 = remote.fetch_execution(name=execution_id)
    ex1 = remote.wait(ex1, poll_interval=timedelta(seconds=1))
    assert ex1.closure.phase == WorkflowExecutionPhase.SUCCEEDED
    assert ex1.outputs["o0"] == 15
