"""
Workflows
----------

Once you've had a handle on tasks, we can move to workflows. Workflow are the other basic building block of Flyte.

Workflows string together two or more tasks. They are also written as Python functions, but it is important to make a
critical distinction between tasks and workflows.

The body of a task's function runs at "run time", i.e. on the K8s cluster, using the task's container. The body of a
workflow is not used for computation, it is only used to structure the tasks, i.e. the output of ``t1`` is an input
of ``t2`` in the workflow below. As such, the body of workflows is run at "registration" time. Please refer to the
registration docs for additional information as well since it is actually a two-step process.

Take a look at the conceptual :std:ref:`discussion <flyte:divedeep-workflow-nodes>`.
behind workflows for additional information.

"""
import typing

from flytekit import task, workflow


@task
def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
    return a + 2, "world"


@task
def t2(a: str, b: str) -> str:
    return b + a


# %%
# You can treat the outputs of a task as you normally would a Python function. Assign the output to two variables
# and use them in subsequent tasks as normal. See :py:func:`flytekit.workflow`
@workflow
def my_wf(a: int, b: str) -> (int, str):
    x, y = t1(a=a)
    d = t2(a=y, b=b)
    return x, d


# %%
# Execute the Workflow, simply by invoking it like a function and passing in
# the necessary parameters
#
# .. note::
#
#   One thing to remember, currently we only support ``Keyword arguments``. So
#   every argument should be passed in the form ``arg=value``. Failure to do so
#   will result in an error
if __name__ == "__main__":
    print(f"Running my_wf(a=50, b='hello') {my_wf(a=50, b='hello')}")
