"""
Map Tasks
---------

.. tags:: Intermediate

A map task lets you run a pod task or a regular task over a list of inputs within a single workflow node.
This means you can run thousands of instances of the task without creating a node for every instance, providing valuable performance gains!

Some use cases of map tasks include:

* Several inputs must run through the same code logic
* Multiple data batches need to be processed in parallel
* Hyperparameter optimization

Let's look at an example now!
"""

# %%
# First, import the libraries.
from typing import List

from flytekit import Resources, map_task, task, workflow


# %%
# Next, define a task to use in the map task.
#
# .. note::
#   A map task can only accept one input and produce one output.
@task
def a_mappable_task(a: int) -> str:
    inc = a + 2
    stringified = str(inc)
    return stringified


# %%
# Also define a task to reduce the mapped output to a string.
@task
def coalesce(b: List[str]) -> str:
    coalesced = "".join(b)
    return coalesced


# %%
# We send ``a_mappable_task`` to be repeated across a collection of inputs to the :py:func:`~flytekit:flytekit.map_task` function.
# In the example, ``a`` of type ``List[int]`` is the input.
# The task ``a_mappable_task`` is run for each element in the list.
#
# ``with_overrides`` is useful to set resources for individual map task.
@workflow
def my_map_workflow(a: List[int]) -> str:
    mapped_out = map_task(a_mappable_task)(a=a).with_overrides(
        requests=Resources(mem="300Mi"),
        limits=Resources(mem="500Mi"),
        retries=1,
    )
    coalesced = coalesce(b=mapped_out)
    return coalesced


# %%
# Lastly, we can run the workflow locally!
if __name__ == "__main__":
    result = my_map_workflow(a=[1, 2, 3, 4, 5])
    print(f"{result}")

# %%
# When defining a map task, avoid calling other tasks in it. Flyte
# can't accurately register tasks that call other tasks.  While Flyte
# will correctly execute a task that calls other tasks, it will not be
# able to give full performance advantages. This is
# especially true for map tasks.
#
# In this example, the map task ``suboptimal_mappable_task`` would not
# give you the best performance.
@task
def upperhalf(a: int) -> int:
    return a / 2 + 1

@task
def suboptimal_mappable_task(a: int) -> str:
    inc = upperhalf(a=a)
    stringified = str(inc)
    return stringified


# %%
#
# By default, the map task uses the K8s Array plugin. Map tasks can
# also run on alternate execution backends, such as
# `AWS Batch <https://docs.flyte.org/en/latest/deployment/plugin_setup/aws/batch.html#deployment-plugin-setup-aws-array>`__,
# a provisioned service that can scale to great sizes.


# %%
# Map a Task with Multiple Inputs
# ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# You might need to map a task with multiple inputs.
#
# For example, we have a task that takes 3 inputs.
@task
def multi_input_task(quantity: int, price: float, shipping: float) -> float:
    return quantity * price * shipping

# %%
# But we only want to map this task with the ``quantity`` input
# while the other inputs stay the same. Since a map task accepts only
# one input, we can do this by creating a new task that prepares the
# map task's inputs.
#
# We start by putting the inputs in a Dataclass and
# ``dataclass_json``. We also define our helper task to prepare the map
# task's inputs.
from dataclasses import dataclass
from dataclasses_json import dataclass_json

@dataclass_json
@dataclass
class MapInput:
    quantity: float
    price: float
    shipping: float

@task
def prepare_map_inputs(list_q: List[float], p: float, s: float) -> List[MapInput]:
    return [MapInput(q, p, s) for q in list_q]

# %%
# Then we refactor ``multi_input_task``. Instead of 3 inputs, ``mappable_task``
# has a single input.
@task
def mappable_task(input: MapInput) -> float:
    return input.quantity * input.price * input.shipping

# %%
# Our workflow prepares a new list of inputs for the map task.
@workflow
def multiple_workflow(list_q: List[float], p: float, s: float) -> List[float]:
    prepared = prepare_map_inputs(list_q=list_q, p=p, s=s)
    return map_task(mappable_task)(input=prepared)

# %%
# We can run our multi-input map task locally.
if __name__ == "__main__":
    result = multiple_workflow(list_q=[1.0, 2.0, 3.0, 4.0, 5.0], p=6.0, s=7.0)
    print(f"{result}")

# %%
# .. panels::
#     :header: text-center
#     :column: col-lg-12 p-2
#
#     .. link-button:: https://blog.flyte.org/map-tasks-in-flyte
#        :type: url
#        :text: Blog Post
#        :classes: btn-block stretched-link
#     ^^^^^^^^^^^^
#     An article on how to use Map Taks in Flyte.
#
# .. toctree::
#     :maxdepth: -1
#     :caption: Contents
#     :hidden:
#
#     Blog Post <https://blog.flyte.org/map-tasks-in-flyte>
