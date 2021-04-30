"""
.. _advanced_custom_task_plugin:

How to write your own flytekit task plugins?
----------------------------------------------

Flytekit is designed to be extremely extensible. You can add new task-types that are useful only for your use-cases.
Flyte does come with the capability of extending the backend, but that is only required if you want the capability to be
extended to all users of Flyte, or there is a cost/visibility benefit of doing so. To understand when you should write
a backend plugin refer to - TODO: Flow chart that helps you decide when to write a backend / flytekit plugin?

This tutorial will walk you through writing your own Sensor style plugin, that allows users to wait for a file to land
in the object store. Remember, if you follow the Flyte/flytekit constructs, you will automatically make your plugin portable
across all cloud platforms that Flyte supports

Sensor Plugin
^^^^^^^^^^^^^^
A sensor Plugin waits for some event to happen, before marking the task as success. You do not need to worry about the
timeout as that will be handled by the flyte engine itself, when running in production

Plugin API
^^^^^^^^^^^^

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
# ^^^^^^^^^^^^^^^^^
# As illustrated above to achieve this structure we need to create a Class called  ``WaitForObjectStoreFile``, which
# derives from :py:class:`flytekit.core.PythonFunctionTask` as follows.
#
class WaitForObjectStoreFile(PythonTask):
    """
    Add documentation here for your Plugin.
    This plugin creates an object store file sensor, that waits and exits only when the file exists.
    """

    _VAR_NAME: str = "path"

    def __init__(
        self, name: str, poll_interval: timedelta = timedelta(seconds=10), **kwargs,
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
        # No need to check for existence, as that is guaranteed
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
# Note about Config Objects
# ^^^^^^^^^^^^^^^^^^^^^^^^^^
#
# Flytekit routes to the right plugin only based on the type of task_config class, if using the @task decorator.
# Config is very useful for cases when one wants to customize the behavior of the plugin or pass the config information
# to the backend plugin, but in this case we have no real configuration. The config object is any class, that your
# plugin understands
#
# .. note::
#
#   Observe that the base class is Generic, it is parameterized with the desired config class
#
# .. note::
#
#   To create a task decorator based plugin the Config is required In this example, we are creating a named class plugin
#   This construct does not need a plugin
#
# We will try to cover an example of Config objects in a subsequent tutorial
#
# Actual Usage
# ^^^^^^^^^^^^^

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
# Ofcourse you can run the workflow using your own new shiny plugin locally
if __name__ == "__main__":
    f = "/tmp/some-file"
    with open(f, "w") as w:
        w.write("Hello World!")

    print(my_workflow(path=f))
