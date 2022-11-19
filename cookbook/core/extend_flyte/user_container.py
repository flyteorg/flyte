"""
.. _user_container:

User Container Task Plugin
--------------------------

.. tags:: Extensibility, Contribute, Intermediate

A user container task plugin runs a user-defined container that has the user code.

This tutorial will walk you through writing your own sensor-style plugin that allows users to wait for a file to land
in the object store. Remember that if you follow the flyte/flytekit constructs, you will automatically make your plugin portable
across all cloud platforms that Flyte supports.

Sensor Plugin
*************

A sensor plugin waits for some event to happen before marking the task as success. You need not worry about the
timeout as that will be handled by the flyte engine itself when running in production.

Plugin API
^^^^^^^^^^

.. code-block:: python

    sensor = WaitForObjectStoreFile(metadata=metadata(timeout="1H", retries=10))

    @workflow
    def wait_and_run(path: str) -> int:
        # To demonstrate how to create outputs, we will also return the output from the sensor. The output will be the
        # same as the path
        path = sensor(path=path)
        return do_next(path=path)

"""
import typing
from datetime import timedelta
from time import sleep

from flytekit import TaskMetadata, task, workflow
from flytekit.extend import Interface, PythonTask, context_manager


# %%
# Plugin Structure
# ^^^^^^^^^^^^^^^^
#
# As illustrated above, to achieve this structure we need to create a class named ``WaitForObjectStoreFile``, which
# derives from :py:class:`flytekit.PythonFunctionTask` as follows.
#
class WaitForObjectStoreFile(PythonTask):
    """
    Add documentation here for your plugin.
    This plugin creates an object store file sensor that waits and exits only when the file exists.
    """

    _VAR_NAME: str = "path"

    def __init__(
        self,
        name: str,
        poll_interval: timedelta = timedelta(seconds=10),
        **kwargs,
    ):
        super(WaitForObjectStoreFile, self).__init__(
            task_type="object-store-sensor",
            name=name,
            task_config=None,
            interface=Interface(
                inputs={self._VAR_NAME: str}, outputs={self._VAR_NAME: str}
            ),
            **kwargs,
        )
        self._poll_interval = poll_interval

    def execute(self, **kwargs) -> typing.Any:
        # No need to check for existence, as that is guaranteed.
        path = kwargs[self._VAR_NAME]
        ctx = context_manager.FlyteContext.current_context()
        user_context = ctx.user_space_params
        while True:
            user_context.logging.info(f"Sensing file in path {path}...")
            if ctx.file_access.exists(path):
                user_context.logging.info(f"file in path {path} exists!")
                return path
            user_context.logging.warning(f"file in path {path} does not exists!")
            sleep(self._poll_interval.seconds)


# %%
# Config Objects
# ==============
#
# Flytekit routes to the right plugin based on the type of ``task_config`` class if using the ``@task`` decorator.
# Config is very useful for cases when you want to customize the behavior of the plugin or pass the config information
# to the backend plugin; however, in this case there's no real configuration. The config object can be any class that your
# plugin understands.
#
# .. note::
#
#   Observe that the base class is Generic; it is parameterized with the desired config class.
#
# .. note::
#
#   To create a task decorator-based plugin, ``task_config`` is required.
#   In this example, we are creating a named class plugin, and hence, this construct does not need a plugin.
#
# Refer to the `spark plugin <https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-spark>`__ for an example of a config object.

# %%
# Actual Usage
# ^^^^^^^^^^^^

sensor = WaitForObjectStoreFile(
    name="my-objectstore-sensor",
    metadata=TaskMetadata(retries=10, timeout=timedelta(minutes=20)),
    poll_interval=timedelta(seconds=1),
)


@task
def print_file(path: str) -> str:
    print(path)
    return path


@workflow
def my_workflow(path: str) -> str:
    return print_file(path=sensor(path=path))


# %%
# And of course, you can run the workflow locally using your own new shiny plugin!
if __name__ == "__main__":
    f = "/tmp/some-file"
    with open(f, "w") as w:
        w.write("Hello World!")

    print(my_workflow(path=f))

# %%
# The key takeaways of a user container task plugin are:
#
# - The task object that gets serialized at compile-time is recreated using the user's code at run time.
# - At platform-run-time, the user-decorated function is executed.