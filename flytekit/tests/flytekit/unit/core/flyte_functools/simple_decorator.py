"""Script used for testing local execution of functool.wraps-wrapped tasks"""
from __future__ import annotations

import os
from functools import wraps

from flytekit import task, workflow


def task_decorator(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        print(f"before running {fn.__name__}")
        try:
            print(f"try running {fn.__name__}")
            out = fn(*args, **kwargs)
        except Exception as exc:
            print(f"error running {fn.__name__}: {exc}")
        finally:
            print(f"finally after running {fn.__name__}")
        print(f"after running {fn.__name__}")
        return out

    return wrapper


@task
@task_decorator
def my_task(x: int) -> int:
    assert x > 0, f"my_task failed with input: {x}"
    print("running my_task")
    return x + 1


@workflow
def my_workflow(x: int) -> int:
    return my_task(x=x)


if __name__ == "__main__":
    print(my_workflow(x=int(os.getenv("SCRIPT_INPUT", 0))))
