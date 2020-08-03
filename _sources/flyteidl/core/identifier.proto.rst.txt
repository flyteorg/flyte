.. _api_file_flyteidl/core/identifier.proto:

identifier.proto
==============================

.. _api_msg_flyteidl.core.Identifier:

flyteidl.core.Identifier
------------------------

`[flyteidl.core.Identifier proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/identifier.proto#L19>`_

Encapsulation of fields that uniquely identifies a Flyte resource.

.. code-block:: json

  {
    "resource_type": "...",
    "project": "...",
    "domain": "...",
    "name": "...",
    "version": "..."
  }

.. _api_field_flyteidl.core.Identifier.resource_type:

resource_type
  (:ref:`flyteidl.core.ResourceType <api_enum_flyteidl.core.ResourceType>`) Identifies the specific type of resource that this identifer corresponds to.
  
  
.. _api_field_flyteidl.core.Identifier.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project the resource belongs to.
  
  
.. _api_field_flyteidl.core.Identifier.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the resource belongs to.
  A domain can be considered as a subset within a specific project.
  
  
.. _api_field_flyteidl.core.Identifier.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User provided value for the resource.
  
  
.. _api_field_flyteidl.core.Identifier.version:

version
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Specific version of the resource.
  
  


.. _api_msg_flyteidl.core.WorkflowExecutionIdentifier:

flyteidl.core.WorkflowExecutionIdentifier
-----------------------------------------

`[flyteidl.core.WorkflowExecutionIdentifier proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/identifier.proto#L38>`_

Encapsulation of fields that uniquely identifies a Flyte workflow execution

.. code-block:: json

  {
    "project": "...",
    "domain": "...",
    "name": "..."
  }

.. _api_field_flyteidl.core.WorkflowExecutionIdentifier.project:

project
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the project the resource belongs to.
  
  
.. _api_field_flyteidl.core.WorkflowExecutionIdentifier.domain:

domain
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Name of the domain the resource belongs to.
  A domain can be considered as a subset within a specific project.
  
  
.. _api_field_flyteidl.core.WorkflowExecutionIdentifier.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) User or system provided value for the resource.
  
  


.. _api_msg_flyteidl.core.NodeExecutionIdentifier:

flyteidl.core.NodeExecutionIdentifier
-------------------------------------

`[flyteidl.core.NodeExecutionIdentifier proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/identifier.proto#L51>`_

Encapsulation of fields that identify a Flyte node execution entity.

.. code-block:: json

  {
    "node_id": "...",
    "execution_id": "{...}"
  }

.. _api_field_flyteidl.core.NodeExecutionIdentifier.node_id:

node_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.core.NodeExecutionIdentifier.execution_id:

execution_id
  (:ref:`flyteidl.core.WorkflowExecutionIdentifier <api_msg_flyteidl.core.WorkflowExecutionIdentifier>`) 
  


.. _api_msg_flyteidl.core.TaskExecutionIdentifier:

flyteidl.core.TaskExecutionIdentifier
-------------------------------------

`[flyteidl.core.TaskExecutionIdentifier proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/identifier.proto#L58>`_

Encapsulation of fields that identify a Flyte task execution entity.

.. code-block:: json

  {
    "task_id": "{...}",
    "node_execution_id": "{...}",
    "retry_attempt": "..."
  }

.. _api_field_flyteidl.core.TaskExecutionIdentifier.task_id:

task_id
  (:ref:`flyteidl.core.Identifier <api_msg_flyteidl.core.Identifier>`) 
  
.. _api_field_flyteidl.core.TaskExecutionIdentifier.node_execution_id:

node_execution_id
  (:ref:`flyteidl.core.NodeExecutionIdentifier <api_msg_flyteidl.core.NodeExecutionIdentifier>`) 
  
.. _api_field_flyteidl.core.TaskExecutionIdentifier.retry_attempt:

retry_attempt
  (`uint32 <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  

.. _api_enum_flyteidl.core.ResourceType:

Enum flyteidl.core.ResourceType
-------------------------------

`[flyteidl.core.ResourceType proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/identifier.proto#L7>`_

Indicates a resource type within Flyte.

.. _api_enum_value_flyteidl.core.ResourceType.UNSPECIFIED:

UNSPECIFIED
  *(DEFAULT)* ⁣
  
.. _api_enum_value_flyteidl.core.ResourceType.TASK:

TASK
  ⁣
  
.. _api_enum_value_flyteidl.core.ResourceType.WORKFLOW:

WORKFLOW
  ⁣
  
.. _api_enum_value_flyteidl.core.ResourceType.LAUNCH_PLAN:

LAUNCH_PLAN
  ⁣
  
.. _api_enum_value_flyteidl.core.ResourceType.DATASET:

DATASET
  ⁣A dataset represents an entity modeled in Flyte DataCatalog. A Dataset is also a versioned entity and can be a compilation of multiple individual objects.
  Eventually all Catalog objects should be modeled similar to Flyte Objects. The Dataset entities makes it possible for the UI  and CLI to act on the objects 
  in a similar manner to other Flyte objects
  
  
