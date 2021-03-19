.. _api_file_flyteidl/event/event.proto:

event.proto
==========================

.. _api_msg_flyteidl.event.WorkflowExecutionEvent:

flyteidl.event.WorkflowExecutionEvent
-------------------------------------

`[flyteidl.event.WorkflowExecutionEvent proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L12>`_


.. code-block:: json

  {
    "execution_id": "{...}",
    "producer_id": "...",
    "phase": "...",
    "occurred_at": "{...}",
    "output_uri": "...",
    "error": "{...}"
  }

.. _api_field_flyteidl.event.WorkflowExecutionEvent.execution_id:

execution_id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) Workflow execution id
  
  
.. _api_field_flyteidl.event.WorkflowExecutionEvent.producer_id:

producer_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) the id of the originator (Propeller) of the event
  
  
.. _api_field_flyteidl.event.WorkflowExecutionEvent.phase:

phase
  (:ref:`flyteidl.core.WorkflowExecution.Phase <api_enum_flyteidl.core.WorkflowExecution.Phase>`) 
  
.. _api_field_flyteidl.event.WorkflowExecutionEvent.occurred_at:

occurred_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) This timestamp represents when the original event occurred, it is generated
  by the executor of the workflow.
  
  
.. _api_field_flyteidl.event.WorkflowExecutionEvent.output_uri:

output_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URL to the output of the execution, it encodes all the information
  including Cloud source provider. ie., s3://...
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.WorkflowExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.WorkflowExecutionEvent.error>` may be set.
  
.. _api_field_flyteidl.event.WorkflowExecutionEvent.error:

error
  (:ref:`flyteidl.core.ExecutionError <api_msg_flyteidl.core.ExecutionError>`) Error information for the execution
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.WorkflowExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.WorkflowExecutionEvent.error>` may be set.
  


.. _api_msg_flyteidl.event.NodeExecutionEvent:

flyteidl.event.NodeExecutionEvent
---------------------------------

`[flyteidl.event.NodeExecutionEvent proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L35>`_


.. code-block:: json

  {
    "id": "{...}",
    "producer_id": "...",
    "phase": "...",
    "occurred_at": "{...}",
    "input_uri": "...",
    "output_uri": "...",
    "error": "{...}",
    "workflow_node_metadata": "{...}",
    "task_node_metadata": "{...}",
    "parent_task_metadata": "{...}",
    "parent_node_metadata": "{...}",
    "retry_group": "...",
    "spec_node_id": "...",
    "node_name": "..."
  }

.. _api_field_flyteidl.event.NodeExecutionEvent.id:

id
  (:ref:`flyteidl.core.NodeExecutionIdentifier <api_msg_flyteidl.core.NodeExecutionIdentifier>`) Unique identifier for this node execution
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.producer_id:

producer_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) the id of the originator (Propeller) of the event
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.phase:

phase
  (:ref:`flyteidl.core.NodeExecution.Phase <api_enum_flyteidl.core.NodeExecution.Phase>`) 
  
.. _api_field_flyteidl.event.NodeExecutionEvent.occurred_at:

occurred_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) This timestamp represents when the original event occurred, it is generated
  by the executor of the node.
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.input_uri:

input_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.event.NodeExecutionEvent.output_uri:

output_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URL to the output of the execution, it encodes all the information
  including Cloud source provider. ie., s3://...
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.NodeExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.NodeExecutionEvent.error>` may be set.
  
.. _api_field_flyteidl.event.NodeExecutionEvent.error:

error
  (:ref:`flyteidl.core.ExecutionError <api_msg_flyteidl.core.ExecutionError>`) Error information for the execution
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.NodeExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.NodeExecutionEvent.error>` may be set.
  
.. _api_field_flyteidl.event.NodeExecutionEvent.workflow_node_metadata:

