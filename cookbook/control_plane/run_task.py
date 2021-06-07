
# NOTE this page was moved from the old Flyte "how to" section and needs to be re-worked with the
# flytekit control plane before being published
"""
.. _howto_exec_single_task:

######################
Running a Single Task
######################


What Are Single Task Executions?
================================

Tasks are the most atomic unit of execution in Flyte.  Although workflows are traditionally composed of multiple tasks with dependencies
defined by shared inputs and outputs, it can be helpful to execute a single task during the process of iterating on its definition.
It can be tedious to write a new workflow definition every time you want to excecute a single task under development, but single task
executions can be used to easily iterate on task logic.

Launching a Single Task
=======================

After building an image with your updated task code, create an execution using launch:

.. code-block:: python

     @inputs(plant=Types.String)
     @outputs(out=Types.String)
     @python_task
     def my_task(wf_params, plant, out)
         ...


     my_single_task_execution = my_task.launch(project="my_flyte_project", domain="development", inputs={'plant': 'ficus'})
     print("Created {}".format(my_single_task_execution.id))

Just like workflow executions, you can optionally pass a user-defined name, labels, annotations, and/or notifications when launching a single task.

The type of ``my_single_task_execution`` is `SdkWorkflowExecution <https://github.com/flyteorg/flytekit/blob/1926b1285591ae941d7fc9bd4c2e4391c5c1b21b/flytekit/common/workflow_execution.py#L14>`_
and has the full set of methods and functionality available for conventional WorkflowExecutions.


Fetching and Launching a Single Task
====================================

Single task executions aren't limited to just tasks you've defined in your code. You can reference previously registered tasks and launch a single task execution like so:

.. code-block:: python

     from flytekit.common.tasks import task as _task

     my_task = _task.SdkTask.fetch("my_flyte_project", "production", "workflows.my_task", "abc123")  # project, domain, name, version

     my_task_exec = my_task.launch(project="my_other_project", domain="development", inputs={'plant': 'philodendron'})
     my_task_exec.wait_for_completion()


Launching a Single Task From the Commandline
============================================

Previously registered tasks can also be launched from the command-line using :ref:`flyte-cli <howto-flytecli>`

.. code-block:: console

    $ flyte-cli -h example.com -p my_flyte_project -d development launch-task \
        -u tsk:my_flyte_project:production:my_complicated_task:abc123 -- an_input=hi \
        other_input=123 more_input=qwerty


Monitoring Single Task Executions in the Flyte Console
======================================================

Single task executions don't have native support in the Flyte console yet, but they are accessible using the same URLs as ordinary workflow executions.

For a console hosted example.com, visit ``example.com/console/projects/<my_project>/domains/<my_domain>/executions/<execution_name>`` to track the progress of your execution. Log links and status changes will be available as your execution progresses.


Registering and Launching a Single Task
=======================================

A certain category of tasks don't rely on custom containers with registered images to run. Therefore, you may find it convenient to use
``register_and_launch`` on a task definition to immediately launch a single task execution, like so:

.. code-block:: python

    containerless_task = SdkPrestoTask(
        task_inputs=inputs(ds=Types.String, count=Types.Integer, rg=Types.String),
        statement="SELECT * FROM flyte.widgets WHERE ds = '{{ .Inputs.ds}}' LIMIT {{ .Inputs.count}}",
        output_schema=Types.Schema([("a", Types.String), ("b", Types.Integer)]),
        routing_group="{{ .Inputs.rg }}",
    )

    my_single_task_execution = containerless_task.register_and_launch(project="my_flyte_project", domain="development",
        inputs={'ds': '2020-02-29', 'count': 10, 'rg': 'my_routing_group'})

"""
