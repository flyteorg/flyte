"""
Running a Workflow
------------------

Workflows on their own are not runnable directly. However, a launchplan is always bound to a workflow and you can use
launchplans to **launch** a workflow. For cases in which you want the launchplan to have the same arguments as a workflow,
if you are using one of the SDK's to author your workflows - like flytekit, flytekit-java etc, then they should
automatically create a ``default launchplan`` for the workflow.

A ``default launchplan`` has the same name as the workflow and all argument defaults are similar. See
:ref:`sphx_glr_auto_remote_access_run_launchplan.py` to run a workflow via the default launchplan.

:ref:`Tasks also can be executed <sphx_glr_auto_remote_access_run_task.py>` using the launch command.
One difference between running a task and a workflow via launchplans is that launchplans cannot be associated with a
task. This is to avoid triggers and scheduling.

"""
