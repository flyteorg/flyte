"""
04: Write a dynamic task
------------------------------------------

Please make sure you understand the difference between a task and a workflow - feel free to revisit the earlier
discussions and examples of those concepts. The key thing to note is that the body of tasks run at run time and the
body of workflows run at compile (aka registration) time. A dynamic task is a mix between the two - it's effectively
a workflow that is compiled at run time.

"""
import typing

from flytekit import dynamic, task, workflow


@task
def t1(a: int) -> str:
    a = a + 2
    return "world-" + str(a)


@task
def t2(a: str, b: str) -> str:
    return b + a


# %%
# This is how you declare a dynamic task. Note that you can call ``range`` on the input ``a``. In a normal workflow,
# you cannot do this and compilation would fail very early on. Because this isn't run until there are literal values
# for all inputs, it does work.
#
# Local execution will also work as it's treated like a task that will be run with Python native inputs.
#
# The name of this function was intentionally chosen. Even though these are implemented as tasks in flytekit, you
# should think about them as workflows. At execution time, flytekit is running the compilation step, and producing
# a ``WorkflowTemplate``, which it then passes back to Flyte Propeller for further running, much like how sub-workflows
# are handled.
@dynamic
def my_subwf(a: int) -> (typing.List[str], int):
    s = []
    for i in range(a):
        s.append(t1(a=i))
    return s, 5


@workflow
def my_wf(a: int, b: str) -> (str, typing.List[str], int):
    x = t2(a=b, b=b)
    v, z = my_subwf(a=a)
    return x, v, z


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running sub-wf directly my_subwf(a=3) = {my_subwf(a=3)}")
    print(f"Running my_wf(a=5, b='hello') {my_wf(a=5, b='hello')}")
