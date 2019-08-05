.. _api_file_flyteidl/core/dynamic_job.proto:

dynamic_job.proto
===============================

.. _api_msg_flyteidl.core.DynamicJobSpec:

flyteidl.core.DynamicJobSpec
----------------------------

`[flyteidl.core.DynamicJobSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/dynamic_job.proto#L11>`_

Describes a set of tasks to execute and how the final outputs are produced.

.. code-block:: json

  {
    "nodes": [],
    "min_successes": "...",
    "outputs": [],
    "tasks": [],
    "subworkflows": []
  }

.. _api_field_flyteidl.core.DynamicJobSpec.nodes:

nodes
  (:ref:`flyteidl.core.Node <api_msg_flyteidl.core.Node>`) A collection of nodes to execute.
  
  
.. _api_field_flyteidl.core.DynamicJobSpec.min_successes:

min_successes
  (`int64 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) An absolute number of successful completions of nodes required to mark this job as succeeded. As soon as this
  criteria is met, the dynamic job will be marked as successful and outputs will be computed. If this number
  becomes impossible to reach (e.g. number of currently running tasks + number of already succeeded tasks <
  min_successes) the task will be aborted immediately and marked as failed. The default value of this field, if not
  specified, is the count of nodes repeated field.
  
  
.. _api_field_flyteidl.core.DynamicJobSpec.outputs:

outputs
  (:ref:`flyteidl.core.Binding <api_msg_flyteidl.core.Binding>`) Describes how to bind the final output of the dynamic job from the outputs of executed nodes. The referenced ids
  in bindings should have the generated id for the subtask.
  
  
.. _api_field_flyteidl.core.DynamicJobSpec.tasks:

tasks
  (:ref:`flyteidl.core.TaskTemplate <api_msg_flyteidl.core.TaskTemplate>`) [Optional] A complete list of task specs referenced in nodes.
  
  
.. _api_field_flyteidl.core.DynamicJobSpec.subworkflows:

subworkflows
  (:ref:`flyteidl.core.WorkflowTemplate <api_msg_flyteidl.core.WorkflowTemplate>`) [Optional] A complete list of task specs referenced in nodes.
  
  

