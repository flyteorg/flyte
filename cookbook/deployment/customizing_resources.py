"""
Customizing Task Resources
---------------------------

One of the reasons to use a hosted Flyte environment is the potential of leveraging CPU, memory and storage resources, far greater than what's available locally.
Flytekit makes it possible to specify these requirements declaratively and close to where the task itself is declared.

"""

# %%
# In this example, the memory required by the function increases as the dataset size increases.
# Large datasets may not be able to run locally, so we would want to provide hints to flyte backend to request for more memory.
# This is done by simply decorating the task with the hints as shown in the following code sample.
#
# Tasks can have ``requests`` and ``limits`` which mirror the native `equivalents in kubernetes <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits>`__
# A task can possibly be allocated more resources than it requests, but never more than its limit.
# Requests are treated as hints which are used to schedule tasks on nodes with available resources, whereas limits
# are hard constraints.
#
# For either a request or limit, refer to the :py:class:`flytekit:flytekit.Resources` documentation.
#
# The following attributes can be specified for a ``Resource``.
#
# #. ``cpu``
# #. ``mem``
# #. ``gpu``
#
# To ensure regular tasks that do not require GPUs are not scheduled on GPU nodes, a separate node group for GPU nodes can be configured with `taints <https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/>`_.
#
# To ensure tasks that do require GPUs get the needed tolerations on their pods, set up flytepropeller using the following `configuration <https://github.com/flyteorg/flytepropeller/blob/v0.10.5/config.yaml#L51,L56>`_. Make sure that this toleration config matches the taint config that you have configured to protect your gpu providing nodes from having to deal with regular non-gpu workloads (pods).
#
# The ``storage`` resources option is not yet supported, but coming soon.
#
# The acutal values follow the `kubernetes convention <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes>`_.


import typing

from flytekit import Resources, task, workflow


@task(requests=Resources(cpu="1", mem="100Mi"), limits=Resources(cpu="2", mem="150Mi"))
def count_unique_numbers(x: typing.List[int]) -> int:
    s = set()
    for i in x:
        s.add(i)
    return len(s)


# %%
# Now lets create a dummy task that squares the number
@task
def square(x: int) -> int:
    return x * x


# %%
# The tasks decorated with memory and storage hints can be used like regular tasks in a workflow, as follows


@workflow
def my_workflow(x: typing.List[int]) -> int:
    return square(x=count_unique_numbers(x=x))


# %%
# The workflow and task can be executed locally
if __name__ == "__main__":
    print(count_unique_numbers(x=[1, 1, 2]))
    print(my_workflow(x=[1, 1, 2]))
