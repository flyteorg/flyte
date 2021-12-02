.. _ui:

How to Use Flyte UI
===================

Flyte UI is a web-based user interface for Flyte. It helps interact with Flyte objects and builds DAGs out of your workflows.

With Flyte UI, you can:

* Launch Execution
* View Versioned Executions
* Trigger Versioned Executions
* Inspect Execution through Inputs, Outputs, Logs, and Graphs
* Clone Execution
* Relaunch Execution
* Recover Execution

.. note::
    `Flyte Console <https://github.com/flyteorg/flyteconsole>`__ hosts the Flyte user interface code.

Launch Execution
----------------

Launch an execution by clicking on the **Launch** button. Workflows and tasks are viewable after they are registered.
UI should be accessible at http://localhost:8000/console.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/launch_execution_001.png
    :alt: "Launch workflow" button

    Launch a workflow using the "Launch Workflow" button.

|

The end-to-end process from writing code to registering workflows is present in the :std:ref:`gettingstarted_implement`.

After clicking on the **Launch** button, pop-up window gets displayed with inputs that the execution requires.
If default inputs are given, they get auto-populated.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/launch_execution_002.png
    :alt: Launch form

    A pop-window gets displayed after clicking on the "Launch Workflow" button.

|

An execution can be terminated/aborted by clicking on the **Terminate** button.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/launch_execution_003.png
    :alt: "Terminate" button

    Terminate an execution by clicking on the "Terminate" button.

|

View Versioned Executions
-------------------------

Every registered Flyte entity is tagged with a version. All the registered versions of workflows and tasks are viewable in the UI.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/versioned_executions.png
    :alt: Versioned executions

    Versioned executions are viewable in the UI.

|

Trigger Versioned Executions
----------------------------

Every registered workflow is versioned and can be triggered anytime.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/trigger_versioned_executions.png
    :alt: Trigger versioned execution

    Trigger versioned executions.

|

Inspect Executions
------------------

Executions can be inspected through the UI. Inputs and Outputs can be viewed for every node and execution.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/inspect_execution_001.png
    :alt: Node's inputs and outputs

    Every execution node's inputs and outputs are viewable.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/inspect_execution_002.png
    :alt: Execution's inputs and outputs

    Every execution's inputs and outputs are viewable.

|

Logs should be accessible as well.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/inspect_execution_003.png
    :alt: Logs

    Kubernetes logs are viewable.

|

Every execution has two views: Node and Graph.

A node will encapsulate an instance of a task, but it can also contain an entire subworkflow or trigger a child workflow.
More about nodes can be found in :std:ref:`divedeep-nodes`.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/inspect_execution_004.png
    :alt: Nodes

    Inspect execution's nodes in the UI.

|

Graph view showcases the DAG.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/inspect_execution_005.png
    :alt: DAG

    Inspect execution's DAG in the UI.

|

Clone Execution
----------------

An execution in the RUNNING state can be cloned.

Click on the kebab menu on the top right corner of the UI.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/clone_execution_001.png
    :alt: Clone execution

    Step 1: Click on the kebab menu.

|

Click on the **Clone Execution** button.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/clone_execution_002.png
    :alt: Clone execution

    Step 2: Clone execution.

|

Relaunch Execution
------------------

**Relaunch** button allows you to relaunch an execution with pre-populated inputs.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/relaunch_execution.png
    :alt: Relaunch execution

    Relaunch an execution.

|

A pop-up window gets displayed and it allows you to modify the version and inputs.

Recover Execution
-----------------

Recovery mode allows you to recover an individual execution by copying all successful node executions and running from the failed nodes.
A **Recover** button is available to recover an execution.

|

.. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flyteconsole/recover_execution.png
    :alt: Recover execution

    Recover an execution.

|
