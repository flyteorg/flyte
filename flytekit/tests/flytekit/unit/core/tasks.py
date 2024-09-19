"""
NOTE: This file exists to test failure in case of instantiating a task inside a function in a non test_* module.
Refer to test_python_function_task.py to look at the usage.
"""
from flytekit import task


def tasks():
    @task
    def foo():
        pass
