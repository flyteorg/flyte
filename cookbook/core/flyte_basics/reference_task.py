"""
Reference Task
--------------

A :py:func:`flytekit.reference_task` references the Flyte tasks that have already been defined, serialized, and registered.
You can reference tasks from other projects and create a workflow that uses tasks declared by others.
These tasks can be in their own containers, python runtimes, flytekit versions, and even different languages.

This example demonstrates the usage of reference tasks.

.. note::
    Reference tasks cannot be run locally. You must mock them out.
"""

# %%
# First, let's import the dependencies.
from flytekit import reference_task, workflow
from flytekit.types.file import JPEGImageFile


# %%
# We define a ``reference_task`` by importing the ``rotate()`` function.
@reference_task(
    project="flytesnacks",
    domain="development",
    name="core.flyte_basics.files.rotate",
    version="{{ registration.version }}",
)
def rotate(image_location: JPEGImageFile) -> JPEGImageFile:
    ...


# %%
# .. note::
#     The Macro ``{{ registration.version }}`` is filled during the registration time by `flytectl register`. This is usually not required for reference tasks; you should
#     ideally bind to a specific version of the entity -- task / launchplan. But in the case of this example, we are registering both the actual task ``core.flyte_basics.files.rotate`` and
#     the workflow that references it. Thus we want it to actually be updated to the version of a specific release of Flytesnacks. This is why we use the ``{{ registration.version }}`` macro.
#
#     A typical example of reference task would look more like this:
#
#     .. code-block:: python
#
#          @reference_task(
#               project="flytesnacks",
#               domain="development",
#               name="core.flyte_basics.files.rotate",
#               version="d06cebcfbeabc02b545eefa13a01c6ca992940c8", # If using GIT for versioning OR 0.16.0, if semver
#           )
#           def rotate(...):
#               ...

# %%
# Next, we define a workflow to call the ``rotate()`` task.
@workflow
def wf():
    return rotate(
        image_location="https://upload.wikimedia.org/wikipedia/commons/d/d2/Julia_set_%28C_%3D_0.285%2C_0.01%29.jpg"
    )


# %%
# Lastly, we can trigger the workflow in a sandbox environment.
