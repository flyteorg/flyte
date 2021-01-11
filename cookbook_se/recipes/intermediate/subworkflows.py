"""
Call another workflow by subworkflow
------------------------------------------

Launch plans were introduced in the Basics section of this book. Subworkflows are similar in that they allow users
to kick off one workflow from inside another. What's the difference? Think of launch plans as passing by pointer and
subworkflows as passing by value.

When you include a launch plan of workflow A inside workflow B, when B gets run, a new workflow execution,
replete with a new workflow execution ID, a new Flyte UI link, will be run.

When you include workflow A as a subworkflow of workflow B, when B gets run, the entire workflow A graph is basically
copied into workflow B at the point where it is called.
"""

import typing

from flytekit import task, workflow


@task
def t1(a: int) -> typing.NamedTuple("OutputsBC", t1_int_output=int, c=str):
    return a + 2, "world"


# %%
# This will be the subworkflow of our examples, but note that this is a workflow like any other. It can be run just
# like any other workflow. Note here that the workflow has been declared with a default.
@workflow
def my_subwf(a: int = 42) -> (str, str):
    x, y = t1(a=a)
    u, v = t1(a=x)
    return y, v


# %%
# This is the parent workflow. In it, we call the workflow declared above.
# This also showcases how to override the node name of a task (or subworkflow). Typically, nodes are just named
# sequentially, ``node-0``, ``node-1``, and so on. Because the inner ``my_subwf`` also has a ``node-0`` you may
# wish to change the name of the first one. Not doing so is also fine - Flyte will automatically prepend something
# to the inner ``node-0``, since node IDs need to be distinct within a workflow graph. This issue does not exist
# when calling something by launch plan since those launch a separate execution entirely.
@workflow
def parent_wf(a: int) -> (int, str, str):
    x, y = t1(a=a).with_overrides(node_name="node-t1-parent")
    u, v = my_subwf(a=x)
    return x, u, v


if __name__ == "__main__":
    print(f"Running parent_wf(a=3) {parent_wf(a=3)}")
