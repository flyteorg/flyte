"""
Workflows
---------

.. tags:: Basic

Once you have a handle on :ref:`tasks <basics_of_tasks>`, you can dive into Flyte workflows.
Together, Flyte tasks and workflows make up the fundamental building blocks of Flyte.

Workflows bind together two or more tasks. They can be written as Python functions, but it is essential to make a
critical distinction between tasks and workflows.

The body of a task runs at "run time", i.e., on a K8s cluster (using the task's container), in a Query
Engine like BigQuery, or some other hosted service like AWS Batch, Sagemaker, etc. The body of a
workflow is not used for computation; it is only used to structure tasks.
The body of a workflow runs at "registration" time, i.e., when the workflow unwraps during registration.
Registration refers to uploading the packaged (serialized) code to Flyte backend so that the workflow can be triggered.
Please refer to the :std:ref:`registration docs <flyte:divedeep-registration>` for more information.

Now, let's get started with a simple workflow.
"""
import typing
from typing import Tuple

from flytekit import task, workflow


@task
def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
    return a + 2, "world"


@task
def t2(a: str, b: str) -> str:
    return b + a


@workflow
def my_wf(a: int, b: str) -> Tuple[int, str]:
    x, y = t1(a=a)
    d = t2(a=y, b=b)
    return x, d


# %%
# ``my_wf`` is the Flyte workflow that accepts two inputs and calls two tasks from within it.
#
# How does a Flyte workflow work?
# ===============================
#
# The ``@workflow`` decorator wraps Python functions (Flyte tasks) which can be assumed as ``lazily`` evaluated promises. The functions are
# parsed, and each function call is deferred until the execution time. The function calls return "Promises", which can be
# passed downstream to other functions but cannot be accessed within the workflow.
# The actual ``evaluation`` (evaluation refers to the execution of a task within the workflow) happens when the
# workflow is executed.
#
# A workflow can be executed locally where the evaluation will happen immediately, or using the CLI, UI, etc., which will trigger an evaluation.
# Although Flyte workflows decorated with ``@workflow`` look like Python functions, they are actually python-esque, Domain Specific Language (DSL) entities
# that recognize the ``@task`` decorators. When a workflow encounters a ``@task``-decorated Python function, it creates a
# :py:class:`flytekit.core.promise.Promise` object. This promise doesn't contain the actual output of the task, and is only fulfilled at execution time.
#
# .. note::
#   Refer to :py:func:`flytekit.dynamic` to create Flyte workflows dynamically. In a dynamic workflow, unlike a simple workflow,
#   the actual inputs are pre-materialized; however, every invocation of a task still
#   results in a promise to be evaluated lazily.


# %%
# Now, we can execute the workflow by invoking it like a function and sending in the required inputs.
#
# .. note::
#
#   Currently we only support ``keyword arguments``, so
#   every argument should be passed in the form of ``arg=value``. Failure to do so
#   will result in an error.
if __name__ == "__main__":
    print(f"Running my_wf(a=50, b='hello') {my_wf(a=50, b='hello')}")

# %%
# To know more about workflows, take a look at the conceptual :std:ref:`discussion <flyte:divedeep-workflow-nodes>`.
