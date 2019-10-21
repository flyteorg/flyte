.. _api_file_flyteidl/admin/task_execution.proto:

task_execution.proto
===================================

.. _api_msg_flyteidl.admin.TaskExecutionGetRequest:

flyteidl.admin.TaskExecutionGetRequest
--------------------------------------

`[flyteidl.admin.TaskExecutionGetRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L13>`_

A message used to fetch a single task execution entity.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionGetRequest.id:

id
  (:ref:`flyteidl.core.TaskExecutionIdentifier <api_msg_flyteidl.core.TaskExecutionIdentifier>`) Unique identifier for the task execution.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionListRequest:

flyteidl.admin.TaskExecutionListRequest
---------------------------------------

`[flyteidl.admin.TaskExecutionListRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L19>`_

Represents a request structure to retrieve a list of task execution entities.

.. code-block:: json

  {
    "node_execution_id": "{...}",
    "limit": "...",
    "token": "...",
    "filters": "...",
    "sort_by": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionListRequest.node_execution_id:

node_execution_id
  (:ref:`flyteidl.core.NodeExecutionIdentifier <api_msg_flyteidl.core.NodeExecutionIdentifier>`) Indicates the node execution to filter by.
  
  
.. _api_field_flyteidl.admin.TaskExecutionListRequest.limit:

limit
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates the number of resources to be returned.
  
  
.. _api_field_flyteidl.admin.TaskExecutionListRequest.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query.
  +optional
  
  
.. _api_field_flyteidl.admin.TaskExecutionListRequest.filters:

filters
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Indicates a list of filters passed as string.
  More info on constructing filters : <Link>
  +optional
  
  
.. _api_field_flyteidl.admin.TaskExecutionListRequest.sort_by:

sort_by
  (:ref:`flyteidl.admin.Sort <api_msg_flyteidl.admin.Sort>`) Sort ordering for returned list.
  +optional
  
  


.. _api_msg_flyteidl.admin.TaskExecution:

flyteidl.admin.TaskExecution
----------------------------

`[flyteidl.admin.TaskExecution proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L44>`_

Encapsulates all details for a single task execution entity.
A task execution represents an instantiated task, including all inputs and additional
metadata as well as computed results included state, outputs, and duration-based attributes.

.. code-block:: json

  {
    "id": "{...}",
    "input_uri": "...",
    "closure": "{...}",
    "is_parent": "..."
  }

.. _api_field_flyteidl.admin.TaskExecution.id:

id
  (:ref:`flyteidl.core.TaskExecutionIdentifier <api_msg_flyteidl.core.TaskExecutionIdentifier>`) Unique identifier for the task execution.
  
  
.. _api_field_flyteidl.admin.TaskExecution.input_uri:

input_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Path to remote data store where input blob is stored.
  
  
.. _api_field_flyteidl.admin.TaskExecution.closure:

closure
  (:ref:`flyteidl.admin.TaskExecutionClosure <api_msg_flyteidl.admin.TaskExecutionClosure>`) Task execution details and results.
  
  
.. _api_field_flyteidl.admin.TaskExecution.is_parent:

is_parent
  (`bool <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Whether this task spawned nodes.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionList:

flyteidl.admin.TaskExecutionList
--------------------------------

`[flyteidl.admin.TaskExecutionList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L59>`_

Response structure for a query to list of task execution entities.

.. code-block:: json

  {
    "task_executions": [],
    "token": "..."
  }

.. _api_field_flyteidl.admin.TaskExecutionList.task_executions:

task_executions
  (:ref:`flyteidl.admin.TaskExecution <api_msg_flyteidl.admin.TaskExecution>`) 
  
.. _api_field_flyteidl.admin.TaskExecutionList.token:

token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) In the case of multiple pages of results, the server-provided token can be used to fetch the next page
  in a query. If there are no more results, this value will be empty.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionClosure:

flyteidl.admin.TaskExecutionClosure
-----------------------------------

`[flyteidl.admin.TaskExecutionClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L68>`_

Container for task execution details and results.

.. code-block:: json

  {
    "output_uri": "...",
    "error": "{...}",
    "phase": "...",
    "logs": [],
    "started_at": "{...}",
    "duration": "{...}",
    "created_at": "{...}",
    "updated_at": "{...}",
    "custom_info": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionClosure.output_uri:

output_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Path to remote data store where output blob is stored if the execution succeeded (and produced outputs).
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.admin.TaskExecutionClosure.output_uri>`, :ref:`error <api_field_flyteidl.admin.TaskExecutionClosure.error>` may be set.
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.error:

error
  (:ref:`flyteidl.core.ExecutionError <api_msg_flyteidl.core.ExecutionError>`) Error information for the task execution. Populated if the execution failed.
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.admin.TaskExecutionClosure.output_uri>`, :ref:`error <api_field_flyteidl.admin.TaskExecutionClosure.error>` may be set.
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.phase:

phase
  (:ref:`flyteidl.core.TaskExecution.Phase <api_enum_flyteidl.core.TaskExecution.Phase>`) The last recorded phase for this task execution.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.logs:

logs
  (:ref:`flyteidl.core.TaskLog <api_msg_flyteidl.core.TaskLog>`) Detailed log information output by the task execution.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.started_at:

started_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the task execution began running.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.duration:

duration
  (:ref:`google.protobuf.Duration <api_msg_google.protobuf.Duration>`) The amount of time the task execution spent running.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.created_at:

created_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the task execution was created.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.updated_at:

updated_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) Time at which the task execution was last updated.
  
  
.. _api_field_flyteidl.admin.TaskExecutionClosure.custom_info:

custom_info
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) Custom data specific to the task plugin.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionGetDataRequest:

flyteidl.admin.TaskExecutionGetDataRequest
------------------------------------------

`[flyteidl.admin.TaskExecutionGetDataRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L100>`_

Request structure to fetch inputs and output urls for a task execution.

.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionGetDataRequest.id:

id
  (:ref:`flyteidl.core.TaskExecutionIdentifier <api_msg_flyteidl.core.TaskExecutionIdentifier>`) The identifier of the task execution for which to fetch inputs and outputs.
  
  


.. _api_msg_flyteidl.admin.TaskExecutionGetDataResponse:

flyteidl.admin.TaskExecutionGetDataResponse
-------------------------------------------

`[flyteidl.admin.TaskExecutionGetDataResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/task_execution.proto#L106>`_

Response structure for TaskExecutionGetDataRequest which contains inputs and outputs for a task execution.

.. code-block:: json

  {
    "inputs": "{...}",
    "outputs": "{...}"
  }

.. _api_field_flyteidl.admin.TaskExecutionGetDataResponse.inputs:

inputs
  (:ref:`flyteidl.admin.UrlBlob <api_msg_flyteidl.admin.UrlBlob>`) Signed url to fetch a core.LiteralMap of task execution inputs.
  
  
.. _api_field_flyteidl.admin.TaskExecutionGetDataResponse.outputs:

outputs
  (:ref:`flyteidl.admin.UrlBlob <api_msg_flyteidl.admin.UrlBlob>`) Signed url to fetch a core.LiteralMap of task execution outputs.
  
  

