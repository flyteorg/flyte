from flytekitplugins.flyteinteractive import vscode

from flytekit import task


@task()
@vscode(run_task_first=True)
def t1(a: int, b: int) -> int:
    return a // b
