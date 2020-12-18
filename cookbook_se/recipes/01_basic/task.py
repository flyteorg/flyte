"""
01: Tasks
----------

This example shows how to write a task in flytekit python.
Recap: In Flyte a task is a fundamental building block and an extension point. Flyte has multiple plugins for tasks,
which can be either a backend-plugin or can be a simple extension that is available in flytekit.

A task in flytekit can be 2 types:

1. A task that has a python function associated with it. The execution of the task would be an execution of this
   function
#. A task that does not have a python function, for e.g a SQL query or some other portable task like Sagemaker prebuilt
   algorithms, or something that just invokes an API

This section will talk about how to write a Python Function task. Other type of tasks will be covered in later sections

"""
# %%
# For any task in flyte, there is always one required import
from flytekit import task


# %%
# A ``PythonFunctionTask`` must always be decorated with the ``@task`` ``flytekit.task`` decorator.
# The task itself is a regular python function, with one exception, it needs all the inputs and outputs to be clearly
# annotated with the types. The types are regular python types, more on this in the type-system section.
# :py:func:`flytekit.task`
@task
def square(n: int) -> int:
    """
     Parameters:
        n (int): name of the parameter for the task will be derived from the name of the input variable
               the type will be automatically deduced to be Types.Integer

    Return:
        int: The label for the output will be automatically assigned and type will be deduced from the annotation

    """
    return n * n


# %%
# In this task, one input is ``n`` which has type ``int``.
# the task ``square`` takes the number ``n`` and returns a new integer (squared value)
#
# .. note::
#
#   Flytekit will assign a default name to the output variable like ``out0``
#   In case of multiple outputs, each output will be numbered in the order
#   starting with 0. For e.g. -> ``out0, out1, out2, ...``
#
# Flyte tasks can be executed like normal functions
if __name__ == "__main__":
    print(square(n=10))
