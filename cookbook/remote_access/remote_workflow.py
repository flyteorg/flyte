"""
Running a Workflow
------------------

Workflows on their own are not runnable directly. However, a launchplan is always bound to a workflow and you can use
launchplans to **launch** a workflow. For cases in which you want the launchplan to have the same arguments as a workflow,
if you are using one of the SDK's to author your workflows - like flytekit, flytekit-java etc, then they should
automatically create a ``default launchplan`` for the workflow.

A ``default launchplan`` has the same name as the workflow and all argument defaults are similar. See
:ref:`Launch Plans` to run a workflow via the default launchplan.

:ref:`Tasks also can be executed <basics_of_tasks>` using the launch command.
One difference between running a task and a workflow via launchplans is that launchplans cannot be associated with a
task. This is to avoid triggers and scheduling.

FlyteRemote
===========

Workflows can be executed with FlyteRemote because under the hood it fetches and triggers a default launch plan.

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config

    # FlyteRemote object is the main entrypoint to API
    remote = FlyteRemote(
        config=Config.for_endpoint(endpoint="flyte.example.net"),
        default_project="flytesnacks",
        default_domain="development",
    )

    # Fetch workflow
    flyte_workflow = remote.fetch_workflow(name="workflows.example.wf", version="v1")

    # Execute
    execution = remote.execute(
        flyte_workflow, inputs={"mean": 1}, execution_name="workflow_execution", wait=True
    )

"""
