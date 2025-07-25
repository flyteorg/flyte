import re
from datetime import timedelta
from functools import partial
from subprocess import run

from flytekit import WorkflowExecutionPhase

_run = partial(run, capture_output=True, text=True, check=True)


def test_onboarding_wf(workflows_dir, remote, config_path):
    """Check onboarding wf does not error."""
    result = _run(
        [
            "union",
            "--config",
            config_path,
            "run",
            "--remote",
            "onboarding_workflows.py",
            "approve_pending_users",
        ],
        cwd=workflows_dir / "onboarding",
    )
    assert result.returncode == 0, result.stderr

    match = re.search(r"executions/(\w+)", result.stdout)

    execution_id = match.group(1)
    ex1 = remote.fetch_execution(name=execution_id)
    ex1 = remote.wait(ex1, poll_interval=timedelta(seconds=1))
    assert ex1.closure.phase == WorkflowExecutionPhase.SUCCEEDED
