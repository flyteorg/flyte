.. _api_file_flyteidl/core/compiler.proto:

compiler.proto
============================

.. _api_msg_flyteidl.core.ConnectionSet:

flyteidl.core.ConnectionSet
---------------------------

`[flyteidl.core.ConnectionSet proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/compiler.proto#L11>`_

Adjacency list for the workflow. This is created as part of the compilation process. Every process after the compilation
step uses this created ConnectionSet

.. code-block:: json

  {
    "downstream": "{...}",
    "upstream": "{...}"
  }

.. _api_field_flyteidl.core.ConnectionSet.downstream:

downstream
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`flyteidl.core.ConnectionSet.IdList <api_msg_flyteidl.core.ConnectionSet.IdList>`>) A list of all the node ids that are downstream from a given node id
  
  
.. _api_field_flyteidl.core.ConnectionSet.upstream:

upstream
  (map<`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_, :ref:`flyteidl.core.ConnectionSet.IdList <api_msg_flyteidl.core.ConnectionSet.IdList>`>) A list of all the node ids, that are upstream of this node id
  
  
.. _api_msg_flyteidl.core.ConnectionSet.IdList:

flyteidl.core.ConnectionSet.IdList
----------------------------------

`[flyteidl.core.ConnectionSet.IdList proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/compiler.proto#L12>`_


.. code-block:: json

  {
    "ids": []
  }

.. _api_field_flyteidl.core.ConnectionSet.IdList.ids:

ids
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) 
  





.. _api_msg_flyteidl.core.CompiledWorkflow:

flyteidl.core.CompiledWorkflow
------------------------------

`[flyteidl.core.CompiledWorkflow proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/compiler.proto#L24>`_

Output of the compilation Step. This object represents one workflow. We store more metadata at this layer

.. code-block:: json

  {
    "template": "{...}",
    "connections": "{...}"
  }

.. _api_field_flyteidl.core.CompiledWorkflow.template:

template
  (:ref:`flyteidl.core.WorkflowTemplate <api_msg_flyteidl.core.WorkflowTemplate>`) Completely contained Workflow Template
  
  
.. _api_field_flyteidl.core.CompiledWorkflow.connections:

connections
  (:ref:`flyteidl.core.ConnectionSet <api_msg_flyteidl.core.ConnectionSet>`) For internal use only! This field is used by the system and must not be filled in. Any values set will be ignored.
  
  


.. _api_msg_flyteidl.core.CompiledTask:

flyteidl.core.CompiledTask
--------------------------

`[flyteidl.core.CompiledTask proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/compiler.proto#L32>`_

Output of the Compilation step. This object represent one Task. We store more metadata at this layer

.. code-block:: json

  {
    "template": "{...}"
  }

.. _api_field_flyteidl.core.CompiledTask.template:

template
  (:ref:`flyteidl.core.TaskTemplate <api_msg_flyteidl.core.TaskTemplate>`) Completely contained TaskTemplate
  
  


.. _api_msg_flyteidl.core.CompiledWorkflowClosure:

flyteidl.core.CompiledWorkflowClosure
-------------------------------------

`[flyteidl.core.CompiledWorkflowClosure proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/compiler.proto#L41>`_

A Compiled Workflow Closure contains all the information required to start a new execution, or to visualize a workflow
and its details. The CompiledWorkflowClosure should always contain a primary workflow, that is the main workflow that
will being the execution. All subworkflows are denormalized. WorkflowNodes refer to the workflow identifiers of
compiled subworkflows.

.. code-block:: json

  {
    "primary": "{...}",
    "sub_workflows": [],
    "tasks": []
  }

.. _api_field_flyteidl.core.CompiledWorkflowClosure.primary:

primary
  (:ref:`flyteidl.core.CompiledWorkflow <api_msg_flyteidl.core.CompiledWorkflow>`) required
  
  
.. _api_field_flyteidl.core.CompiledWorkflowClosure.sub_workflows:

sub_workflows
  (:ref:`flyteidl.core.CompiledWorkflow <api_msg_flyteidl.core.CompiledWorkflow>`) Guaranteed that there will only exist one and only one workflow with a given id, i.e., every sub workflow has a
  unique identifier. Also every enclosed subworkflow is used either by a primary workflow or by a subworkflow
  as an inlined workflow
  optional
  
  
.. _api_field_flyteidl.core.CompiledWorkflowClosure.tasks:

tasks
  (:ref:`flyteidl.core.CompiledTask <api_msg_flyteidl.core.CompiledTask>`) Guaranteed that there will only exist one and only one task with a given id, i.e., every task has a unique id
  required (atleast 1)
  
  

