"""
.. _basics_of_tasks:

Tasks
-----

.. tags:: Basic

Task is a fundamental building block and an extension point of Flyte, which encapsulates the users' code. They possess the following properties:

#. Versioned (usually tied to the ``git sha``)
#. Strong interfaces (specified inputs and outputs)
#. Declarative
#. Independently executable
#. Unit testable

A task in Flytekit can be of two types:

#. A task that has a Python function associated with it. The execution of the task is equivalent to the execution of this function.
#. A task that doesn't have a Python function, e.g., an SQL query or any portable task like Sagemaker prebuilt algorithms, or a service that invokes an API.

Flyte provides multiple plugins for tasks, which can be a backend plugin as well (`Athena <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-aws-athena/flytekitplugins/athena/task.py>`__).

In this example, you will learn how to write and execute a ``Python function task``. Other types of tasks will be covered in the later sections.
"""
# %%
# For any task in Flyte, there is one necessary import, which is:
from flytekit import task


# %%
# A ``PythonFunctionTask`` must always be decorated with the ``@task`` :py:func:`flytekit.task` decorator.
# The task in itself is a regular Python function, although with one exception: it needs all the inputs and outputs to be clearly
# annotated with the types. The types are regular Python types; we'll go over more on this in the :ref:`type-system section <sphx_glr_auto_core_type_system_flyte_python_types.py>`.
@task
def square(n: int) -> int:
    """
     Parameters:
        n (int): name of the parameter for the task will be derived from the name of the input variable
               the type will be automatically deduced to be Types.Integer

    Return:
        int: The label for the output will be automatically assigned and type will be deduced from the annotation

    """
    return n * n


# %%
# In this task, one input is ``n`` which has type ``int``.
# The task ``square`` takes the number ``n`` and returns a new integer (squared value).
#
# .. note::
#
#   Flytekit will assign a default name to the output variable like ``out0``.
#   In case of multiple outputs, each output will be numbered in the order
#   starting with 0, e.g., -> ``out0, out1, out2, ...``.
#
# You can execute a Flyte task as any normal function.
if __name__ == "__main__":
    print(square(n=10))

# %%
#
# Invoke a Task within a Workflow
# ===============================
#
# The primary way to use Flyte tasks is to invoke them in the context of a workflow.

from flytekit import workflow


@workflow
def wf(n: int) -> int:
    return square(n=square(n=n))

# %%
# In this toy example, we're calling the ``square`` task twice and returning the result.


# %%
# .. _single_task_execution:
#
# .. dropdown:: Execute a Single Task without a Workflow
#
#    Although workflows are traditionally composed of multiple tasks with dependencies defined by shared inputs and outputs,
#    it can be helpful to execute a single task during the process of iterating on its definition.
#    It can be tedious to write a new workflow definition every time you want to execute a single task under development
#    but "single task executions" can be used to iterate on task logic easily.
#
#    You can launch a task on Flyte console by providing a Kubernetes service account.
#
#    Alternatively, you can use ``flytectl`` to launch the task. Run the following commands in the ``cookbook`` directory.
#
#    .. note::
#      This example is building a Docker image and pushing it only for sandbox
#      (for non-sandbox, you will have to push the image to a Docker registry).
#
#    Build a Docker image to package the task.
#
#    .. prompt:: bash $
#
#      flytectl sandbox exec -- docker build . --tag "flytebasics:v1" -f core/Dockerfile
#
#    Package the task.
#
#    .. prompt:: bash $
#
#      pyflyte --pkgs core.flyte_basics package --image flytebasics:v1
#
#    Register the task.
#
#    .. prompt:: bash $
#
#      flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz --version v1
#
#    Generate an execution spec file.
#
#    .. prompt:: bash $
#
#      flytectl get task --domain development --project flytesnacks core.flyte_basics.task.square --version v1 --execFile exec_spec.yaml
#
#    Create an execution using the exec spec file.
#
#    .. prompt:: bash $
#
#      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
#
#    .. note::
#      For subsequent executions, you can simply run ``flytectl create execution ...`` and skip the previous commands.
#      Alternatively, you can launch the task from the Flyte console.
#
#    Monitor the execution by providing the execution name from the create execution command.
#
#    .. prompt:: bash $
#
#      flytectl get execution --project flytesnacks --domain development <execname>
