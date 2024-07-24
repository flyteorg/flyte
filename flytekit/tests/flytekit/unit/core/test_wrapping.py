from functools import wraps

from flytekit import task, workflow


def test_task_correctly_wrapped():
    @task
    def my_task(a: int) -> int:
        return a

    assert my_task.__wrapped__ == my_task._task_function


def test_stacked_decorators():
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
        """Some function doc"""
        print("running my_task")
        return x + 1

    assert my_task.__wrapped__.__doc__ == "Some function doc"
    assert my_task.__wrapped__ == my_task._task_function
    assert my_task(x=10) == 11


def test_wf_correctly_wrapped():
    @workflow
    def my_workflow(a: int) -> int:
        return a

    assert my_workflow.__wrapped__ == my_workflow._workflow_function
