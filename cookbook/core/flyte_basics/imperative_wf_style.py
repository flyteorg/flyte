"""
.. _imperative_wf_style:

Imperative Workflows
--------------------

.. tags:: Basic

Workflows are typically created and specified by decorating a function with the ``@workflow`` decorator. This will
run through the body of the function at compile time, using the subsequent calls of the underlying tasks to determine
and record the workflow structure. This is the declarative style and makes sense when a human is writing it up by hand.

For cases where workflows are constructed programmatically, an imperative style makes more sense. For example, when tasks have already been defined, their order and dependencies have
been specified in text format of some kind (perhaps you're converting from a legacy system), and your goal is to orchestrate those tasks.

"""
import typing

from flytekit import Workflow, task


# %%
# Assume we have the following tasks, and they are meant to represent more complicated tasks. ``t1`` has simple scalar I/O, ``t2`` is
# a pure side effect task (though we typically don't recommend these, they are inevitable), and ``t3`` takes in a list
# as an input.
@task
def t1(a: str) -> str:
    return a + " world"


@task
def t2():
    print("side effect")


@task
def t3(a: typing.List[str]) -> str:
    """
    This is a pedagogical demo that happens to do a reduction step. Flyte is higher-order orchestration
    platform, not a map-reduce framework and is not meant to supplant Spark et. al.
    """
    return ",".join(a)


# %%
# Start by creating an imperative style workflow, which is aliased to just ``Workflow`` from ``flytekit``
wf = Workflow(name="my.imperative.workflow.example")


# %%
# Inputs have to be added to the workflow before they can be used. Add them by specifying the name and the type.
wf.add_workflow_input("in1", str)

# %%
# Next associate a task, and pass in the workflow level input.
node_t1 = wf.add_entity(t1, a=wf.inputs["in1"])

# %%
# Create a workflow output linked to the output of that task.
wf.add_workflow_output("output_from_t1", node_t1.outputs["o0"])

# %%
# To add a task that has no inputs or outputs, just add the entity. We don't need to capture the resulting node
# because we have no use for it.
wf.add_entity(t2)

# %%
# We can also pass in a list to a task. Also creating a workflow input returns an object that can be used as an
# alternate way of linking workflow inputs. Here, ``t3`` uses both workflow inputs.
wf_in2 = wf.add_workflow_input("in2", str)
node_t3 = wf.add_entity(t3, a=[wf.inputs["in1"], wf_in2])

# %%
# You can also create a workflow output as a list from multiple task outputs
wf.add_workflow_output(
    "output_list",
    [node_t1.outputs["o0"], node_t3.outputs["o0"]],
    python_type=typing.List[str],
)


if __name__ == "__main__":
    print(wf)
    print(wf(in1="hello", in2="foo"))
