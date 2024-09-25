import re
from datetime import timedelta
from functools import partial
from subprocess import run

from flytekit import WorkflowExecutionPhase

_run = partial(run, capture_output=True, text=True, check=True)


def test_pod_template_actor(workflows_dir, remote, config_path):
    """Check pod template actor example.

    1. Run pod_template_actor.py
    2. Check output of pod template.
    """
    result = _run(
        [
            "union",
            "--config",
            config_path,
            "run",
            "--remote",
            "pod_template_actor.py",
            "wf",
        ],
        cwd=workflows_dir,
    )
    assert result.returncode == 0, result.stderr

    match = re.search(r"executions/(\w+)", result.stdout)

    execution_id = match.group(1)
    ex1 = remote.fetch_execution(name=execution_id)
    ex1 = remote.wait(ex1, poll_interval=timedelta(seconds=1))
    assert ex1.closure.phase == WorkflowExecutionPhase.SUCCEEDED
    assert "hello" in ex1.outputs["o0"]
