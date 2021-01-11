"""
.. _hosted_multi_images:

Working with Multiple Container Images in the same workflow
----------------------------------------------------------------

Typically when you are working locally, it is preferable to install all requirements of your project locally
(maybe in a single virtual environment). This gets complicated when you want to deploy your code to a remote
environment. This is because most tasks in Flyte (function tasks) get deployed using a Docker Container. A docker
container allows you to create an exepected environment for your tasks. Though it is completely possible to build a
single Container Image that contains all your dependencies, it is complicated to achieve this in practice.
Reasons why it is complicated and not recommended,

#. All dependencies in one container increases the size of the container image
#. Some task executions like Spark, Sagemaker based Training, Deep Learning using GPUs need specific runtime configurations, for e.g.,

     - Spark needs JavaVirtualMachine to be installed and Spark entrypoints to be set
     - To use GPUs for deep learning, NVIDIA drivers and other corresponding libraries need to be installed. These are not required for a CPU
     - Sagemaker expects the entrypoint to be specifically designed to accept its parameters

#. Building a singular image may increase the build time for the image itself

.. note::

   Flyte (Service) by default has no requirement that a Workflow is bound to a single Container Image. Flytekit offers a simple interface to easily alter the images that should be associated per task, yet keeping the local execution simple for the user.


For every :py:class:`flytekit.annotated.task.PythonFunctionTask` type task or simply a task that is decorated with the ``@task`` decorator, users can supply rules of how the container image should be bound. By default, flytekit will associate one container image with all tasks. This image is called the ``default`` image.
To alter the image, users should use the ``container_image`` parameter available in the :py:func:`flytekit.task` decorator. Any one of the following is an acceptable

#. Image reference is specified, but the version is derived from the default images version ``container_image="docker.io/redis:{{.image.default.version}},``
#. Both the FQN and the version are derived from the default image ``container_image="{{.image.default.fqn}}:spark-{{.image.default.version}},``

The images themselves are parameterizable in the config in the following format
 ``{{.image.<name>.<attribute>}}``

- ``name`` refers to the name of the image in the image configuration. The name ``default`` is a reserved keyword and will automatically apply to the default image name for this repository
- ``fqn`` refers to the fully qualified name of the image. For example it includes the repository and domain url of the image. e.g. docker.io/my_repo/xyz
- ``version`` refers to the tag of the image. e.g. latest, or python-3.8 etc. If the container_image is not specified then the default configured image for the project is used.

.. note::

    The default image (name + version) is always ``{{.image.default.fqn}}:{{.image.default.version}}``

.. warning:

    It is the responsibility of the user to push a container image that matches the new name described.

"""

# %%
# Lets declare a task that uses an image that is derived from the defaults FQN and version
#
# .. todo::
#
#   Another example for multi-images here
#
# For another example where this is done please refer to :any:`intermediate_spark_dataframes_passing`
