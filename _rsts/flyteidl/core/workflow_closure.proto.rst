.. _api_file_flyteidl/core/workflow_closure.proto:

workflow_closure.proto
====================================

.. _api_msg_flyteidl.core.WorkflowClosure:

flyteidl.core.WorkflowClosure
-----------------------------

`[flyteidl.core.WorkflowClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/workflow_closure.proto#L10>`_

Defines an enclosed package of workflow and tasks it references.

.. code-block:: json

  {
    "workflow": "{...}",
    "tasks": []
  }

.. _api_field_flyteidl.core.WorkflowClosure.workflow:

workflow
  (:ref:`flyteidl.core.WorkflowTemplate <api_msg_flyteidl.core.WorkflowTemplate>`) equired. Workflow template.
  
  
.. _api_field_flyteidl.core.WorkflowClosure.tasks:

tasks
  (:ref:`flyteidl.core.TaskTemplate <api_msg_flyteidl.core.TaskTemplate>`) ptional. A collection of tasks referenced by the workflow. Only needed if the workflow
  references tasks.
  
  