workflow_node_metadata
  (:ref:`flyteidl.event.WorkflowNodeMetadata <api_msg_flyteidl.event.WorkflowNodeMetadata>`) 
  Additional metadata to do with this event's node target based
  on the node type
  
  
  Only one of :ref:`workflow_node_metadata <api_field_flyteidl.event.NodeExecutionEvent.workflow_node_metadata>`, :ref:`task_node_metadata <api_field_flyteidl.event.NodeExecutionEvent.task_node_metadata>` may be set.
  
.. _api_field_flyteidl.event.NodeExecutionEvent.task_node_metadata:

task_node_metadata
  (:ref:`flyteidl.event.TaskNodeMetadata <api_msg_flyteidl.event.TaskNodeMetadata>`) 
  Additional metadata to do with this event's node target based
  on the node type
  
  
  Only one of :ref:`workflow_node_metadata <api_field_flyteidl.event.NodeExecutionEvent.workflow_node_metadata>`, :ref:`task_node_metadata <api_field_flyteidl.event.NodeExecutionEvent.task_node_metadata>` may be set.
  
.. _api_field_flyteidl.event.NodeExecutionEvent.parent_task_metadata:

parent_task_metadata
  (:ref:`flyteidl.event.ParentTaskExecutionMetadata <api_msg_flyteidl.event.ParentTaskExecutionMetadata>`) [To be deprecated] Specifies which task (if any) launched this node.
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.parent_node_metadata:

parent_node_metadata
  (:ref:`flyteidl.event.ParentNodeExecutionMetadata <api_msg_flyteidl.event.ParentNodeExecutionMetadata>`) Specifies the parent node of the current node execution. Node executions at level zero will not have a parent node.
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.retry_group:

retry_group
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Retry group to indicate grouping of nodes by retries
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.spec_node_id:

spec_node_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifier of the node in the original workflow/graph
  This maps to value of WorkflowTemplate.nodes[X].id
  
  
.. _api_field_flyteidl.event.NodeExecutionEvent.node_name:

node_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Friendly readable name for the node
  
  


.. _api_msg_flyteidl.event.WorkflowNodeMetadata:

flyteidl.event.WorkflowNodeMetadata
-----------------------------------

`[flyteidl.event.WorkflowNodeMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L84>`_

For Workflow Nodes we need to send information about the workflow that's launched

.. code-block:: json

  {
    "execution_id": "{...}"
  }

.. _api_field_flyteidl.event.WorkflowNodeMetadata.execution_id:

execution_id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) 
  


.. _api_msg_flyteidl.event.TaskNodeMetadata:

flyteidl.event.TaskNodeMetadata
-------------------------------

`[flyteidl.event.TaskNodeMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L88>`_


.. code-block:: json

  {
    "cache_status": "...",
    "catalog_key": "{...}"
  }

.. _api_field_flyteidl.event.TaskNodeMetadata.cache_status:

cache_status
  (:ref:`flyteidl.core.CatalogCacheStatus <api_enum_flyteidl.core.CatalogCacheStatus>`) Captures the status of caching for this execution.
  
  
.. _api_field_flyteidl.event.TaskNodeMetadata.catalog_key:

catalog_key
  (:ref:`flyteidl.core.CatalogMetadata <api_msg_flyteidl.core.CatalogMetadata>`) This structure carries the catalog artifact information
  
  


.. _api_msg_flyteidl.event.ParentTaskExecutionMetadata:

flyteidl.event.ParentTaskExecutionMetadata
------------------------------------------

`[flyteidl.event.ParentTaskExecutionMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L96>`_


.. code-block:: json

  {
    "id": "{...}"
  }

.. _api_field_flyteidl.event.ParentTaskExecutionMetadata.id:

id
  (:ref:`flyteidl.core.TaskExecutionIdentifier <api_msg_flyteidl.core.TaskExecutionIdentifier>`) 
  


.. _api_msg_flyteidl.event.ParentNodeExecutionMetadata:

flyteidl.event.ParentNodeExecutionMetadata
------------------------------------------

`[flyteidl.event.ParentNodeExecutionMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L100>`_


.. code-block:: json

  {
    "node_id": "..."
  }

.. _api_field_flyteidl.event.ParentNodeExecutionMetadata.node_id:

