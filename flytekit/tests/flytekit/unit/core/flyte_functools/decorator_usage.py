from flytekit import task

from .decorator_source import task_setup


@task
@task_setup
def get_data(x: int) -> int:
    return x + 1
