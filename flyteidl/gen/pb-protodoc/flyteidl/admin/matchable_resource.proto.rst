.. _api_file_flyteidl/admin/matchable_resource.proto:

matchable_resource.proto
=======================================

.. _api_msg_flyteidl.admin.TaskResourceSpec:

flyteidl.admin.TaskResourceSpec
-------------------------------

`[flyteidl.admin.TaskResourceSpec proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L21>`_


.. code-block:: json

  {
    "cpu": "...",
    "gpu": "...",
    "memory": "...",
    "storage": "..."
  }

.. _api_field_flyteidl.admin.TaskResourceSpec.cpu:

cpu
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.TaskResourceSpec.gpu:

gpu
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.TaskResourceSpec.memory:

memory
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  
.. _api_field_flyteidl.admin.TaskResourceSpec.storage:

storage
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  


.. _api_msg_flyteidl.admin.TaskResourceAttributes:

flyteidl.admin.TaskResourceAttributes
-------------------------------------

`[flyteidl.admin.TaskResourceAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L31>`_


.. code-block:: json

  {
    "defaults": "{...}",
    "limits": "{...}"
  }

.. _api_field_flyteidl.admin.TaskResourceAttributes.defaults:

defaults
  (:ref:`flyteidl.admin.TaskResourceSpec <api_msg_flyteidl.admin.TaskResourceSpec>`) 
  
.. _api_field_flyteidl.admin.TaskResourceAttributes.limits:

limits
  (:ref:`flyteidl.admin.TaskResourceSpec <api_msg_flyteidl.admin.TaskResourceSpec>`) 
  


.. _api_msg_flyteidl.admin.ClusterResourceAttributes:

flyteidl.admin.ClusterResourceAttributes
----------------------------------------

`[flyteidl.admin.ClusterResourceAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L37>`_


.. code-block:: json

  {
    "attributes": "{...}"
  }

.. _api_field_flyteidl.admin.ClusterResourceAttributes.attributes:

attributes
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, `string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_>) Custom resource attributes which will be applied in cluster resource creation (e.g. quotas).
  Map keys are the *case-sensitive* names of variables in templatized resource files.
  Map values should be the custom values which get substituted during resource creation.
  
  


.. _api_msg_flyteidl.admin.ExecutionQueueAttributes:

flyteidl.admin.ExecutionQueueAttributes
---------------------------------------

`[flyteidl.admin.ExecutionQueueAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L44>`_


.. code-block:: json

  {
    "tags": []
  }

.. _api_field_flyteidl.admin.ExecutionQueueAttributes.tags:

tags
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Tags used for assigning execution queues for tasks defined within this project.
  
  


.. _api_msg_flyteidl.admin.ExecutionClusterLabel:

flyteidl.admin.ExecutionClusterLabel
------------------------------------

`[flyteidl.admin.ExecutionClusterLabel proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L49>`_


.. code-block:: json

  {
    "value": "..."
  }

.. _api_field_flyteidl.admin.ExecutionClusterLabel.value:

value
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Label value to determine where the execution will be run
  
  


.. _api_msg_flyteidl.admin.MatchingAttributes:

flyteidl.admin.MatchingAttributes
---------------------------------

`[flyteidl.admin.MatchingAttributes proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L55>`_

Generic container for encapsulating all types of the above attributes messages.

.. code-block:: json

  {
    "task_resource_attributes": "{...}",
    "cluster_resource_attributes": "{...}",
    "execution_queue_attributes": "{...}",
    "execution_cluster_label": "{...}"
  }

.. _api_field_flyteidl.admin.MatchingAttributes.task_resource_attributes:

task_resource_attributes
  (:ref:`flyteidl.admin.TaskResourceAttributes <api_msg_flyteidl.admin.TaskResourceAttributes>`) 
  
  
  Only one of :ref:`task_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.task_resource_attributes>`, :ref:`cluster_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.cluster_resource_attributes>`, :ref:`execution_queue_attributes <api_field_flyteidl.admin.MatchingAttributes.execution_queue_attributes>`, :ref:`execution_cluster_label <api_field_flyteidl.admin.MatchingAttributes.execution_cluster_label>` may be set.
  
.. _api_field_flyteidl.admin.MatchingAttributes.cluster_resource_attributes:

cluster_resource_attributes
  (:ref:`flyteidl.admin.ClusterResourceAttributes <api_msg_flyteidl.admin.ClusterResourceAttributes>`) 
  
  
  Only one of :ref:`task_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.task_resource_attributes>`, :ref:`cluster_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.cluster_resource_attributes>`, :ref:`execution_queue_attributes <api_field_flyteidl.admin.MatchingAttributes.execution_queue_attributes>`, :ref:`execution_cluster_label <api_field_flyteidl.admin.MatchingAttributes.execution_cluster_label>` may be set.
  
.. _api_field_flyteidl.admin.MatchingAttributes.execution_queue_attributes:

execution_queue_attributes
  (:ref:`flyteidl.admin.ExecutionQueueAttributes <api_msg_flyteidl.admin.ExecutionQueueAttributes>`) 
  
  
  Only one of :ref:`task_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.task_resource_attributes>`, :ref:`cluster_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.cluster_resource_attributes>`, :ref:`execution_queue_attributes <api_field_flyteidl.admin.MatchingAttributes.execution_queue_attributes>`, :ref:`execution_cluster_label <api_field_flyteidl.admin.MatchingAttributes.execution_cluster_label>` may be set.
  
.. _api_field_flyteidl.admin.MatchingAttributes.execution_cluster_label:

execution_cluster_label
  (:ref:`flyteidl.admin.ExecutionClusterLabel <api_msg_flyteidl.admin.ExecutionClusterLabel>`) 
  
  
  Only one of :ref:`task_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.task_resource_attributes>`, :ref:`cluster_resource_attributes <api_field_flyteidl.admin.MatchingAttributes.cluster_resource_attributes>`, :ref:`execution_queue_attributes <api_field_flyteidl.admin.MatchingAttributes.execution_queue_attributes>`, :ref:`execution_cluster_label <api_field_flyteidl.admin.MatchingAttributes.execution_cluster_label>` may be set.
  

.. _api_enum_flyteidl.admin.MatchableResource:

Enum flyteidl.admin.MatchableResource
-------------------------------------

`[flyteidl.admin.MatchableResource proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/matchable_resource.proto#L7>`_

Defines a resource that can be configured by customizable Project-, ProjectDomain- or WorkflowAttributes
based on matching tags.

.. _api_enum_value_flyteidl.admin.MatchableResource.TASK_RESOURCE:

TASK_RESOURCE
  *(DEFAULT)* ⁣Applies to customizable task resource requests and limits.
  
  
.. _api_enum_value_flyteidl.admin.MatchableResource.CLUSTER_RESOURCE:

CLUSTER_RESOURCE
  ⁣Applies to configuring templated kubernetes cluster resources.
  
  
.. _api_enum_value_flyteidl.admin.MatchableResource.EXECUTION_QUEUE:

EXECUTION_QUEUE
  ⁣Configures task and dynamic task execution queue assignment.
  
  
.. _api_enum_value_flyteidl.admin.MatchableResource.EXECUTION_CLUSTER_LABEL:

EXECUTION_CLUSTER_LABEL
  ⁣Configures the K8s cluster label to be used for execution to be run
  
  