node_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique identifier of the parent node id within the execution
  This is value of core.NodeExecutionIdentifier.node_id of the parent node 
  
  


.. _api_msg_flyteidl.event.TaskExecutionEvent:

flyteidl.event.TaskExecutionEvent
---------------------------------

`[flyteidl.event.TaskExecutionEvent proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L107>`_

Plugin specific execution event information. For tasks like Python, Hive, Spark, DynamicJob.

.. code-block:: json

  {
    "task_id": "{...}",
    "parent_node_execution_id": "{...}",
    "retry_attempt": "...",
    "phase": "...",
    "producer_id": "...",
    "logs": [],
    "occurred_at": "{...}",
    "input_uri": "...",
    "output_uri": "...",
    "error": "{...}",
    "custom_info": "{...}",
    "phase_version": "...",
    "reason": "...",
    "task_type": "...",
    "metadata": "{...}"
  }

.. _api_field_flyteidl.event.TaskExecutionEvent.task_id:

task_id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) ID of the task. In combination with the retryAttempt this will indicate
  the task execution uniquely for a given parent node execution.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.parent_node_execution_id:

parent_node_execution_id
  (:ref:`flyteidl.core.NodeExecutionIdentifier <api_msg_flyteidl.core.NodeExecutionIdentifier>`) A task execution is always kicked off by a node execution, the event consumer
  will use the parent_id to relate the task to it's parent node execution
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.retry_attempt:

retry_attempt
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) retry attempt number for this task, ie., 2 for the second attempt
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.phase:

phase
  (:ref:`flyteidl.core.TaskExecution.Phase <api_enum_flyteidl.core.TaskExecution.Phase>`) Phase associated with the event
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.producer_id:

producer_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) id of the process that sent this event, mainly for trace debugging
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.logs:

logs
  (:ref:`flyteidl.core.TaskLog <api_msg_flyteidl.core.TaskLog>`) log information for the task execution
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.occurred_at:

occurred_at
  (:ref:`google.protobuf.Timestamp <api_msg_google.protobuf.Timestamp>`) This timestamp represents when the original event occurred, it is generated
  by the executor of the task.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.input_uri:

input_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URI of the input file, it encodes all the information
  including Cloud source provider. ie., s3://...
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.output_uri:

