"""Script used for testing local execution of functool.wraps-wrapped tasks for stacked decorators"""

import os
from functools import wraps

from flytekit import task, workflow


def task_decorator_1(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        print("running task_decorator_1")
        return fn(*args, **kwargs)

    return wrapper


def task_decorator_2(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        print("running task_decorator_2")
        return fn(*args, **kwargs)

    return wrapper


def task_decorator_3(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        print("running task_decorator_3")
        return fn(*args, **kwargs)

    return wrapper


@task
@task_decorator_1
@task_decorator_2
@task_decorator_3
def my_task(x: int) -> int:
    print("running my_task")
    return x + 1


@workflow
def my_workflow(x: int) -> int:
    return my_task(x=x)


if __name__ == "__main__":
    print(my_workflow(x=int(os.getenv("SCRIPT_INPUT", 0))))
