.. _ui:

How to Use Flyte UI
===================

.. tags:: Basic, UI

Flyte UI is a web-based user interface for Flyte. It helps interact with Flyte objects and builds DAGs out of your workflows.

With Flyte UI, you can:

* Launch tasks
* Launch workflows
* View Versioned Tasks and Workflows
* Trigger Versioned Tasks and Workflows
* Inspect Executions through Inputs, Outputs, Logs, and Graphs
* Clone Executions
* Relaunch Executions
* Recover Executions

.. note::
    `FlyteConsole <https://github.com/flyteorg/flyteconsole>`__ hosts the Flyte user interface code.

Launching Workflows
-------------------

You can launch a workflow by clicking on the **Launch Workflow** button. Workflows are viewable after they are registered.
The UI should be accessible at http://localhost:30081/console.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/launch_execution_001.png
    :alt: "Launch Workflow" button

    Launch a workflow using the "Launch Workflow" button.

|

The end-to-end process from writing code to registering workflows is present in the :std:ref:`getting-started`.

A pop-up window appears with input fields that the execution requires upon clicking the **Launch Workflow** button.
If the default inputs are given, they will be auto-populated.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/launch_execution_002.png
    :alt: Launch form

    A pop-up window appears after clicking the "Launch Workflow" button.

|

An execution can be terminated/aborted by clicking on the **Terminate** button.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/launch_execution_003.png
    :alt: "Terminate" button

    Terminate an execution by clicking the "Terminate" button.

|

Launching Tasks
---------------

You can launch a task by clicking on the **Launch Task** button. Tasks are viewable after they are registered.
The UI should be accessible at http://localhost:30081/console.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/launch_task_001.png
    :alt: "Launch Task" button

    Launch a task by clicking the "Launch Task" button.

|

A pop-up window appears with input fields that the task requires and the role with which the task has to run on clicking the **Launch Task** button.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/launch_task_002.png
    :alt: Launch form

    A pop-up window appears on clicking the "Launch Task" button.

|

Viewing Versioned Tasks and Workflows
-------------------------------------

Every registered Flyte entity is tagged with a version. All the registered versions of workflows and tasks are viewable in the UI.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/versioned_executions.png
    :alt: Versioned workflows

    View versioned workflows.

|

Triggering Versioned Tasks and Workflows
----------------------------------------

Every registered Flyte entity is versioned and can be triggered anytime.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/trigger_versioned_executions.png
    :alt: Trigger versioned workflows

    Trigger versioned workflows.

|

Inspecting Executions
---------------------

Executions can be inspected through the UI. Inputs and Outputs for every node and execution can be viewed.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/inspect_execution_001.png
    :alt: Node's inputs and outputs

    View every execution node's inputs and outputs.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/inspect_execution_002.png
    :alt: Execution's inputs and outputs

    View every execution's inputs and outputs.

|

Logs are accessible as well.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/inspect_execution_003.png
    :alt: Logs

    View Kubernetes logs.

|

Every execution has two views: Nodes and Graph.

A node in the nodes view encapsulates an instance of a task, but it can also contain an entire subworkflow or trigger an external workflow.
More about nodes can be found in :std:ref:`divedeep-nodes`.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/inspect_execution_004.png
    :alt: Nodes

    Inspect execution's nodes in the UI.

|

Graph view showcases a static DAG.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/inspect_execution_005.png
    :alt: DAG

    Inspect execution's DAG in the UI.

|

Cloning Executions
------------------

An execution in the ``RUNNING`` state can be cloned.

Click on the ellipsis on the top right corner of the UI.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/clone_execution_001.png
    :alt: Clone execution

    Step 1: Click on the ellipsis.

|

Click on the **Clone Execution** button.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/clone_execution_002.png
    :alt: Clone execution

    Step 2: "Clone execution" button.

|

Relaunching Executions
----------------------

The **Relaunch** button allows you to relaunch a terminated execution with pre-populated inputs.
This option can be helpful to try out a new version of a Flyte entity.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/relaunch_execution.png
    :alt: Relaunch an execution

    Relaunch an execution.

|

A pop-up window appears on clicking the relaunch button, allowing you to modify the version and inputs.

Recovering Executions
---------------------

Recovery mode allows you to recover an individual execution by copying all successful node executions and running from the failed nodes.
The **Recover** button helps recover a failed execution.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/flyteconsole/recover_execution.png
    :alt: Recover an execution

    Recover an execution.

|