output_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URI to the output of the execution, it will be in a format that encodes all the information
  including Cloud source provider. ie., s3://...
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.TaskExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.TaskExecutionEvent.error>` may be set.
  
.. _api_field_flyteidl.event.TaskExecutionEvent.error:

error
  (:ref:`flyteidl.core.ExecutionError <api_msg_flyteidl.core.ExecutionError>`) Error information for the execution
  
  
  
  Only one of :ref:`output_uri <api_field_flyteidl.event.TaskExecutionEvent.output_uri>`, :ref:`error <api_field_flyteidl.event.TaskExecutionEvent.error>` may be set.
  
.. _api_field_flyteidl.event.TaskExecutionEvent.custom_info:

custom_info
  (:ref:`google.protobuf.Struct <api_msg_google.protobuf.Struct>`) Custom data that the task plugin sends back. This is extensible to allow various plugins in the system.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.phase_version:

phase_version
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Some phases, like RUNNING, can send multiple events with changed metadata (new logs, additional custom_info, etc)
  that should be recorded regardless of the lack of phase change.
  The version field should be incremented when metadata changes across the duration of an individual phase.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.reason:

reason
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) An optional explanation for the phase transition.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.task_type:

task_type
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) A predefined yet extensible Task type identifier. If the task definition is already registered in flyte admin
  this type will be identical, but not all task executions necessarily use pre-registered definitions and this
  type is useful to render the task in the UI, filter task executions, etc.
  
  
.. _api_field_flyteidl.event.TaskExecutionEvent.metadata:

metadata
  (:ref:`flyteidl.event.TaskExecutionMetadata <api_msg_flyteidl.event.TaskExecutionMetadata>`) Metadata around how a task was executed.
  
  


.. _api_msg_flyteidl.event.ExternalResourceInfo:

flyteidl.event.ExternalResourceInfo
-----------------------------------

`[flyteidl.event.ExternalResourceInfo proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L167>`_

This message contains metadata about external resources produced or used by a specific task execution.

.. code-block:: json

  {
    "external_id": "..."
  }

.. _api_field_flyteidl.event.ExternalResourceInfo.external_id:

external_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Identifier for an external resource created by this task execution, for example Qubole query ID or presto query ids.
  
  


.. _api_msg_flyteidl.event.ResourcePoolInfo:

flyteidl.event.ResourcePoolInfo
-------------------------------

`[flyteidl.event.ResourcePoolInfo proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L176>`_

This message holds task execution metadata specific to resource allocation used to manage concurrent
executions for a project namespace.

.. code-block:: json

  {
    "allocation_token": "...",
    "namespace": "..."
  }

.. _api_field_flyteidl.event.ResourcePoolInfo.allocation_token:

allocation_token
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique resource ID used to identify this execution when allocating a token.
  
  
.. _api_field_flyteidl.event.ResourcePoolInfo.namespace:

namespace
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Namespace under which this task execution requested an allocation token.
  
  


.. _api_msg_flyteidl.event.TaskExecutionMetadata:

flyteidl.event.TaskExecutionMetadata
------------------------------------

`[flyteidl.event.TaskExecutionMetadata proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L188>`_

Holds metadata around how a task was executed.
As a task transitions across event phases during execution some attributes, such its generated name, generated external resources,
and more may grow in size but not change necessarily based on the phase transition that sparked the event update.
Metadata is a container for these attributes across the task execution lifecycle.

.. code-block:: json

  {
    "generated_name": "...",
    "external_resources": [],
    "resource_pool_info": [],
    "plugin_identifier": "...",
    "instance_class": "..."
  }

.. _api_field_flyteidl.event.TaskExecutionMetadata.generated_name:

generated_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Unique, generated name for this task execution used by the backend.
  
  
.. _api_field_flyteidl.event.TaskExecutionMetadata.external_resources:

external_resources
  (:ref:`flyteidl.event.ExternalResourceInfo <api_msg_flyteidl.event.ExternalResourceInfo>`) Additional data on external resources on other back-ends or platforms (e.g. Hive, Qubole, etc) launched by this task execution.
  
  
.. _api_field_flyteidl.event.TaskExecutionMetadata.resource_pool_info:

resource_pool_info
  (:ref:`flyteidl.event.ResourcePoolInfo <api_msg_flyteidl.event.ResourcePoolInfo>`) Includes additional data on concurrent resource management used during execution..
  This is a repeated field because a plugin can request multiple resource allocations during execution.
  
  
.. _api_field_flyteidl.event.TaskExecutionMetadata.plugin_identifier:

plugin_identifier
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The identifier of the plugin used to execute this task.
  
  
.. _api_field_flyteidl.event.TaskExecutionMetadata.instance_class:

instance_class
  (:ref:`flyteidl.event.TaskExecutionMetadata.InstanceClass <api_enum_flyteidl.event.TaskExecutionMetadata.InstanceClass>`) 
  

.. _api_enum_flyteidl.event.TaskExecutionMetadata.InstanceClass:

Enum flyteidl.event.TaskExecutionMetadata.InstanceClass
-------------------------------------------------------

`[flyteidl.event.TaskExecutionMetadata.InstanceClass proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/event/event.proto#L204>`_

Includes the broad cateogry of machine used for this specific task execution. 

.. _api_enum_value_flyteidl.event.TaskExecutionMetadata.InstanceClass.DEFAULT:

DEFAULT
  *(DEFAULT)* ⁣The default instance class configured for the flyte application platform.
  
  
.. _api_enum_value_flyteidl.event.TaskExecutionMetadata.InstanceClass.INTERRUPTIBLE:

INTERRUPTIBLE
  ⁣The instance class configured for interruptible tasks.
  
  
