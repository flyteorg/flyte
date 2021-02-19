"""
Customizing task resources like mem/cpu
--------------------------------------------

One of the reasons why you would want to use a hosted Flyte environment is due to the potential of leveraging CPU, memory and storage resources, far greater than whats available locally.
Flytekit makes it possible to specify these requirements declaratively and close to where the task itself is declared.

"""
import typing

from flytekit import task, workflow


# %%
# In this example the memory required by the function increases as the dataset size increases. For large datasets we may not be able to run locally. Thus we want to provide hints to flyte backend that we want to request for more memory.
# This is done by simply decorating the task with the hints as shown in the following code sample. The following parameters are available to be changed for a task
#
# #. ``cpu_request`` minimum cpu requested
# #. ``cpu_limit`` maximum cpu possible
# #. ``memory_request`` min memory requested
# #. ``memory_limit`` max memory possible
# #. ``gpu_request`` min gpus requested
# #. ``gpu_limit`` max gpus requested, they should be the same?,kub
#
# The following resources are not yet supported, but coming soon
#
# #. ``storage_request`` min storage requested
# #. ``storage_limit`` max storage possible
#
# The acutal values follow the `kubernetes convention <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/>`_.
#
# .. warning::
#
#    We are working on changing how the cpu-request, memory is configured. This API is likely to change in alpha-3. The changes will be minimal
#
@task(cpu_request="1", memory_request="2048")
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
