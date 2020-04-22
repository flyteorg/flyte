.. _api_file_flyteidl/plugins/waitable.proto:

waitable.proto
===============================

.. _api_msg_flyteidl.plugins.Waitable:

flyteidl.plugins.Waitable
-------------------------

`[flyteidl.plugins.Waitable proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/plugins/waitable.proto#L10>`_

Represents an Execution that was launched and could be waited on.

.. code-block:: json

  {
    "wf_exec_id": "{...}",
    "phase": "...",
    "workflow_id": "..."
  }

.. _api_field_flyteidl.plugins.Waitable.wf_exec_id:

wf_exec_id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) 
  
.. _api_field_flyteidl.plugins.Waitable.phase:

phase
  (:ref:`flyteidl.core.WorkflowExecution.Phase <api_enum_flyteidl.core.WorkflowExecution.Phase>`) 
  
.. _api_field_flyteidl.plugins.Waitable.workflow_id:

workflow_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

