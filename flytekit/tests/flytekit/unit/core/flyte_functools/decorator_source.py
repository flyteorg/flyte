"""Script used for testing local execution of functool.wraps-wrapped tasks for stacked decorators"""
import functools
import typing
from functools import wraps
from typing import List


def task_setup(function: typing.Callable, *, integration_requests: typing.Optional[List] = None) -> typing.Callable:
    integration_requests = integration_requests or []

    @wraps(function)
    def wrapper(*args, **kwargs):
        # Preprocessing of task
        print("preprocessing")

        # Execute function
        output = function(*args, **kwargs)

        # Postprocessing of output
        print("postprocessing")

        return output

    return functools.partial(task_setup, integration_requests=integration_requests) if function is None else wrapper
