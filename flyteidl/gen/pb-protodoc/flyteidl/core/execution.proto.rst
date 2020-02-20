.. _api_file_flyteidl/core/execution.proto:

execution.proto
=============================

.. _api_msg_flyteidl.core.WorkflowExecution:

flyteidl.core.WorkflowExecution
-------------------------------

`[flyteidl.core.WorkflowExecution proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L9>`_

Indicates various phases of Workflow Execution

.. code-block:: json

  {}



.. _api_enum_flyteidl.core.WorkflowExecution.Phase:

Enum flyteidl.core.WorkflowExecution.Phase
------------------------------------------

`[flyteidl.core.WorkflowExecution.Phase proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L10>`_


.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.UNDEFINED:

UNDEFINED
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.QUEUED:

QUEUED
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.RUNNING:

RUNNING
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.SUCCEEDING:

SUCCEEDING
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.SUCCEEDED:

SUCCEEDED
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.FAILING:

FAILING
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.FAILED:

FAILED
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.ABORTED:

ABORTED
  ⁣
  
.. _api_enum_value_flyteidl.core.WorkflowExecution.Phase.TIMED_OUT:

TIMED_OUT
  ⁣
  

.. _api_msg_flyteidl.core.NodeExecution:

flyteidl.core.NodeExecution
---------------------------

`[flyteidl.core.NodeExecution proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L24>`_

Indicates various phases of Node Execution

.. code-block:: json

  {}



.. _api_enum_flyteidl.core.NodeExecution.Phase:

Enum flyteidl.core.NodeExecution.Phase
--------------------------------------

`[flyteidl.core.NodeExecution.Phase proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L25>`_


.. _api_enum_value_flyteidl.core.NodeExecution.Phase.UNDEFINED:

UNDEFINED
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.QUEUED:

QUEUED
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.RUNNING:

RUNNING
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.SUCCEEDED:

SUCCEEDED
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.FAILING:

FAILING
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.FAILED:

FAILED
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.ABORTED:

ABORTED
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.SKIPPED:

SKIPPED
  ⁣
  
.. _api_enum_value_flyteidl.core.NodeExecution.Phase.TIMED_OUT:

TIMED_OUT
  ⁣
  

.. _api_msg_flyteidl.core.TaskExecution:

flyteidl.core.TaskExecution
---------------------------

`[flyteidl.core.TaskExecution proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L40>`_

Phases that task plugins can go through. Not all phases may be applicable to a specific plugin task,
but this is the cumulative list that customers may want to know about for their task.

.. code-block:: json

  {}



.. _api_enum_flyteidl.core.TaskExecution.Phase:

Enum flyteidl.core.TaskExecution.Phase
--------------------------------------

`[flyteidl.core.TaskExecution.Phase proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L41>`_


.. _api_enum_value_flyteidl.core.TaskExecution.Phase.UNDEFINED:

UNDEFINED
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.TaskExecution.Phase.QUEUED:

QUEUED
  ⁣
  
.. _api_enum_value_flyteidl.core.TaskExecution.Phase.RUNNING:

RUNNING
  ⁣
  
.. _api_enum_value_flyteidl.core.TaskExecution.Phase.SUCCEEDED:

SUCCEEDED
  ⁣
  
.. _api_enum_value_flyteidl.core.TaskExecution.Phase.ABORTED:

ABORTED
  ⁣
  
.. _api_enum_value_flyteidl.core.TaskExecution.Phase.FAILED:

FAILED
  ⁣
  

.. _api_msg_flyteidl.core.ExecutionError:

flyteidl.core.ExecutionError
----------------------------

`[flyteidl.core.ExecutionError proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L53>`_

Represents the error message from the execution.

.. code-block:: json

  {
    "code": "...",
    "message": "...",
    "error_uri": "...",
    "kind": "..."
  }

.. _api_field_flyteidl.core.ExecutionError.code:

code
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Error code indicates a grouping of a type of error.
  More Info: <Link>
  
  
.. _api_field_flyteidl.core.ExecutionError.message:

message
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Detailed description of the error - including stack trace.
  
  
.. _api_field_flyteidl.core.ExecutionError.error_uri:

error_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Full error contents accessible via a URI
  
  
.. _api_field_flyteidl.core.ExecutionError.kind:

kind
  (:ref:`flyteidl.core.ExecutionError.ErrorKind <api_enum_flyteidl.core.ExecutionError.ErrorKind>`) 
  

.. _api_enum_flyteidl.core.ExecutionError.ErrorKind:

Enum flyteidl.core.ExecutionError.ErrorKind
-------------------------------------------

`[flyteidl.core.ExecutionError.ErrorKind proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L62>`_

Error type: System or User

.. _api_enum_value_flyteidl.core.ExecutionError.ErrorKind.UNKNOWN:

UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.ExecutionError.ErrorKind.USER:

USER
  ⁣
  
.. _api_enum_value_flyteidl.core.ExecutionError.ErrorKind.SYSTEM:

SYSTEM
  ⁣
  

.. _api_msg_flyteidl.core.TaskLog:

flyteidl.core.TaskLog
---------------------

`[flyteidl.core.TaskLog proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L72>`_

Log information for the task that is specific to a log sink
When our log story is flushed out, we may have more metadata here like log link expiry

.. code-block:: json

  {
    "uri": "...",
    "name": "...",
    "message_format": "...",
    "ttl": "{...}"
  }

.. _api_field_flyteidl.core.TaskLog.uri:

uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.core.TaskLog.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.core.TaskLog.message_format:

message_format
  (:ref:`flyteidl.core.TaskLog.MessageFormat <api_enum_flyteidl.core.TaskLog.MessageFormat>`) 
  
.. _api_field_flyteidl.core.TaskLog.ttl:

ttl
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) 
  

.. _api_enum_flyteidl.core.TaskLog.MessageFormat:

Enum flyteidl.core.TaskLog.MessageFormat
----------------------------------------

`[flyteidl.core.TaskLog.MessageFormat proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/execution.proto#L74>`_


.. _api_enum_value_flyteidl.core.TaskLog.MessageFormat.UNKNOWN:

UNKNOWN
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.TaskLog.MessageFormat.CSV:

CSV
  ⁣
  
.. _api_enum_value_flyteidl.core.TaskLog.MessageFormat.JSON:

JSON
  ⁣
  
