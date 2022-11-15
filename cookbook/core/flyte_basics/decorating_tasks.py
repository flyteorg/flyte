"""
Decorating Tasks
----------------

A simple way of modifying the behavior of tasks is by using decorators to wrap your task functions.

In order to make sure that your decorated function contains all the type annotation and docstring
information that Flyte needs, you'll need to use the built-in :py:func:`~functools.wraps` decorator.

"""

import logging
from functools import partial, wraps

from flytekit import task, workflow

logger = logging.getLogger(__file__)


# %%
#
# Using a Single Decorator
# ^^^^^^^^^^^^^^^^^^^^^^^^
#
# Here we define decorator that logs the input and output information of a decorated task.


def log_io(fn):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        logger.info(f"task {fn.__name__} called with args: {args}, kwargs: {kwargs}")
        out = fn(*args, **kwargs)
        logger.info(f"task {fn.__name__} output: {out}")
        return out

    return wrapper


# %%
# Next, we define a task called ``t1`` which is decorated with ``log_io``.
#
# .. note::
#    The order of invoking the decorators is important. ``@task`` should always be the outer-most decorator.


@task
@log_io
def t1(x: int) -> int:
    return x + 1


# %%
#
# Stacking Multiple Decorators
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# You can also stack multiple decorators on top of each other as long as ``@task`` is the outer-most decorator.
#
# Here we define a decorator that validates that the output of the decorated function is a positive number before
# returning it, raising a ``ValueError`` if it violates this assumption.
#
# .. note::
#    The ``validate_output`` output uses :py:func:`~functools.partial` to implement parameterized decorators.


def validate_output(fn=None, *, floor=0):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        out = fn(*args, **kwargs)
        if out <= floor:
            raise ValueError(
                f"output of task {fn.__name__} must be a positive number, found {out}"
            )
        return out

    if fn is None:
        return partial(validate_output, floor=floor)

    return wrapper


# %%
# Now let's define a function that uses both the logging and validator decorators:


@task
@log_io
@validate_output(floor=10)
def t2(x: int) -> int:
    return x + 10


# %%
# Finally, we compose a workflow that calls ``t1`` and ``t2``.


@workflow
def wf(x: int) -> int:
    return t2(x=t1(x=x))


if __name__ == "__main__":
    print(f"Running wf(x=10) {wf(x=10)}")


# %%
# In this example, you learned how to modify the behavior of tasks via function decorators using the built-in
# :py:func:`~functools.wraps` decorator pattern. To learn more about how to extend Flyte at a deeper level, for
# example creating custom types, custom tasks, or backend plugins,
# see :ref:`Extending Flyte <plugins_extend>`.
