"""
Customizing task resources like mem/cpu
--------------------------------------------

One of the reasons why you would want to use a hosted Flyte environment is due to the potential of leveraging CPU,
memory and storage resources, far greater than what's available locally.
Flytekit makes it possible to specify these requirements declaratively and close to where the task itself is declared.

"""

# %%
# In this example the memory required by the function increases as the dataset size increases.
# For large datasets we may not be able to run locally. Thus we want to provide hints to flyte backend that we want to request for more memory.
# This is done by simply decorating the task with the hints as shown in the following code sample.
#
# Tasks can have ``requests`` and ``limits`` which mirror the native `equivalents in kubernetes <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits>`__
# A task can possibly be allocated more resources than it requests, but never more than its limit.
# Requests are treated as hints which are used to schedule your tasks on nodes with available resources, whereas limits
# are hard constraints.
#
# For either a request or limit, refer to the :py:class:`flytekit:flytekit.Resource` documentation.
#
# The following attributes can be specified for a ``Resource``.
#
# #. ``cpu``
# #. ``mem``
# #. ``gpu``
#
# The ``storage`` resources option is not yet supported, but coming soon
#
# The acutal values follow the `kubernetes convention <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes>`_.


import typing

from flytekit import Resources, task, workflow


@task(requests=Resources(cpu="1", mem="2048"), limits=Resources(cpu="2", mem="4096"))
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
