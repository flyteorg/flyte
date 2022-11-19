"""
.. _subworkflows:

SubWorkflows
------------

.. tags:: Intermediate

Subworkflows are similar to :ref:`launch plans <Launch plans>` since they allow users to kick off one workflow from within another.

What's the Difference?
Consider launch plans as pass by pointer and subworkflows as pass by value.

.. note::

    Flyte's handling of dynamic workflows necessitates the use of subworkflows.
    We provide this capability at the user level rather than hiding it. The pros and cons of
    using subworkflows are discussed below.

When Should I Use SubWorkflows?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Subworkflows provide a clean solution to control parallelism between a workflow and its launched sub-flows 
because they execute within the same context as the parent workflow. 
Thus, all nodes of a subworkflow are bound by the total constraint on the parent workflow.

Consider this: When Workflow A is included as a subworkflow of Workflow B, and when Workflow B is run, the entire graph of workflow A is
copied into workflow B at the point where it is invoked.

Let's look at an example of subworkflow.
"""

# %%
# Example
# ^^^^^^^^
# Import the required dependencies into the environment.
import typing
from typing import Tuple

from flytekit import task, workflow

# %%
# Next, define a task that uses named outputs.
# As a best practice, usually try and define ``NamedTuple`` as a distinct type (although it can be defined inline).
op = typing.NamedTuple("OutputsBC", t1_int_output=int, c=str)


@task
def t1(a: int) -> op:
    return op(a + 2, "world")


# %%
# Then define a subworkflow like a typical workflow.
@workflow
def my_subwf(a: int = 42) -> Tuple[str, str]:
    x, y = t1(a=a)
    u, v = t1(a=x)
    return y, v


# %%
# We call the above-mentioned workflow above in a `parent` workflow below
# which demonstrates how to override the node name of a task (or subworkflow in this case).
#
# Nodes are typically named sequentially: ``n0``, ``n1``, and so on. Since the inner ``my_subwf`` also has a ``n0``, you might
# want to modify the first node's name. Because node IDs must be different within a workflow graph, 
# Flyte automatically prepends an attribute to the inner ``n0``.
@workflow
def parent_wf(a: int) -> Tuple[int, str, str]:
    x, y = t1(a=a).with_overrides(node_name="node-t1-parent")
    u, v = my_subwf(a=x)
    return x, u, v


# %%
# .. note::
#      For improved presentation or readability, the ``with_overrides`` method provides a new name to the graph-node.

# %%
# You can run the subworkflows locally.
if __name__ == "__main__":
    print(f"Running parent_wf(a=3) {parent_wf(a=3)}")


# Interestingly, we can nest a workflow that has a subworkflow within a workflow.
# Workflows can be simply composed from other workflows, even if they are standalone entities. Each of the
# workflows in this module can exist and run independently.
@workflow
def nested_parent_wf(a: int) -> Tuple[int, str, str, str]:
    x, y = my_subwf(a=a)
    m, n, o = parent_wf(a=a)
    return m, n, o, y


# %%
# You can run the nested workflows locally as well.
if __name__ == "__main__":
    print(f"Running nested_parent_wf(a=3) {nested_parent_wf(a=3)}")

# %%
# .. note:: You can :ref:`chain and execute subworkflows <Chain SubWorkflows>` similar to chained flyte tasks.


# %%
# External Workflow
# ^^^^^^^^^^^^^^^^^^
#
# When launch plans are used within a workflow to launch the execution of a previously defined workflow, a new
# external execution is launched, with a separate execution ID and can be observed as a distinct entity in
# FlyteConsole/Flytectl.
#
# They may have separate parallelism constraints since the context is not shared.
# We refer to such external invocations of a workflow using launch plans from a parent workflow as ``External Workflows``.
#
# .. tip::
#
#    If your deployment uses :ref:`Using Multiple Kubernetes Clusters <flyte:multicluster-setup>`, then external workflows may allow you to distribute the workload of a workflow to multiple clusters.
#
# Here's an example demonstrating external workflows:

# %%
# Import the required dependencies into the environment.
import typing  # noqa: E402
from collections import Counter  # noqa: E402
from typing import Dict, Tuple  # noqa: E402

from flytekit import LaunchPlan, task, workflow  # noqa: E402

# %%
# Define a task that computes the frequency of each word in a string, and returns a dictionary mapping every word to its count.
@task
def count_freq_words(input_string1: str) -> Dict:
    # input_string = "The cat sat on the mat"
    words = input_string1.split()
    wordCount = dict(Counter(words))
    return wordCount


# %%
# Construct a workflow that executes the previously-defined task.
@workflow
def ext_workflow(my_input: str) -> Dict:
    result = count_freq_words(input_string1=my_input)
    return result


# %%
# Next, create a launch plan.
external_lp = LaunchPlan.get_or_create(
    ext_workflow,
    "parent_workflow_execution",
)

# %%
# Define another task that returns the repeated keys (in our case, words) from a dictionary.
@task
def count_repetitive_words(word_counter: Dict) -> typing.List[str]:
    repeated_words = [key for key, value in word_counter.items() if value > 1]
    return repeated_words


# %%
# Define a workflow that triggers the launch plan of the previously-defined workflow.
@workflow
def parent_workflow(my_input1: str) -> typing.List[str]:
    my_op1 = external_lp(my_input=my_input1)
    my_op2 = count_repetitive_words(word_counter=my_op1)
    return my_op2


# %%
# Here, ``parent_workflow`` is an external workflow. This can also be run locally.
if __name__ == "__main__":
    print("Running parent workflow...")
    print(parent_workflow(my_input1="the cat took the apple and ate the apple"))

