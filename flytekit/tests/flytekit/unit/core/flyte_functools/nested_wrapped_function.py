"""Script used for testing local execution of functool.wraps-wrapped tasks"""

from functools import wraps

from flytekit import task, workflow


def task_decorator(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        return fn(*args, **kwargs)

    return wrapper


def foo():
    @task
    @task_decorator
    def my_task(x: int) -> int:
        assert x > 0, f"my_task failed with input: {x}"
        print("running my_task")
        return x + 1

    return my_task


my_task = foo()


@workflow
def my_workflow(x: int) -> int:
    return my_task(x=x)


if __name__ == "__main__":
    print(my_workflow(x=11))
