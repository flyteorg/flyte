"""
Write a map task
------------------------------------------

A map task lets you run a task over a list of inputs within a single workflow node. This means you can run thousands
of instances of that task without creating a node for every instance, providing valuable performance gains. Furthermore
you can run map tasks on alternate execution back-ends, such as `AWS Batch <https://aws.amazon.com/batch/>`__ which is
a provisioned services that can scale to great sizes.

"""

import typing

from flytekit import TaskMetadata, map_task, task, workflow


# %%
# Note that this is the single task that we'll use in our map task. It can only accept one input and produce one output
# although this may expand in the future.
@task
def a_mappable_task(a: int) -> str:
    inc = a + 2
    stringified = str(inc)
    return stringified


@task
def coalesce(b: typing.List[str]) -> str:
    coalesced = "".join(b)
    return coalesced


# %%
# To use a map task in your workflow, use the :py:func:`flytekit:flytekit.map_task` function and pass in an individual
# task to be repeated across a collection of inputs. In this case the type of a, ``typing.List[int]`` is a list of the
# input type defined for ``a_mappable_task``.
@workflow
def my_map_workflow(a: typing.List[int]) -> str:
    mapped_out = map_task(a_mappable_task, metadata=TaskMetadata(retries=1))(a=a)
    coalesced = coalesce(b=mapped_out)
    return coalesced


if __name__ == "__main__":
    result = my_map_workflow(a=[1, 2, 3, 4, 5])
    print(f"{result}")
