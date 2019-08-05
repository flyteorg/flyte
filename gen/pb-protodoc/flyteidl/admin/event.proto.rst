.. _api_file_flyteidl/admin/event.proto:

event.proto
==========================

.. _api_msg_flyteidl.admin.EventErrorAlreadyInTerminalState:

flyteidl.admin.EventErrorAlreadyInTerminalState
-----------------------------------------------

`[flyteidl.admin.EventErrorAlreadyInTerminalState proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L7>`_


.. code-block:: json

  {
    "current_phase": "..."
  }

.. _api_field_flyteidl.admin.EventErrorAlreadyInTerminalState.current_phase:

current_phase
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_flyteidl.admin.EventFailureReason:

flyteidl.admin.EventFailureReason
---------------------------------

`[flyteidl.admin.EventFailureReason proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L11>`_


.. code-block:: json

  {
    "already_in_terminal_state": "{...}"
  }

.. _api_field_flyteidl.admin.EventFailureReason.already_in_terminal_state:

already_in_terminal_state
  (:ref:`flyteidl.admin.EventErrorAlreadyInTerminalState <api_msg_flyteidl.admin.EventErrorAlreadyInTerminalState>`) 
  
  


.. _api_msg_flyteidl.admin.WorkflowExecutionEventRequest:

flyteidl.admin.WorkflowExecutionEventRequest
--------------------------------------------

`[flyteidl.admin.WorkflowExecutionEventRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L18>`_

Request to send a notification that a workflow execution event has occurred.

.. code-block:: json

  {
    "request_id": "...",
    "event": "{...}"
  }

.. _api_field_flyteidl.admin.WorkflowExecutionEventRequest.request_id:

request_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique ID for this request that can be traced between services
  
  
.. _api_field_flyteidl.admin.WorkflowExecutionEventRequest.event:

event
  (:ref:`flyteidl.event.WorkflowExecutionEvent <api_msg_flyteidl.event.WorkflowExecutionEvent>`) Details about the event that occurred.
  
  


.. _api_msg_flyteidl.admin.WorkflowExecutionEventResponse:

flyteidl.admin.WorkflowExecutionEventResponse
---------------------------------------------

`[flyteidl.admin.WorkflowExecutionEventResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L26>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.NodeExecutionEventRequest:

flyteidl.admin.NodeExecutionEventRequest
----------------------------------------

`[flyteidl.admin.NodeExecutionEventRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L31>`_

Request to send a notification that a node execution event has occurred.

.. code-block:: json

  {
    "request_id": "...",
    "event": "{...}"
  }

.. _api_field_flyteidl.admin.NodeExecutionEventRequest.request_id:

request_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique ID for this request that can be traced between services
  
  
.. _api_field_flyteidl.admin.NodeExecutionEventRequest.event:

event
  (:ref:`flyteidl.event.NodeExecutionEvent <api_msg_flyteidl.event.NodeExecutionEvent>`) Details about the event that occurred.
  
  


.. _api_msg_flyteidl.admin.NodeExecutionEventResponse:

flyteidl.admin.NodeExecutionEventResponse
-----------------------------------------

`[flyteidl.admin.NodeExecutionEventResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L39>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.admin.TaskExecutionEventRequest:

flyteidl.admin.TaskExecutionEventRequest
----------------------------------------

`[flyteidl.admin.TaskExecutionEventRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L44>`_

Request to send a notification that a task execution event has occurred.

.. code-block:: json

  {
    "request_id": "...",
    "event": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionEventRequest.request_id:

request_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique ID for this request that can be traced between services
  
  
.. _api_field_flyteidl.admin.TaskExecutionEventRequest.event:

event
  (:ref:`flyteidl.event.TaskExecutionEvent <api_msg_flyteidl.event.TaskExecutionEvent>`) Details about the event that occurred.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionEventResponse:

flyteidl.admin.TaskExecutionEventResponse
-----------------------------------------

`[flyteidl.admin.TaskExecutionEventResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/event.proto#L52>`_


.. code-block:: json

  {}



